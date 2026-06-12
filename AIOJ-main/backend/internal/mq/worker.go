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
	"github.com/terminaloj/backend/internal/utils"
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

	log.Printf("[worker] received task: sub=%d problem=%d lang=%s", task.SubmissionID, task.ProblemID, task.Language)
	if err := w.processNewSubmission(ctx, task); err != nil {
		log.Printf("[worker] process submission %d failed: %v", task.SubmissionID, err)
	}
}

func (w *Worker) processNewSubmission(ctx context.Context, task SubmitTask) error {
	// Load existing submission (created by the handler with auto-increment ID)
	var sub models.Submission
	if err := w.DB.First(&sub, task.SubmissionID).Error; err != nil {
		return fmt.Errorf("load submission %d: %w", task.SubmissionID, err)
	}
	sub.TraceID = task.TraceID
	sub.Status = models.StatusQueueing
	sub.UpdatedAt = time.Now().UTC()
	if err := w.DB.Save(&sub).Error; err != nil {
		return fmt.Errorf("update submission %d: %w", task.SubmissionID, err)
	}
	return w.judgeSubmission(ctx, &sub)
}

func (w *Worker) judgeSubmission(ctx context.Context, sub *models.Submission) error {
	t0 := time.Now()
	var problem models.Problem
	if err := w.DB.Preload("PublishedVersion.TestCases").Preload("PublishedVersion").First(&problem, sub.ProblemID).Error; err != nil {
		w.fail(sub, models.StatusSystemErr, "problem not found")
		return err
	}
	if problem.PublishedVersion == nil {
		w.fail(sub, models.StatusSystemErr, "problem has no published version")
		return errors.New("problem has no published version")
	}
	log.Printf("[worker] sub %d: loaded problem %d with %d test cases (%v)", sub.ID, sub.ProblemID, len(problem.PublishedVersion.TestCases), time.Since(t0))

	judgeStarted := time.Now().UTC()
	sub.Status = models.StatusCompiling
	sub.JudgeStartedAt = &judgeStarted
	if err := w.DB.Save(sub).Error; err != nil {
		return err
	}
	log.Printf("[worker] sub %d: status -> Compiling, calling judger...", sub.ID)

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

	// Retry RPC call with exponential backoff (max 3 attempts)
	var resp *judger.JudgeResponse
	const maxRetries = 3
	judgeStart := time.Now()
	for attempt := 1; attempt <= maxRetries; attempt++ {
		var rpcErr error
		resp, rpcErr = w.Judger.Judge(ctx, req)
		if rpcErr == nil {
			log.Printf("[worker] sub %d: judge RPC success on attempt %d (%v)", sub.ID, attempt, time.Since(judgeStart))
			break
		}
		if attempt < maxRetries {
			backoff := time.Duration(1<<uint(attempt-1)) * 500 * time.Millisecond
			log.Printf("[worker] sub %d: judge RPC failed (attempt %d/%d): %v, retrying in %v", sub.ID, attempt, maxRetries, rpcErr, backoff)
			time.Sleep(backoff)
		} else {
			log.Printf("[worker] sub %d: judge RPC failed after %d attempts: %v (%v)", sub.ID, maxRetries, rpcErr, time.Since(judgeStart))
			w.fail(sub, models.StatusSystemErr, "judger rpc failed after retries: "+rpcErr.Error())
			return rpcErr
		}
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

	// Build lookup map for test case input/expected
	tcMap := make(map[int]string, len(problem.PublishedVersion.TestCases)*2)
	for _, tc := range problem.PublishedVersion.TestCases {
		tcMap[int(tc.CaseNo)*100+1] = tc.Input   // encode: caseNo*100+1 = input
		tcMap[int(tc.CaseNo)*100+2] = tc.Expected // encode: caseNo*100+2 = expected
	}

	caseResults := make([]models.SubmissionCaseResult, len(resp.CaseResults))
	for i, item := range resp.CaseResults {
		cn := int(item.CaseNo)
		caseResults[i] = models.SubmissionCaseResult{
			SubmissionID:  sub.ID,
			CaseNo:        cn,
			Status:        item.Status,
			RuntimeMS:     int(item.RuntimeMS),
			MemoryKB:      int(item.MemoryKB),
			StdoutBytes:   int(item.StdoutBytes),
			StderrBytes:   int(item.StderrBytes),
			Signal:        item.Signal,
			StdoutPreview: item.StdoutPreview,
			StderrPreview: item.StderrPreview,
			Input:         tcMap[cn*100+1],
			Expected:      tcMap[cn*100+2],
		}
	}
	if err := w.DB.Save(sub).Error; err != nil {
		return err
	}
	log.Printf("[worker] sub %d: final status -> %s (total %v)", sub.ID, sub.Status, time.Since(t0))

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
			w.updateUserRating(sub.UserID, sub.ProblemID)
			utils.UpdateMastery(w.DB, sub.UserID)
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

func (w *Worker) updateUserRating(userID, problemID uint64) {
	var user models.User
	if err := w.DB.First(&user, userID).Error; err != nil {
		log.Printf("[worker] failed to load user %d for rating update: %v", userID, err)
		return
	}

	var problem models.Problem
	if err := w.DB.First(&problem, problemID).Error; err != nil {
		log.Printf("[worker] failed to load problem %d for rating update: %v", problemID, err)
		return
	}

	problemRating := problem.Rating
	if problemRating <= 0 {
		problemRating = problem.DifficultyScore
	}
	if problemRating <= 0 {
		problemRating = utils.DefaultProblemRating
	}

	newRating := utils.CalculateUserRatingUpdate(user.Rating, problemRating)
	if newRating != user.Rating {
		if err := w.DB.Model(&user).UpdateColumn("rating", newRating).Error; err != nil {
			log.Printf("[worker] failed to update rating for user %d: %v", userID, err)
			return
		}
		// Record rating history
		history := models.RatingHistory{
			UserID:    userID,
			OldRating: user.Rating,
			NewRating: newRating,
			Delta:     newRating - user.Rating,
			ProblemID: problemID,
			Reason:    "accepted",
		}
		if err := w.DB.Create(&history).Error; err != nil {
			log.Printf("[worker] failed to record rating history for user %d: %v", userID, err)
		}
	}
}
