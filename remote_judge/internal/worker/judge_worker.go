package worker

import (
	"context"
	"time"

	"remote_judge/internal/domain"
	"remote_judge/internal/judger"
	"remote_judge/internal/logger"
	"remote_judge/internal/queue"
	"remote_judge/internal/repository"
	"remote_judge/internal/stats"
)

// JudgeWorker 消费提交消息并分发判题任务。
type JudgeWorker struct {
	q             queue.Queue
	submissions   repository.SubmissionRepository
	problems      repository.ProblemRepository
	judger        judger.Executor
	stats         *stats.Collector
	concurrency   int
	acquireTimout time.Duration
}

// NewJudgeWorker 创建后台判题 Worker。
func NewJudgeWorker(q queue.Queue, submissions repository.SubmissionRepository, problems repository.ProblemRepository, judgeSvc judger.Executor, collector *stats.Collector, concurrency int) *JudgeWorker {
	if concurrency <= 0 {
		concurrency = 1
	}
	if collector != nil {
		collector.SetMaxWorkers(concurrency)
	}
	return &JudgeWorker{
		q:             q,
		submissions:   submissions,
		problems:      problems,
		judger:        judgeSvc,
		stats:         collector,
		concurrency:   concurrency,
		acquireTimout: 5 * time.Second,
	}
}

// Start 开始从队列消费消息。
func (w *JudgeWorker) Start(ctx context.Context) error {
	ch, err := w.q.Consume(ctx)
	if err != nil {
		return err
	}
	tokens := make(chan struct{}, w.concurrency)
	for i := 0; i < w.concurrency; i++ {
		tokens <- struct{}{}
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				acquireCtx, cancel := context.WithTimeout(ctx, w.acquireTimout)
				acquired := false
				select {
				case <-acquireCtx.Done():
				case <-ctx.Done():
				case <-tokens:
					acquired = true
				}
				cancel()
				if !acquired {
					w.rejectBusy(ctx, msg)
					continue
				}
				if w.stats != nil {
					w.stats.WorkerStarted()
				}
				go func(m domain.SubmissionMessage) {
					defer func() {
						tokens <- struct{}{}
						if w.stats != nil {
							w.stats.WorkerFinished()
						}
					}()
					w.handle(ctx, m)
				}(msg)
			}
		}
	}()
	return nil
}

// handle processes a single submission message.
func (w *JudgeWorker) handle(ctx context.Context, msg domain.SubmissionMessage) {
	logger.Info("judge.start", msg.TraceID, "judge task started", map[string]any{
		"submissionId": msg.SubmissionID,
		"problemId":    msg.ProblemID,
		"language":     msg.Language,
	})
	sub, err := w.submissions.GetByID(ctx, msg.SubmissionID)
	if err != nil {
		return
	}
	now := time.Now()
	sub.Status = domain.StatusQueueing
	sub.QueueStartedAt = &now
	_ = w.submissions.Update(ctx, sub)

	problem, err := w.problems.GetByID(ctx, sub.ProblemID)
	if err != nil {
		w.fail(ctx, sub, domain.StatusSystemError, err.Error())
		return
	}
	testCases, err := w.problems.ListCases(ctx, sub.ProblemID)
	if err != nil {
		w.fail(ctx, sub, domain.StatusSystemError, err.Error())
		return
	}

	judgeStart := time.Now()
	sub.Status = domain.StatusCompiling
	sub.JudgeStartedAt = &judgeStart
	_ = w.submissions.Update(ctx, sub)

	result, err := w.judger.Judge(ctx, domain.JudgeRequest{
		SubmissionID:  sub.ID,
		ProblemID:     sub.ProblemID,
		TraceID:       msg.TraceID,
		Language:      sub.Language,
		Code:          sub.Code,
		TimeLimitMs:   problem.TimeLimitMs,
		MemoryLimitMB: problem.MemoryLimitMB,
		OutputLimitKB: problem.OutputLimitKB,
		TestCases:     testCases,
	})
	if err != nil {
		w.fail(ctx, sub, domain.StatusSystemError, err.Error())
		return
	}

	if len(result.CaseResults) > 0 && result.Status != domain.StatusCompileError {
		sub.Status = domain.StatusRunning
		_ = w.submissions.Update(ctx, sub)
	}

	finished := time.Now()
	sub.Status = result.Status
	sub.RuntimeMs = result.RuntimeMs
	sub.MemoryKB = result.MemoryKB
	sub.CompileOutput = result.CompileOut
	sub.ErrorMessage = result.ErrorMessage
	sub.FinishedAt = &finished
	_ = w.submissions.Update(ctx, sub)
	_ = w.submissions.SaveCaseResults(ctx, sub.ID, result.CaseResults)
	if w.stats != nil {
		w.stats.RecordStatus(result.Status)
	}
	logger.Info("judge.finish", msg.TraceID, "judge task finished", map[string]any{
		"submissionId": sub.ID,
		"status":       result.Status,
		"runtimeMs":    result.RuntimeMs,
		"memoryKB":     result.MemoryKB,
		"caseCount":    len(result.CaseResults),
	})
}

// fail marks the submission as failed.
func (w *JudgeWorker) fail(ctx context.Context, sub *domain.Submission, status domain.SubmissionStatus, message string) {
	now := time.Now()
	sub.Status = status
	sub.ErrorMessage = message
	sub.FinishedAt = &now
	_ = w.submissions.Update(ctx, sub)
}

// rejectBusy marks the submission as system busy when the worker pool is full.
func (w *JudgeWorker) rejectBusy(ctx context.Context, msg domain.SubmissionMessage) {
	logger.Error("judge.reject", msg.TraceID, "judge task rejected", map[string]any{
		"submissionId": msg.SubmissionID,
		"reason":       "worker_busy",
	})
	sub, err := w.submissions.GetByID(ctx, msg.SubmissionID)
	if err != nil {
		return
	}
	if w.stats != nil {
		w.stats.RecordBusyReject()
		w.stats.RecordStatus(domain.StatusSystemError)
	}
	w.fail(ctx, sub, domain.StatusSystemError, "judge worker busy")
}
