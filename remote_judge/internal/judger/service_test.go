package judger

import (
	"context"
	"testing"
	"time"

	"remote_judge/internal/domain"
	"remote_judge/internal/sandbox"
)

// TestJudgeAccepted verifies a successful judge flow.
func TestJudgeAccepted(t *testing.T) {
	t.Logf(">>> [Mock] 2 test cases all Accepted")
	svc := NewService(&sandbox.MockSandbox{})
	result, err := svc.Judge(context.Background(), domain.JudgeRequest{
		SubmissionID:  1,
		ProblemID:     1001,
		Language:      "cpp17",
		Code:          "dummy",
		TimeLimitMs:   1000,
		MemoryLimitMB: 128,
		OutputLimitKB: 1024,
		TestCases: []domain.TestCase{
			{ProblemID: 1001, CaseNo: 1, Input: "1 2\n", Expected: "3\n"},
			{ProblemID: 1001, CaseNo: 2, Input: "10 20\n", Expected: "30\n"},
		},
	})
	if err != nil {
		t.Fatalf("Judge() error = %v", err)
	}
	t.Logf("    status=%s | cases=%d", result.Status, len(result.CaseResults))
	for _, cr := range result.CaseResults {
		t.Logf("    case #%d: status=%s stdout=%dB", cr.CaseNo, cr.Status, cr.StdoutBytes)
	}
	if result.Status != domain.StatusAccepted {
		t.Fatalf("unexpected status: %s", result.Status)
	}
	if len(result.CaseResults) != 2 {
		t.Fatalf("unexpected case count: %d", len(result.CaseResults))
	}
	if result.CaseResults[0].StdoutBytes == 0 {
		t.Fatal("expected stdout bytes")
	}
}

// TestJudgeWrongAnswer verifies wrong answer classification.
func TestJudgeWrongAnswer(t *testing.T) {
	t.Logf(">>> [Mock] wrong input -> Wrong Answer")
	svc := NewService(&sandbox.MockSandbox{})
	result, err := svc.Judge(context.Background(), domain.JudgeRequest{
		SubmissionID:  2,
		ProblemID:     1002,
		Language:      "python3.11",
		Code:          "dummy",
		TimeLimitMs:   1000,
		MemoryLimitMB: 128,
		OutputLimitKB: 1024,
		TestCases: []domain.TestCase{
			{ProblemID: 1002, CaseNo: 1, Input: "wrong\n", Expected: "right\n"},
		},
	})
	if err != nil {
		t.Fatalf("Judge() error = %v", err)
	}
	t.Logf("    status=%s", result.Status)
	if result.Status != domain.StatusWrongAnswer {
		t.Fatalf("unexpected status: %s", result.Status)
	}
}

// TestJudgeCompileError verifies compile error classification.
func TestJudgeCompileError(t *testing.T) {
	t.Logf(">>> [Mock] compile_error code -> Compile Error")
	svc := NewService(&sandbox.MockSandbox{})
	result, err := svc.Judge(context.Background(), domain.JudgeRequest{
		SubmissionID:  3,
		ProblemID:     1001,
		Language:      "cpp17",
		Code:          "compile_error",
		TimeLimitMs:   1000,
		MemoryLimitMB: 128,
		OutputLimitKB: 1024,
		TestCases: []domain.TestCase{
			{ProblemID: 1001, CaseNo: 1, Input: "1 2\n", Expected: "3\n"},
		},
	})
	if err != nil {
		t.Fatalf("Judge() error = %v", err)
	}
	t.Logf("    status=%s | compile=%s", result.Status, result.CompileOut)
	if result.Status != domain.StatusCompileError {
		t.Fatalf("unexpected status: %s", result.Status)
	}
}

