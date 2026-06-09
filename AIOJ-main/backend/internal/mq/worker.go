package mq

import (
	"context"
	"encoding/json"
	"errors"
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

	now := time.Now().UTC()
	queueStarted := now
	sub := &models.Submission{
		ID:             task.SubmissionID,
		UserID:         task.UserID,
		ProblemID:      task.ProblemID,
		ProblemTitle:   task.ProblemTitle,
		TraceID:        task.TraceID,
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
		log.Printf("[worker] insert submission %d: %v", task.SubmissionID, err)
		return
	}

	var problem models.Problem
	if err := w.DB.First(&problem, task.ProblemID).Error; err != nil {
		w.fail(sub, models.StatusSystemErr, "problem not found")
		return
	}

	judgeStarted := time.Now().UTC()
	sub.Status = models.StatusCompiling
	sub.JudgeStartedAt = &judgeStarted
	if err := w.DB.Save(sub).Error; err != nil {
		log.Printf("[worker] mark compiling %d: %v", sub.ID, err)
		return
	}

	req := &judger.JudgeRequest{
		SubmissionID:  task.SubmissionID,
		ProblemID:     task.ProblemID,
		TraceID:       task.TraceID,
		Language:      task.Language,
		Code:          task.Code,
		TimeLimitMS:   int32(problem.TimeLimit),
		MemoryLimitMB: int32(problem.MemoryLimit),
		OutputLimitKB: problem.OutputLimitKBOrDefault(),
	}
	for _, tc := range problem.TestCases {
		req.TestCases = append(req.TestCases, judger.TestCase{
			Input:    tc.Input,
			Expected: tc.Expected,
		})
	}

	resp, err := w.Judger.Judge(ctx, req)
	if err != nil {
		log.Printf("[worker] judge rpc failure: %v", err)
		w.fail(sub, models.StatusSystemErr, "judger rpc failed: "+err.Error())
		return
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
		log.Printf("[worker] update submission %d: %v", sub.ID, err)
		return
	}

	w.DB.Where("submission_id = ?", sub.ID).Delete(&models.SubmissionCaseResult{})
	if len(caseResults) > 0 {
		if err := w.DB.Create(&caseResults).Error; err != nil {
			log.Printf("[worker] insert case results %d: %v", sub.ID, err)
		}
	}

	w.DB.Model(&models.Problem{}).Where("id = ?", sub.ProblemID).
		UpdateColumn("submit_count", gorm.Expr("submit_count + 1"))
	if sub.Status == models.StatusAccepted {
		w.DB.Model(&models.Problem{}).Where("id = ?", sub.ProblemID).
			UpdateColumn("accept_count", gorm.Expr("accept_count + 1"))
	}
}

func (w *Worker) fail(sub *models.Submission, status, msg string) {
	now := time.Now().UTC()
	sub.Status = status
	sub.ErrorMessage = msg
	sub.FinishedAt = &now
	sub.UpdatedAt = now
	_ = w.DB.Save(sub).Error
}
