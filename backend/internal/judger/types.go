package judger

// Types shared between the OJ backend (client) and the sandboxed judger
// (server). They mirror the messages declared in proto/judger.proto and are
// serialised over gRPC using the registered JSON codec (see codec.go),
// which keeps the build self-contained (no protoc dependency).

type TestCase struct {
	Input    string `json:"input"`
	Expected string `json:"expected"`
}

type JudgeRequest struct {
	SubmissionID  uint64     `json:"submission_id"`
	ProblemID     uint64     `json:"problem_id"`
	Language      string     `json:"language"`
	Code          string     `json:"code"`
	TimeLimitMS   int32      `json:"time_limit_ms"`
	MemoryLimitMB int32      `json:"memory_limit_mb"`
	TestCases     []TestCase `json:"test_cases"`
}

type JudgeResponse struct {
	SubmissionID uint64 `json:"submission_id"`
	Status       string `json:"status"`
	RuntimeMS    int32  `json:"runtime_ms"`
	MemoryMB     string `json:"memory_mb"`
	ErrorMessage string `json:"error_message,omitempty"`
}

// Fully qualified gRPC method name for the Judge RPC.
const MethodJudge = "/judger.Judger/Judge"
