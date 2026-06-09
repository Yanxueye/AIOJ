package judger

import (
	"context"
	"testing"
	"time"

	"remote_judge/internal/domain"
	"remote_judge/internal/sandbox"
)

// prioritySandbox returns precisely controlled results for testing the priority chain.
type prioritySandbox struct {
	compileRes sandbox.ExecResult
	runRes     sandbox.ExecResult
}

func (s prioritySandbox) Compile(context.Context, sandbox.ExecRequest) (sandbox.ExecResult, error) {
	return s.compileRes, nil
}

func (s prioritySandbox) Run(context.Context, sandbox.ExecRequest) (sandbox.ExecResult, error) {
	return s.runRes, nil
}

func (s prioritySandbox) Health(context.Context) error { return nil }

// TestPriorityTLEOverMLE verifies TLE is reported first when both timeout and OOM occur.
func TestPriorityTLEOverMLE(t *testing.T) {
	t.Logf(">>> Priority: TLE + OOM -> TLE (TLE wins)")
	svc := NewService(prioritySandbox{
		compileRes: sandbox.ExecResult{ExitCode: 0},
		runRes: sandbox.ExecResult{
			ExitCode:  137,
			TimedOut:  true,
			Runtime:   1500 * time.Millisecond,
			MemoryKB:  256*1024 + 1,
			OOMKilled: true,
		},
	})
	result, err := svc.Judge(context.Background(), domain.JudgeRequest{
		SubmissionID: 100, ProblemID: 1001, Language: "cpp17", Code: "x",
		TimeLimitMs: 1000, MemoryLimitMB: 256, OutputLimitKB: 1024,
		TestCases: []domain.TestCase{{ProblemID: 1001, CaseNo: 1, Input: "", Expected: ""}},
	})
	if err != nil {
		t.Fatalf("Judge() error = %v", err)
	}
	t.Logf("    status=%s | runtime=%dms | memory=%dKB | oom=%v", result.Status, result.RuntimeMs, result.MemoryKB, true)
	if result.Status != domain.StatusTimeLimitExceeded {
		t.Fatalf("expected TLE (priority over MLE), got %s", result.Status)
	}
}

// TestPriorityTLEOverRE verifies TLE is reported first when both timeout and runtime error occur.
func TestPriorityTLEOverRE(t *testing.T) {
	t.Logf(">>> Priority: TLE + RE(exit=1) -> TLE (TLE wins)")
	svc := NewService(prioritySandbox{
		compileRes: sandbox.ExecResult{ExitCode: 0},
		runRes: sandbox.ExecResult{
			ExitCode: 1,
			TimedOut: true,
			Runtime:  1500 * time.Millisecond,
			MemoryKB: 4096,
		},
	})
	result, err := svc.Judge(context.Background(), domain.JudgeRequest{
		SubmissionID: 101, ProblemID: 1001, Language: "cpp17", Code: "x",
		TimeLimitMs: 1000, MemoryLimitMB: 256, OutputLimitKB: 1024,
		TestCases: []domain.TestCase{{ProblemID: 1001, CaseNo: 1, Input: "", Expected: ""}},
	})
	if err != nil {
		t.Fatalf("Judge() error = %v", err)
	}
	t.Logf("    status=%s | runtime=%dms", result.Status, result.RuntimeMs)
	if result.Status != domain.StatusTimeLimitExceeded {
		t.Fatalf("expected TLE (priority over RE), got %s", result.Status)
	}
}

