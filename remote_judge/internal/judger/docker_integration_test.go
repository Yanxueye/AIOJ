package judger

import (
	"context"
	"os/exec"
	"testing"

	"remote_judge/internal/domain"
	"remote_judge/internal/sandbox"
)

// newPooledService creates a Service backed by a container pool for integration tests.
func newPooledService(t *testing.T) *Service {
	t.Helper()
	inner := &sandbox.DockerCLISandbox{}
	ps := sandbox.NewPooledSandbox(inner, 4)
	t.Cleanup(func() { ps.Close(context.Background()) })
	return NewService(ps)
}

// TestJudgeWithDockerAccepted uses real Docker to verify accepted C++ code.
func TestJudgeWithDockerAccepted(t *testing.T) {
	if testing.Short() {
		t.Skip("skip docker integration test in short mode")
	}
	if !dockerReady() {
		t.Skip("docker not ready")
	}

	t.Logf(">>> [C++17] a+b code -> Accepted (2 test cases)")
	svc := newPooledService(t)
	result, err := svc.Judge(context.Background(), domain.JudgeRequest{
		SubmissionID:  900001,
		ProblemID:     1001,
		Language:      "cpp17",
		Code:          "#include <iostream>\nint main(){int a,b;std::cin>>a>>b;std::cout<<a+b<<\"\\n\";}",
		TimeLimitMs:   2000,
		MemoryLimitMB: 256,
		OutputLimitKB: 1024,
		TestCases: []domain.TestCase{
			{ProblemID: 1001, CaseNo: 1, Input: "1 2\n", Expected: "3\n"},
			{ProblemID: 1001, CaseNo: 2, Input: "10 20\n", Expected: "30\n"},
		},
	})
	if err != nil {
		t.Fatalf("Judge() error = %v", err)
	}
	t.Logf("    status=%s | runtime=%dms | memory=%dKB | compile=%s", result.Status, result.RuntimeMs, result.MemoryKB, nonEmpty(result.CompileOut, "ok"))
	for _, cr := range result.CaseResults {
		t.Logf("    case #%d: status=%s runtime=%dms memory=%dKB stdout=%dB", cr.CaseNo, cr.Status, cr.RuntimeMs, cr.MemoryKB, cr.StdoutBytes)
	}
	if result.Status != domain.StatusAccepted {
		t.Fatalf("unexpected status: %s, compile=%s, err=%s", result.Status, result.CompileOut, result.ErrorMessage)
	}
}

// TestJudgeWithDockerWrongAnswer uses real Docker to verify wrong-answer classification.
func TestJudgeWithDockerWrongAnswer(t *testing.T) {
	if testing.Short() {
		t.Skip("skip docker integration test in short mode")
	}
	if !dockerReady() {
		t.Skip("docker not ready")
	}

	t.Logf(">>> [C++17] a-b code -> Wrong Answer (expected a+b)")
	svc := newPooledService(t)
	result, err := svc.Judge(context.Background(), domain.JudgeRequest{
		SubmissionID:  900002,
		ProblemID:     1001,
		Language:      "cpp17",
		Code:          "#include <iostream>\nint main(){int a,b;std::cin>>a>>b;std::cout<<a-b<<\"\\n\";}",
		TimeLimitMs:   2000,
		MemoryLimitMB: 256,
		OutputLimitKB: 1024,
		TestCases: []domain.TestCase{
			{ProblemID: 1001, CaseNo: 1, Input: "1 2\n", Expected: "3\n"},
		},
	})
	if err != nil {
		t.Fatalf("Judge() error = %v", err)
	}
	t.Logf("    status=%s | runtime=%dms | memory=%dKB", result.Status, result.RuntimeMs, result.MemoryKB)
	for _, cr := range result.CaseResults {
		t.Logf("    case #%d: status=%s actual=%q expected=%q", cr.CaseNo, cr.Status, cr.StdoutPreview, "3")
	}
	if result.Status != domain.StatusWrongAnswer {
		t.Fatalf("unexpected status: %s", result.Status)
	}
}

