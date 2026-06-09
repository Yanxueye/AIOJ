package mq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/terminaloj/backend/internal/judger"
	"github.com/terminaloj/backend/internal/models"
	"gorm.io/gorm"
)

// Worker drains the submit queue and writes results back to MySQL.
type Worker struct {
	Broker      *Broker
	DB          *gorm.DB
	Judger      judger.JudgerClient
	Concurrency int
}

// Start blocks until ctx is cancelled.
func (w *Worker) Start(ctx context.Context) error {
	if w.Concurrency <= 0 {
		w.Concurrency = 4
	}
	stream, err := w.Broker.Consume(ctx)
	if err != nil {
		return err
	}
	log.Printf("[worker] started with concurrency=%d", w.Concurrency)

	tokens := make(chan struct{}, w.Concurrency)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case body, ok := <-stream:
			if !ok {
				return errors.New("mq stream closed")
			}
			tokens <- struct{}{}
			go func(raw []byte) {
				defer func() { <-tokens }()
				w.process(ctx, raw)
			}(body)
		}
	}
}

func (w *Worker) process(ctx context.Context, raw []byte) {
	var task SubmitTask
	if err := json.Unmarshal(raw, &task); err != nil {
		log.Printf("[worker] bad payload: %v", err)
		return
	}

	if err := w.processNewSubmission(ctx, task); err != nil {
		log.Printf("[worker] process submission %d failed: %v", task.SubmissionID, err)
	}
}