// TestPriorityMLEOverOLE verifies MLE is reported before output-limit check.
func TestPriorityMLEOverOLE(t *testing.T) {
	t.Logf(">>> Priority: MLE + excessive output -> MLE (MLE wins)")
	svc := NewService(prioritySandbox{
		compileRes: sandbox.ExecResult{ExitCode: 0},
		runRes: sandbox.ExecResult{
			ExitCode:    0,
			Stdout:      string(make([]byte, 2048)),
			Runtime:     10 * time.Millisecond,
			MemoryKB:    256*1024 + 1,
			StdoutBytes: 2048,
		},
	})
	result, err := svc.Judge(context.Background(), domain.JudgeRequest{
		SubmissionID: 102, ProblemID: 1001, Language: "cpp17", Code: "x",
		TimeLimitMs: 1000, MemoryLimitMB: 256, OutputLimitKB: 1,
		TestCases: []domain.TestCase{{ProblemID: 1001, CaseNo: 1, Input: "", Expected: ""}},
	})
	if err != nil {
		t.Fatalf("Judge() error = %v", err)
	}
	t.Logf("    status=%s | memory=%dKB | stdoutBytes=%d", result.Status, result.MemoryKB, 2048)
	if result.Status != domain.StatusMemoryLimitExceeded {
		t.Fatalf("expected MLE (priority over OLE), got %s", result.Status)
	}
}

// TestPriorityOLERejectedAboveLimit verifies output limit check works independently.
func TestPriorityOLERejectedAboveLimit(t *testing.T) {
	t.Logf(">>> OLE: stdout 4096B > 1KB limit -> Output Limit Exceeded")
	svc := NewService(prioritySandbox{
		compileRes: sandbox.ExecResult{ExitCode: 0},
		runRes: sandbox.ExecResult{
			ExitCode:    0,
			Stdout:      string(make([]byte, 4096)),
			Runtime:     5 * time.Millisecond,
			MemoryKB:    1024,
			StdoutBytes: 4096,
		},
	})
	result, err := svc.Judge(context.Background(), domain.JudgeRequest{
		SubmissionID: 103, ProblemID: 1001, Language: "python3.11", Code: "x",
		TimeLimitMs: 1000, MemoryLimitMB: 128, OutputLimitKB: 1,
		TestCases: []domain.TestCase{{ProblemID: 1001, CaseNo: 1, Input: "", Expected: ""}},
	})
	if err != nil {
		t.Fatalf("Judge() error = %v", err)
	}
	t.Logf("    status=%s | stdoutBytes=%d | limit=1KB", result.Status, 4096)
	if result.Status != domain.StatusOutputLimitExceeded {
		t.Fatalf("expected Output Limit Exceeded, got %s", result.Status)
	}
}

// TestPriorityOLEOverRE verifies OLE is reported before runtime error.
func TestPriorityOLEOverRE(t *testing.T) {
	t.Logf(">>> Priority: excessive output + exit(1) -> OLE (OLE wins)")
	svc := NewService(prioritySandbox{
		compileRes: sandbox.ExecResult{ExitCode: 0},
		runRes: sandbox.ExecResult{
			ExitCode:   1,
			Stdout:     string(make([]byte, 4096)),
			Runtime:    5 * time.Millisecond,
			MemoryKB:   1024,
			OOMKilled:  false,
			TimedOut:   false,
		},
	})
	result, err := svc.Judge(context.Background(), domain.JudgeRequest{
		SubmissionID: 104, ProblemID: 1001, Language: "python3.11", Code: "x",
		TimeLimitMs: 1000, MemoryLimitMB: 128, OutputLimitKB: 1,
		TestCases: []domain.TestCase{{ProblemID: 1001, CaseNo: 1, Input: "", Expected: ""}},
	})
	if err != nil {
		t.Fatalf("Judge() error = %v", err)
	}
	t.Logf("    status=%s | stdoutBytes=%d | exitCode=1 (ignored)", result.Status, 4096)
	if result.Status != domain.StatusOutputLimitExceeded {
		t.Fatalf("expected OLE (priority over RE), got %s", result.Status)
	}
}

