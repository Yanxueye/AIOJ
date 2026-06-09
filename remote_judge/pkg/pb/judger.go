package pb

// JudgeRequest 表示 gRPC 的判题请求体。
type JudgeRequest struct {
	SubmissionID  int64      `json:"submission_id"`
	ProblemID     int64      `json:"problem_id"`
	TraceID       string     `json:"trace_id,omitempty"`
	Language      string     `json:"language"`
	Code          string     `json:"code"`
	TimeLimitMs   int32      `json:"time_limit_ms"`
	MemoryLimitMB int32      `json:"memory_limit_mb"`
	OutputLimitKB int32      `json:"output_limit_kb"`
	TestCases     []TestCase `json:"test_cases"`
}

// TestCase 表示 gRPC 请求中的测试点。
type TestCase struct {
	CaseNo   int32  `json:"case_no"`
	Input    string `json:"input,omitempty"`
	Expected string `json:"expected,omitempty"`
}

// JudgeResponse 表示 gRPC 的判题响应体。
type JudgeResponse struct {
	SubmissionID int64             `json:"submission_id"`
	Status       string            `json:"status"`
	RuntimeMs    int32             `json:"runtime_ms"`
	MemoryKB     int32             `json:"memory_kb"`
	CompileOut   string            `json:"compile_output"`
	ErrorMessage string            `json:"error_message"`
	CaseResults  []CaseResult      `json:"case_results"`
}

// CaseResult 表示单个测试点响应。
type CaseResult struct {
	CaseNo        int32  `json:"case_no"`
	Status        string `json:"status"`
	RuntimeMs     int32  `json:"runtime_ms"`
	MemoryKB      int32  `json:"memory_kb"`
	StdoutBytes   int32  `json:"stdout_bytes"`
	StderrBytes   int32  `json:"stderr_bytes"`
	Signal        string `json:"signal,omitempty"`
	StdoutPreview string `json:"stdout_preview"`
	StderrPreview string `json:"stderr_preview"`
}

// HealthRequest 表示健康检查请求。
type HealthRequest struct{}

// HealthResponse 表示健康检查响应。
type HealthResponse struct {
	Status             string   `json:"status"`
	DockerReady        bool     `json:"docker_ready"`
	SupportedLanguages []string `json:"supported_languages"`
}
