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

// Worker drains the submit queue and writes results back to MySQL. A small
// pool of goroutines (size=Concurrency) processes messages in parallel.
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

	sub := &models.Submission{
		ID:           task.SubmissionID,
		UserID:       task.UserID,
		ProblemID:    task.ProblemID,
		ProblemTitle: task.ProblemTitle,
		Language:     task.Language,
		Code:         task.Code,
		CodeLength:   len(task.Code),
		Status:       models.StatusPending,
		CreatedAt:    time.Now(),
	}
	if err := w.DB.Create(sub).Error; err != nil {
		log.Printf("[worker] insert submission %d: %v", task.SubmissionID, err)
		return
	}

	var problem models.Problem
	if err := w.DB.First(&problem, task.ProblemID).Error; err != nil {
		w.fail(sub, "Runtime Error", "problem not found")
		return
	}

	req := &judger.JudgeRequest{
		SubmissionID:  task.SubmissionID,
		ProblemID:     task.ProblemID,
		Language:      task.Language,
		Code:          task.Code,
		TimeLimitMS:   int32(problem.TimeLimit),
		MemoryLimitMB: int32(problem.MemoryLimit),
	}
	for _, tc := range problem.TestCases {
		req.TestCases = append(req.TestCases, judger.TestCase{Input: tc.Input, Expected: tc.Expected})
	}

	resp, err := w.Judger.Judge(ctx, req)
	if err != nil {
		log.Printf("[worker] judge rpc failure: %v", err)
		w.fail(sub, "Runtime Error", "judger rpc failed: "+err.Error())
		return
	}

	sub.Status = resp.Status
	sub.Runtime = int(resp.RuntimeMS)
	sub.Memory = resp.MemoryMB
	sub.ErrorMessage = resp.ErrorMessage
	if err := w.DB.Save(sub).Error; err != nil {
		log.Printf("[worker] update submission %d: %v", sub.ID, err)
		return
	}

	w.DB.Model(&models.Problem{}).Where("id = ?", sub.ProblemID).
		UpdateColumn("submit_count", gorm.Expr("submit_count + 1"))
	if sub.Status == models.StatusAccepted {
		w.DB.Model(&models.Problem{}).Where("id = ?", sub.ProblemID).
			UpdateColumn("accept_count", gorm.Expr("accept_count + 1"))
	}
}

func (w *Worker) fail(sub *models.Submission, status, msg string) {
	sub.Status = status
	sub.ErrorMessage = msg
	w.DB.Save(sub)
}