// TestPriorityMLEOverWA verifies MLE is reported before wrong-answer.
func TestPriorityMLEOverWA(t *testing.T) {
	t.Logf(">>> Priority: MLE(exit=137) + correct output -> MLE (MLE wins)")
	svc := NewService(prioritySandbox{
		compileRes: sandbox.ExecResult{ExitCode: 0},
		runRes: sandbox.ExecResult{
			ExitCode:  137,
			Stdout:    "3\n",
			Runtime:   50 * time.Millisecond,
			MemoryKB:  256*1024 + 1,
			OOMKilled: true,
		},
	})
	result, err := svc.Judge(context.Background(), domain.JudgeRequest{
		SubmissionID: 105, ProblemID: 1001, Language: "cpp17", Code: "x",
		TimeLimitMs: 1000, MemoryLimitMB: 256, OutputLimitKB: 1024,
		TestCases: []domain.TestCase{{ProblemID: 1001, CaseNo: 1, Input: "1 2\n", Expected: "3\n"}},
	})
	if err != nil {
		t.Fatalf("Judge() error = %v", err)
	}
	t.Logf("    status=%s | memory=%dKB | output matches expected", result.Status, result.MemoryKB)
	if result.Status != domain.StatusMemoryLimitExceeded {
		t.Fatalf("expected MLE (priority over WA), got %s", result.Status)
	}
}

// TestSystemErrorOnUnsupportedLanguage verifies system error for unknown languages.
func TestSystemErrorOnUnsupportedLanguage(t *testing.T) {
	t.Logf(">>> System Error: unsupported language 'rust'")
	svc := NewService(prioritySandbox{})
	result, err := svc.Judge(context.Background(), domain.JudgeRequest{
		SubmissionID: 106, ProblemID: 1001, Language: "rust", Code: "fn main(){}",
		TimeLimitMs: 1000, MemoryLimitMB: 128, OutputLimitKB: 1024,
		TestCases: []domain.TestCase{{ProblemID: 1001, CaseNo: 1, Input: "", Expected: ""}},
	})
	if err != nil {
		t.Fatalf("Judge() error = %v", err)
	}
	t.Logf("    status=%s | error=%s", result.Status, result.ErrorMessage)
	if result.Status != domain.StatusSystemError {
		t.Fatalf("expected System Error for unsupported language, got %s", result.Status)
	}
}

// TestCompileReturnsErrorButExitZero verifies the edge case where Compile returns an error
// with exit code 0 (system-level failure, not user code failure).
func TestCompileReturnsErrorButExitZero(t *testing.T) {
	t.Logf(">>> Compile: exitCode=0 + error -> System Error (not compile error)")
	svc := NewService(errorSandbox{compileErr: true, exitCode: 0})
	result, err := svc.Judge(context.Background(), domain.JudgeRequest{
		SubmissionID: 107, ProblemID: 1001, Language: "cpp17", Code: "x",
		TimeLimitMs: 1000, MemoryLimitMB: 128, OutputLimitKB: 1024,
		TestCases: []domain.TestCase{{ProblemID: 1001, CaseNo: 1, Input: "1 2\n", Expected: "3\n"}},
	})
	if err != nil {
		t.Fatalf("Judge() error = %v", err)
	}
	t.Logf("    status=%s | error=%s", result.Status, result.ErrorMessage)
	if result.Status != domain.StatusSystemError {
		t.Fatalf("expected System Error, got %s", result.Status)
	}
}

// TestRunReturnsErrorExitZero verifies system error when Run fails with exit code 0.
func TestRunReturnsErrorExitZero(t *testing.T) {
	t.Logf(">>> Run: exitCode=0 + error + noTimeout -> System Error")
	svc := NewService(runErrorSandbox{})
	result, err := svc.Judge(context.Background(), domain.JudgeRequest{
		SubmissionID: 108, ProblemID: 1001, Language: "python3.11", Code: "x",
		TimeLimitMs: 1000, MemoryLimitMB: 128, OutputLimitKB: 1024,
		TestCases: []domain.TestCase{{ProblemID: 1001, CaseNo: 1, Input: "", Expected: ""}},
	})
	if err != nil {
		t.Fatalf("Judge() error = %v", err)
	}
	t.Logf("    status=%s | error=%s", result.Status, result.ErrorMessage)
	if result.Status != domain.StatusSystemError {
		t.Fatalf("expected System Error, got %s", result.Status)
	}
}