// TestJudgeMemoryLimitExceeded verifies memory-limit classification.
func TestJudgeMemoryLimitExceeded(t *testing.T) {
	t.Logf(">>> [Custom Sandbox] memory=128MB+1 -> Memory Limit Exceeded")
	t.Logf("    limit=128MB | used=128MB+1B")
	svc := NewService(memoryLimitSandbox{})
	result, err := svc.Judge(context.Background(), domain.JudgeRequest{
		SubmissionID:  4,
		ProblemID:     1001,
		Language:      "cpp17",
		Code:          "dummy",
		TimeLimitMs:   1000,
		MemoryLimitMB: 128,
		OutputLimitKB: 1024,
		TestCases: []domain.TestCase{
			{ProblemID: 1001, CaseNo: 1, Input: "mle\n", Expected: "ok\n"},
		},
	})
	if err != nil {
		t.Fatalf("Judge() error = %v", err)
	}
	t.Logf("    status=%s | runtime=%dms | memory=%dKB", result.Status, result.RuntimeMs, result.MemoryKB)
	if result.Status != domain.StatusMemoryLimitExceeded {
		t.Fatalf("unexpected status: %s", result.Status)
	}
}

// TestJudgeRunModeSkipsOutputCompare verifies Run Code mode returns Accepted
// on successful execution even when expected output is intentionally ignored.
func TestJudgeRunModeSkipsOutputCompare(t *testing.T) {
	t.Logf(">>> [Mock] run mode skips output compare and only validates execution")
	svc := NewService(&sandbox.MockSandbox{})
	result, err := svc.Judge(context.Background(), domain.JudgeRequest{
		SubmissionID:  5,
		ProblemID:     1001,
		Language:      "cpp17",
		Code:          "dummy",
		TimeLimitMs:   1000,
		MemoryLimitMB: 128,
		OutputLimitKB: 1024,
		RunMode:       "run",
		TestCases: []domain.TestCase{
			{ProblemID: 1001, CaseNo: 1, Input: "wrong\n", Expected: "never checked\n"},
		},
	})
	if err != nil {
		t.Fatalf("Judge() error = %v", err)
	}
	t.Logf("    status=%s", result.Status)
	if result.Status != domain.StatusAccepted {
		t.Fatalf("unexpected status: %s", result.Status)
	}
}

func TestCompareOutputRejectsNumericPrefixOnly(t *testing.T) {
	t.Logf(">>> Output compare: numeric prefixes must not hide trailing junk")
	cases := []struct {
		actual   string
		expected string
	}{
		{actual: "1 99\n", expected: "1 2\n"},
		{actual: "3abc\n", expected: "3\n"},
		{actual: "4\nextra\n", expected: "4\n"},
	}
	for _, tc := range cases {
		if compareOutput(tc.actual, tc.expected) {
			t.Fatalf("compareOutput(%q, %q) = true, want false", tc.actual, tc.expected)
		}
	}
}

func TestCompareOutputFloatTokens(t *testing.T) {
	t.Logf(">>> Output compare: token-wise float tolerance supports scientific notation")
	cases := []struct {
		actual   string
		expected string
	}{
		{actual: "1.0000004 2e-6\n", expected: "1.0 0.000002\n"},
		{actual: "-.5\n", expected: "-0.5000001\n"},
	}
	for _, tc := range cases {
		if !compareOutput(tc.actual, tc.expected) {
			t.Fatalf("compareOutput(%q, %q) = false, want true", tc.actual, tc.expected)
		}
	}
}

type memoryLimitSandbox struct{}

func (memoryLimitSandbox) Compile(context.Context, sandbox.ExecRequest) (sandbox.ExecResult, error) {
	return sandbox.ExecResult{ExitCode: 0}, nil
}

func (memoryLimitSandbox) Run(context.Context, sandbox.ExecRequest) (sandbox.ExecResult, error) {
	return sandbox.ExecResult{
		ExitCode: 0,
		Stdout:   "ok\n",
		Runtime:  10 * time.Millisecond,
		MemoryKB: 128*1024 + 1,
	}, nil
}

func (memoryLimitSandbox) Health(context.Context) error {
	return nil
}
