package judger

import "context"

// JudgerClient is the interface for judging submissions. Both the built-in
// gRPC Client and the RemoteJudger adapter satisfy this interface.
type JudgerClient interface {
	Judge(ctx context.Context, req *JudgeRequest) (*JudgeResponse, error)
	Close() error
}

type TestCase struct {
	Input    string `json:"input"`
	Expected string `json:"expected"`
}

type JudgeRequest struct {
	SubmissionID  uint64     `json:"submission_id"`
	ProblemID     uint64     `json:"problem_id"`
	TraceID       string     `json:"trace_id,omitempty"`
	Language      string     `json:"language"`
	Code          string     `json:"code"`
	TimeLimitMS   int32      `json:"time_limit_ms"`
	MemoryLimitMB int32      `json:"memory_limit_mb"`
	OutputLimitKB int32      `json:"output_limit_kb"`
	TestCases     []TestCase `json:"test_cases"`
}

type CaseResult struct {
	CaseNo        int32  `json:"case_no"`
	Status        string `json:"status"`
	RuntimeMS     int32  `json:"runtime_ms"`
	MemoryKB      int32  `json:"memory_kb"`
	StdoutBytes   int32  `json:"stdout_bytes"`
	StderrBytes   int32  `json:"stderr_bytes"`
	Signal        string `json:"signal,omitempty"`
	StdoutPreview string `json:"stdout_preview,omitempty"`
	StderrPreview string `json:"stderr_preview,omitempty"`
}

type JudgeResponse struct {
	SubmissionID uint64       `json:"submission_id"`
	Status       string       `json:"status"`
	RuntimeMS    int32        `json:"runtime_ms"`
	MemoryMB     string       `json:"memory_mb"`
	MemoryKB     int32        `json:"memory_kb"`
	CompileOut   string       `json:"compile_output,omitempty"`
	ErrorMessage string       `json:"error_message,omitempty"`
	CaseResults  []CaseResult `json:"case_results,omitempty"`
}

// Fully qualified gRPC method name for the Judge RPC.
const MethodJudge = "/judger.Judger/Judge"