// TestShortCircuitOnFirstFailure verifies judging stops after the first non-AC case.
func TestShortCircuitOnFirstFailure(t *testing.T) {
	t.Logf(">>> Short circuit: case#1 WA -> skip case#2 & case#3")
	svc := NewService(failOnFirstSandbox{})
	result, err := svc.Judge(context.Background(), domain.JudgeRequest{
		SubmissionID: 109, ProblemID: 1001, Language: "cpp17", Code: "x",
		TimeLimitMs: 1000, MemoryLimitMB: 128, OutputLimitKB: 1024,
		TestCases: []domain.TestCase{
			{ProblemID: 1001, CaseNo: 1, Input: "", Expected: "expected"},
			{ProblemID: 1001, CaseNo: 2, Input: "", Expected: ""},
			{ProblemID: 1001, CaseNo: 3, Input: "", Expected: ""},
		},
	})
	if err != nil {
		t.Fatalf("Judge() error = %v", err)
	}
	t.Logf("    status=%s | caseCount=%d (should be 3, short circuit fills remaining)", result.Status, len(result.CaseResults))
	if result.Status != domain.StatusWrongAnswer {
		t.Fatalf("expected Wrong Answer, got %s", result.Status)
	}
	if len(result.CaseResults) != 3 {
		t.Fatalf("expected 3 case results (short circuit fills remaining), got %d", len(result.CaseResults))
	}
	if result.CaseResults[1].Status != domain.StatusWrongAnswer {
		t.Fatalf("short-circuited case#2 should be WA, got %s", result.CaseResults[1].Status)
	}
}

// errorSandbox returns an error from Compile with controlled exit code.
type errorSandbox struct {
	compileErr bool
	exitCode   int
}

func (s errorSandbox) Compile(context.Context, sandbox.ExecRequest) (sandbox.ExecResult, error) {
	if s.compileErr {
		return sandbox.ExecResult{ExitCode: s.exitCode}, context.DeadlineExceeded
	}
	return sandbox.ExecResult{ExitCode: 0}, nil
}

func (s errorSandbox) Run(context.Context, sandbox.ExecRequest) (sandbox.ExecResult, error) {
	return sandbox.ExecResult{ExitCode: 0, Stdout: "3\n", Runtime: 10 * time.Millisecond, MemoryKB: 1024}, nil
}

func (s errorSandbox) Health(context.Context) error { return nil }

// runErrorSandbox returns an error from Run with exit code 0.
type runErrorSandbox struct{}

func (runErrorSandbox) Compile(context.Context, sandbox.ExecRequest) (sandbox.ExecResult, error) {
	return sandbox.ExecResult{ExitCode: 0}, nil
}

func (runErrorSandbox) Run(context.Context, sandbox.ExecRequest) (sandbox.ExecResult, error) {
	return sandbox.ExecResult{ExitCode: 0, Runtime: 10 * time.Millisecond}, context.DeadlineExceeded
}

func (runErrorSandbox) Health(context.Context) error { return nil }

// failOnFirstSandbox returns WA on the first case.
type failOnFirstSandbox struct{}

func (failOnFirstSandbox) Compile(context.Context, sandbox.ExecRequest) (sandbox.ExecResult, error) {
	return sandbox.ExecResult{ExitCode: 0}, nil
}

func (failOnFirstSandbox) Run(context.Context, sandbox.ExecRequest) (sandbox.ExecResult, error) {
	return sandbox.ExecResult{ExitCode: 0, Stdout: "wrong\n", Runtime: 5 * time.Millisecond, MemoryKB: 1024}, nil
}

func (failOnFirstSandbox) Health(context.Context) error { return nil }