func (w *Worker) processNewSubmission(ctx context.Context, task SubmitTask) error {
	now := time.Now().UTC()
	queueStarted := now
	sub := &models.Submission{
		ID:             task.SubmissionID,
		UserID:         task.UserID,
		ProblemID:      task.ProblemID,
		ProblemTitle:   task.ProblemTitle,
		TraceID:        task.TraceID,
		Source:         task.Source,
		Language:       task.Language,
		Code:           task.Code,
		CodeLength:     len(task.Code),
		Status:         models.StatusQueueing,
		Runtime:        0,
		RuntimeMS:      0,
		Memory:         "0.0",
		MemoryKB:       0,
		QueueStartedAt: &queueStarted,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := w.DB.Create(sub).Error; err != nil {
		return fmt.Errorf("insert submission %d: %w", task.SubmissionID, err)
	}
	return w.judgeSubmission(ctx, sub)
}

func (w *Worker) RejudgeSubmission(ctx context.Context, submissionID uint64) error {
	var sub models.Submission
	if err := w.DB.Where("id = ? AND source = ?", submissionID, "submit").First(&sub).Error; err != nil {
		return err
	}

	now := time.Now().UTC()
	sub.Status = models.StatusQueueing
	sub.Runtime = 0
	sub.RuntimeMS = 0
	sub.Memory = "0.0"
	sub.MemoryKB = 0
	sub.CompileOutput = ""
	sub.ErrorMessage = ""
	sub.QueueStartedAt = &now
	sub.JudgeStartedAt = nil
	sub.FinishedAt = nil
	sub.UpdatedAt = now
	if err := w.DB.Save(&sub).Error; err != nil {
		return err
	}
	return w.judgeSubmission(ctx, &sub)
}

func (w *Worker) ProcessRejudgeJob(ctx context.Context, jobID uint64) error {
	var job models.RejudgeJob
	if err := w.DB.First(&job, jobID).Error; err != nil {
		return err
	}
	if job.Status != "pending" {
		return nil
	}

	now := time.Now().UTC()
	job.Status = "running"
	job.StartedAt = &now
	job.UpdatedAt = now
	if err := w.DB.Save(&job).Error; err != nil {
		return err
	}

	var subs []models.Submission
	if err := w.DB.Where("problem_id = ? AND source = ?", job.ProblemID, "submit").Order("id ASC").Find(&subs).Error; err != nil {
		job.Status = "failed"
		job.UpdatedAt = time.Now().UTC()
		_ = w.DB.Save(&job).Error
		return err
	}

	job.TotalSubmissions = len(subs)
	_ = w.DB.Save(&job).Error

	for _, item := range subs {
		err := w.RejudgeSubmission(ctx, item.ID)
		job.ProcessedCount++
		if err != nil {
			job.FailedCount++
			log.Printf("[rejudge] submission %d failed: %v", item.ID, err)
		} else {
			job.SucceededCount++
		}
		job.UpdatedAt = time.Now().UTC()
		_ = w.DB.Save(&job).Error
	}

	finished := time.Now().UTC()
	job.Status = "finished"
	job.FinishedAt = &finished
	job.UpdatedAt = finished
	return w.DB.Save(&job).Error
}

func (w *Worker) judgeSubmission(ctx context.Context, sub *models.Submission) error {
	var problem models.Problem
	if err := w.DB.Preload("PublishedVersion.TestCases").Preload("PublishedVersion").First(&problem, sub.ProblemID).Error; err != nil {
		w.fail(sub, models.StatusSystemErr, "problem not found")
		return err
	}
	if problem.PublishedVersion == nil {
		w.fail(sub, models.StatusSystemErr, "problem has no published version")
		return errors.New("problem has no published version")
	}

	judgeStarted := time.Now().UTC()
	sub.Status = models.StatusCompiling
	sub.JudgeStartedAt = &judgeStarted
	if err := w.DB.Save(sub).Error; err != nil {
		return err
	}

	req := &judger.JudgeRequest{
		SubmissionID:  sub.ID,
		ProblemID:     sub.ProblemID,
		TraceID:       sub.TraceID,
		Language:      sub.Language,
		Code:          sub.Code,
		TimeLimitMS:   int32(problem.PublishedVersion.TimeLimit),
		MemoryLimitMB: int32(problem.PublishedVersion.MemoryLimit),
		OutputLimitKB: problem.PublishedVersion.OutputLimitKB,
	}
	for _, tc := range problem.PublishedVersion.TestCases {
		req.TestCases = append(req.TestCases, judger.TestCase{
			CaseNo:   int32(tc.CaseNo),
			Input:    tc.Input,
			Expected: tc.Expected,
		})
	}

	resp, err := w.Judger.Judge(ctx, req)
	if err != nil {
		w.fail(sub, models.StatusSystemErr, "judger rpc failed: "+err.Error())
		return err
	}

	finishedAt := time.Now().UTC()
	sub.Status = resp.Status
	sub.Runtime = int(resp.RuntimeMS)
	sub.RuntimeMS = int(resp.RuntimeMS)
	sub.Memory = resp.MemoryMB
	sub.MemoryKB = int(resp.MemoryKB)
	sub.CompileOutput = resp.CompileOut
	sub.ErrorMessage = resp.ErrorMessage
	sub.FinishedAt = &finishedAt
	sub.UpdatedAt = finishedAt

	caseResults := make([]models.SubmissionCaseResult, len(resp.CaseResults))
	for i, item := range resp.CaseResults {
		caseResults[i] = models.SubmissionCaseResult{
			SubmissionID:  sub.ID,
			CaseNo:        int(item.CaseNo),
			Status:        item.Status,
			RuntimeMS:     int(item.RuntimeMS),
			MemoryKB:      int(item.MemoryKB),
			StdoutBytes:   int(item.StdoutBytes),
			StderrBytes:   int(item.StderrBytes),
			Signal:        item.Signal,
			StdoutPreview: item.StdoutPreview,
			StderrPreview: item.StderrPreview,
		}
	}
	if err := w.DB.Save(sub).Error; err != nil {
		return err
	}

	w.DB.Where("submission_id = ?", sub.ID).Delete(&models.SubmissionCaseResult{})
	if len(caseResults) > 0 {
		if err := w.DB.Create(&caseResults).Error; err != nil {
			log.Printf("[worker] insert case results %d: %v", sub.ID, err)
		}
	}

	if sub.QueueStartedAt != nil && sub.CreatedAt.Equal(*sub.QueueStartedAt) {
		w.DB.Model(&models.Problem{}).Where("id = ?", sub.ProblemID).
			UpdateColumn("submit_count", gorm.Expr("submit_count + 1"))
		if sub.Status == models.StatusAccepted {
			w.DB.Model(&models.Problem{}).Where("id = ?", sub.ProblemID).
				UpdateColumn("accept_count", gorm.Expr("accept_count + 1"))
			w.updateStudyPlanProgress(sub.UserID, sub.ProblemID, sub.ID)
		}
	}
	return nil
}

func (w *Worker) fail(sub *models.Submission, status, msg string) {
	now := time.Now().UTC()
	sub.Status = status
	sub.ErrorMessage = msg
	sub.FinishedAt = &now
	sub.UpdatedAt = now
	_ = w.DB.Save(sub).Error
}

func (w *Worker) updateStudyPlanProgress(userID, problemID, submissionID uint64) {
	var priorCount int64
	w.DB.Model(&models.Submission{}).
		Where("user_id = ? AND problem_id = ? AND status = ? AND id <> ?", userID, problemID, models.StatusAccepted, submissionID).
		Count(&priorCount)
	if priorCount > 0 {
		return
	}

	var planItems []models.StudyPlanItem
	if err := w.DB.Where("problem_id = ?", problemID).Find(&planItems).Error; err != nil {
		return
	}
	now := time.Now().UTC()
	date := now.Format("2006-01-02")
	for _, item := range planItems {
		var progressItem models.UserPlanProgressItem
		pItemErr := w.DB.Where("user_id = ? AND plan_id = ? AND problem_id = ?", userID, item.PlanID, problemID).First(&progressItem).Error
		if pItemErr == gorm.ErrRecordNotFound {
			progressItem = models.UserPlanProgressItem{
				UserID:      userID,
				PlanID:      item.PlanID,
				ProblemID:   problemID,
				Completed:   true,
				CompletedAt: &now,
			}
			_ = w.DB.Create(&progressItem).Error
		} else if pItemErr == nil {
			progressItem.Completed = true
			progressItem.CompletedAt = &now
			_ = w.DB.Save(&progressItem).Error
		}

		var progress models.UserPlanProgress
		err := w.DB.Where("user_id = ? AND plan_id = ?", userID, item.PlanID).First(&progress).Error
		if err == nil {
			progress.CompletedCount++
			progress.LastCompletedAt = &now
			progress.UpdatedAt = now
			_ = w.DB.Save(&progress).Error
			continue
		}
		if err != gorm.ErrRecordNotFound {
			continue
		}
		progress = models.UserPlanProgress{
			UserID:          userID,
			PlanID:          item.PlanID,
			CompletedCount:  1,
			LastCompletedAt: &now,
		}
		_ = w.DB.Create(&progress).Error
	}

	var checkin models.StudyCheckin
	err := w.DB.Where("user_id = ? AND date = ?", userID, date).First(&checkin).Error
	if err == gorm.ErrRecordNotFound {
		checkin = models.StudyCheckin{
			UserID: userID,
			Date:   date,
			Count:  1,
		}
		_ = w.DB.Create(&checkin).Error
		return
	}
	if err == nil {
		checkin.Count++
		_ = w.DB.Save(&checkin).Error
	}
}
