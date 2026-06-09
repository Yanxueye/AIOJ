package domain

// Problem 描述题目与判题限制。
type Problem struct {
	ID            int64
	Title         string
	TimeLimitMs   int
	MemoryLimitMB int
	OutputLimitKB int
}

// TestCase 描述单个测试用例。
type TestCase struct {
	ProblemID int64
	CaseNo    int
	Input     string
	Expected  string
}
