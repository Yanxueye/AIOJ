package domain

// JudgeRequest 表示一次内部判题请求。
type JudgeRequest struct {
	SubmissionID  int64
	ProblemID     int64
	TraceID       string
	Language      string
	Code          string
	TimeLimitMs   int
	MemoryLimitMB int
	OutputLimitKB int
	TestCases     []TestCase
}

// JudgeResult 表示一次内部判题响应。
type JudgeResult struct {
	SubmissionID int64
	Status       SubmissionStatus
	RuntimeMs    int
	MemoryKB     int
	CompileOut   string
	ErrorMessage string
	CaseResults  []SubmissionCaseResult
}
