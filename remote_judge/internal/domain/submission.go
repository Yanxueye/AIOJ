package domain

import "time"

// SubmissionStatus 表示提交流程中的状态。
type SubmissionStatus string

const (
	StatusPending             SubmissionStatus = "Pending"
	StatusQueueing            SubmissionStatus = "Queueing"
	StatusCompiling           SubmissionStatus = "Compiling"
	StatusRunning             SubmissionStatus = "Running"
	StatusAccepted            SubmissionStatus = "Accepted"
	StatusWrongAnswer         SubmissionStatus = "Wrong Answer"
	StatusCompileError        SubmissionStatus = "Compile Error"
	StatusRuntimeError        SubmissionStatus = "Runtime Error"
	StatusTimeLimitExceeded   SubmissionStatus = "Time Limit Exceeded"
	StatusMemoryLimitExceeded SubmissionStatus = "Memory Limit Exceeded"
	StatusOutputLimitExceeded SubmissionStatus = "Output Limit Exceeded"
	StatusSystemError         SubmissionStatus = "System Error"
)

// SupportedLanguages 定义当前支持的语言集合。
var SupportedLanguages = map[string]LanguageSpec{
	"cpp17": {
		ID:           "cpp17",
		Label:        "C++17",
		SourceFile:   "main.cpp",
		Compiled:     true,
		Version:      "g++ 11",
		CompileCmd:   []string{"g++", "-O2", "-std=c++17", "main.cpp", "-o", "main"},
		RunCmd:       []string{"./main"},
		DockerImage:  "remote-judge-cpp17",
		BinaryTarget: "main",
	},
	"go1.22": {
		ID:           "go1.22",
		Label:        "Go 1.22",
		SourceFile:   "main.go",
		Compiled:     true,
		Version:      "go1.22",
		CompileCmd:   []string{"sh", "-lc", "GOCACHE=/tmp/go-build HOME=/tmp GOMAXPROCS=1 /usr/local/go/bin/go build -p=1 -o main main.go"},
		RunCmd:       []string{"./main"},
		DockerImage:  "remote-judge-go122",
		BinaryTarget: "main",
	},
	"python3.11": {
		ID:          "python3.11",
		Label:       "Python 3.11",
		SourceFile:  "main.py",
		Compiled:    false,
		Version:     "python3.11",
		RunCmd:      []string{"python3", "main.py"},
		DockerImage: "remote-judge-python311",
	},
}

// LanguageSpec 描述一种提交语言的执行方式。
type LanguageSpec struct {
	ID           string
	Label        string
	SourceFile   string
	Compiled     bool
	Version      string
	CompileCmd   []string
	RunCmd       []string
	DockerImage  string
	BinaryTarget string
}

// IsTerminalStatus 判断状态是否为终态。
func IsTerminalStatus(status SubmissionStatus) bool {
	switch status {
	case StatusAccepted,
		StatusWrongAnswer,
		StatusCompileError,
		StatusRuntimeError,
		StatusTimeLimitExceeded,
		StatusMemoryLimitExceeded,
		StatusOutputLimitExceeded,
		StatusSystemError:
		return true
	default:
		return false
	}
}

// Submission 表示一条代码提交。
type Submission struct {
	ID             int64            `json:"id"`
	UserID         int64            `json:"userId"`
	ProblemID      int64            `json:"problemId"`
	TraceID        string           `json:"traceId"`
	Language       string           `json:"language"`
	Code           string           `json:"-"`
	CodeLength     int              `json:"codeLength"`
	Status         SubmissionStatus `json:"status"`
	RuntimeMs      int              `json:"runtimeMs"`
	MemoryKB       int              `json:"memoryKb"`
	CompileOutput  string           `json:"compileOutput"`
	ErrorMessage   string           `json:"errorMessage"`
	QueueStartedAt *time.Time       `json:"queueStartedAt"`
	JudgeStartedAt *time.Time       `json:"judgeStartedAt"`
	FinishedAt     *time.Time       `json:"finishedAt"`
	CreatedAt      time.Time        `json:"createdAt"`
	UpdatedAt      time.Time        `json:"updatedAt"`
}

// SubmissionCaseResult 表示单个测试点的评测结果。
type SubmissionCaseResult struct {
	SubmissionID  int64            `json:"submissionId"`
	CaseNo        int              `json:"caseNo"`
	Status        SubmissionStatus `json:"status"`
	RuntimeMs     int              `json:"runtimeMs"`
	MemoryKB      int              `json:"memoryKb"`
	StdoutBytes   int              `json:"stdoutBytes"`
	StderrBytes   int              `json:"stderrBytes"`
	Signal        string           `json:"signal"`
	StdoutPreview string           `json:"stdoutPreview"`
	StderrPreview string           `json:"stderrPreview"`
}

// SubmissionFilter 描述提交查询条件。
type SubmissionFilter struct {
	UserID    int64
	Page      int
	PageSize  int
	ProblemID int64
	Status    string
	Language  string
	SortBy    string
}