// TestJudgeWithDockerCompileError uses real Docker to verify compile-error classification.
func TestJudgeWithDockerCompileError(t *testing.T) {
	if testing.Short() {
		t.Skip("skip docker integration test in short mode")
	}
	if !dockerReady() {
		t.Skip("docker not ready")
	}

	t.Logf(">>> [C++17] syntax error -> Compile Error")
	svc := newPooledService(t)
	result, err := svc.Judge(context.Background(), domain.JudgeRequest{
		SubmissionID:  900003,
		ProblemID:     1001,
		Language:      "cpp17",
		Code:          "#include <iostream>\nint main( { return 0; }",
		TimeLimitMs:   2000,
		MemoryLimitMB: 256,
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
		t.Fatalf("unexpected status: %s, compile=%s", result.Status, result.CompileOut)
	}
}

// TestJudgeWithDockerRuntimeError uses real Docker to verify runtime-error classification.
func TestJudgeWithDockerRuntimeError(t *testing.T) {
	if testing.Short() {
		t.Skip("skip docker integration test in short mode")
	}
	if !dockerReady() {
		t.Skip("docker not ready")
	}

	t.Logf(">>> [C++17] null pointer dereference -> Runtime Error")
	svc := newPooledService(t)
	result, err := svc.Judge(context.Background(), domain.JudgeRequest{
		SubmissionID:  900004,
		ProblemID:     1001,
		Language:      "cpp17",
		Code:          "#include <iostream>\nint main(){int *p=nullptr; std::cout<<*p;}",
		TimeLimitMs:   2000,
		MemoryLimitMB: 256,
		OutputLimitKB: 1024,
		TestCases: []domain.TestCase{
			{ProblemID: 1001, CaseNo: 1, Input: "1 2\n", Expected: "3\n"},
		},
	})
	if err != nil {
		t.Fatalf("Judge() error = %v", err)
	}
	t.Logf("    status=%s | runtime=%dms | memory=%dKB", result.Status, result.RuntimeMs, result.MemoryKB)
	for _, cr := range result.CaseResults {
		t.Logf("    case #%d: status=%s signal=%s stderr=%s", cr.CaseNo, cr.Status, cr.Signal, cr.StderrPreview)
	}
	if result.Status != domain.StatusRuntimeError {
		t.Fatalf("unexpected status: %s", result.Status)
	}
}

// TestJudgeWithDockerTimeLimitExceeded uses real Docker to verify timeout classification.
func TestJudgeWithDockerTimeLimitExceeded(t *testing.T) {
	if testing.Short() {
		t.Skip("skip docker integration test in short mode")
	}
	if !dockerReady() {
		t.Skip("docker not ready")
	}

	t.Logf(">>> [C++17] sleep(3s) with 300ms limit -> Time Limit Exceeded")
	svc := newPooledService(t)
	result, err := svc.Judge(context.Background(), domain.JudgeRequest{
		SubmissionID:  900005,
		ProblemID:     1001,
		Language:      "cpp17",
		Code:          "#include <chrono>\n#include <thread>\nint main(){std::this_thread::sleep_for(std::chrono::seconds(3));}",
		TimeLimitMs:   300,
		MemoryLimitMB: 256,
		OutputLimitKB: 1024,
		TestCases: []domain.TestCase{
			{ProblemID: 1001, CaseNo: 1, Input: "", Expected: ""},
		},
	})
	if err != nil {
		t.Fatalf("Judge() error = %v", err)
	}
	t.Logf("    status=%s | runtime=%dms | limit=%dms", result.Status, result.RuntimeMs, 300)
	if result.Status != domain.StatusTimeLimitExceeded {
		t.Fatalf("unexpected status: %s", result.Status)
	}
}

// TestJudgeWithDockerPythonAccepted uses real Docker to verify Python submissions.
func TestJudgeWithDockerPythonAccepted(t *testing.T) {
	if testing.Short() {
		t.Skip("skip docker integration test in short mode")
	}
	if !dockerReady() {
		t.Skip("docker not ready")
	}

	t.Logf(">>> [Python 3.11] a+b code -> Accepted")
	svc := newPooledService(t)
	result, err := svc.Judge(context.Background(), domain.JudgeRequest{
		SubmissionID:  900006,
		ProblemID:     1001,
		Language:      "python3.11",
		Code:          "a,b=map(int,input().split())\nprint(a+b)",
		TimeLimitMs:   2000,
		MemoryLimitMB: 256,
		OutputLimitKB: 1024,
		TestCases: []domain.TestCase{
			{ProblemID: 1001, CaseNo: 1, Input: "4 5\n", Expected: "9\n"},
		},
	})
	if err != nil {
		t.Fatalf("Judge() error = %v", err)
	}
	t.Logf("    status=%s | runtime=%dms | memory=%dKB", result.Status, result.RuntimeMs, result.MemoryKB)
	for _, cr := range result.CaseResults {
		t.Logf("    case #%d: status=%s runtime=%dms memory=%dKB stdout=%dB", cr.CaseNo, cr.Status, cr.RuntimeMs, cr.MemoryKB, cr.StdoutBytes)
	}
	if result.Status != domain.StatusAccepted {
		t.Fatalf("unexpected status: %s", result.Status)
	}
}

// TestJudgeWithDockerGoAccepted uses real Docker to verify Go submissions.
func TestJudgeWithDockerGoAccepted(t *testing.T) {
	if testing.Short() {
		t.Skip("skip docker integration test in short mode")
	}
	if !dockerReady() {
		t.Skip("docker not ready")
	}

	t.Logf(">>> [Go 1.22] a+b code -> Accepted")
	svc := newPooledService(t)
	result, err := svc.Judge(context.Background(), domain.JudgeRequest{
		SubmissionID:  900007,
		ProblemID:     1001,
		Language:      "go1.22",
		Code:          "package main\nimport \"fmt\"\nfunc main(){var a,b int; fmt.Scan(&a,&b); fmt.Println(a+b)}",
		TimeLimitMs:   2000,
		MemoryLimitMB: 256,
		OutputLimitKB: 1024,
		TestCases: []domain.TestCase{
			{ProblemID: 1001, CaseNo: 1, Input: "6 7\n", Expected: "13\n"},
		},
	})
	if err != nil {
		t.Fatalf("Judge() error = %v", err)
	}
	t.Logf("    status=%s | runtime=%dms | memory=%dKB | compile=%s", result.Status, result.RuntimeMs, result.MemoryKB, nonEmpty(result.CompileOut, "ok"))
	for _, cr := range result.CaseResults {
		t.Logf("    case #%d: status=%s runtime=%dms memory=%dKB stdout=%dB", cr.CaseNo, cr.Status, cr.RuntimeMs, cr.MemoryKB, cr.StdoutBytes)
	}
	if result.Status != domain.StatusAccepted {
		t.Fatalf("unexpected status: %s, compile=%s, err=%s", result.Status, result.CompileOut, result.ErrorMessage)
	}
}

// TestJudgeWithDockerOutputLimitExceeded uses real Docker to verify output limit classification.
func TestJudgeWithDockerOutputLimitExceeded(t *testing.T) {
	if testing.Short() {
		t.Skip("skip docker integration test in short mode")
	}
	if !dockerReady() {
		t.Skip("docker not ready")
	}

	t.Logf(">>> [Python 3.11] print 4096 bytes with 1KB limit -> Output Limit Exceeded")
	svc := newPooledService(t)
	result, err := svc.Judge(context.Background(), domain.JudgeRequest{
		SubmissionID:  900008,
		ProblemID:     1001,
		Language:      "python3.11",
		Code:          "print('x' * 4096)",
		TimeLimitMs:   2000,
		MemoryLimitMB: 256,
		OutputLimitKB: 1,
		TestCases: []domain.TestCase{
			{ProblemID: 1001, CaseNo: 1, Input: "", Expected: ""},
		},
	})
	if err != nil {
		t.Fatalf("Judge() error = %v", err)
	}
	t.Logf("    status=%s | stdoutBytes=4096 | limit=1KB", result.Status)
	for _, cr := range result.CaseResults {
		t.Logf("    case #%d: status=%s stdoutBytes=%d", cr.CaseNo, cr.Status, cr.StdoutBytes)
	}
	if result.Status != domain.StatusOutputLimitExceeded {
		t.Fatalf("unexpected status: %s", result.Status)
	}
}

// dockerReady checks whether the local Docker daemon is reachable.
func dockerReady() bool {
	cmd := exec.Command("docker", "version", "--format", "{{.Server.Version}}")
	return cmd.Run() == nil
}
