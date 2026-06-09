package judger

import (
	"context"
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// languageMap translates AIOJ language names to remote_judge language names.
var languageMap = map[string]string{
	"cpp":    "cpp17",
	"python": "python3.11",
	"go":     "go1.22",
}

// rjJudgeRequest mirrors remote_judge's pb.JudgeRequest for wire compatibility.
type rjJudgeRequest struct {
	SubmissionID  int64        `json:"submission_id"`
	ProblemID     int64        `json:"problem_id"`
	TraceID       string       `json:"trace_id,omitempty"`
	Language      string       `json:"language"`
	Code          string       `json:"code"`
	TimeLimitMs   int32        `json:"time_limit_ms"`
	MemoryLimitMB int32        `json:"memory_limit_mb"`
	OutputLimitKB int32        `json:"output_limit_kb"`
	RunMode       string       `json:"run_mode,omitempty"`
	TestCases     []rjTestCase `json:"test_cases"`
}

type rjTestCase struct {
	CaseNo   int32  `json:"case_no"`
	Input    string `json:"input,omitempty"`
	Expected string `json:"expected,omitempty"`
}

type rjJudgeResponse struct {
	SubmissionID int64          `json:"submission_id"`
	Status       string         `json:"status"`
	RuntimeMs    int32          `json:"runtime_ms"`
	MemoryKB     int32          `json:"memory_kb"`
	CompileOut   string         `json:"compile_output"`
	ErrorMessage string         `json:"error_message"`
	CaseResults  []rjCaseResult `json:"case_results"`
}

type rjCaseResult struct {
	CaseNo        int32  `json:"case_no"`
	Status        string `json:"status"`
	RuntimeMs     int32  `json:"runtime_ms"`
	MemoryKB      int32  `json:"memory_kb"`
	StdoutBytes   int32  `json:"stdout_bytes"`
	StderrBytes   int32  `json:"stderr_bytes"`
	Signal        string `json:"signal"`
	StdoutPreview string `json:"stdout_preview"`
	StderrPreview string `json:"stderr_preview"`
}

// RemoteJudger translates AIOJ's JudgeRequest to remote_judge's gRPC protocol
// and back. It speaks the same gRPC method path (/judger.Judger/Judge) as the
// built-in MockSandbox, but adapts the wire format and language naming.
type RemoteJudger struct {
	addr    string
	timeout time.Duration

	mu   sync.Mutex
	conn *grpc.ClientConn
}

// NewRemoteJudger creates a RemoteJudger that dials the given gRPC address.
func NewRemoteJudger(addr string, timeoutSeconds int) *RemoteJudger {
	if timeoutSeconds <= 0 {
		timeoutSeconds = 15
	}
	return &RemoteJudger{addr: addr, timeout: time.Duration(timeoutSeconds) * time.Second}
}

func (r *RemoteJudger) ensureConn() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.conn != nil {
		return nil
	}
	conn, err := grpc.NewClient(r.addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.CallContentSubtype(codecName)),
	)
	if err != nil {
		return fmt.Errorf("grpc dial %s: %w", r.addr, err)
	}
	r.conn = conn
	return nil
}

// Close releases the underlying gRPC connection.
func (r *RemoteJudger) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.conn == nil {
		return nil
	}
	err := r.conn.Close()
	r.conn = nil
	return err
}

// Judge translates an AIOJ JudgeRequest to remote_judge's protocol, calls the
// remote gRPC judger, and translates the response back.
func (r *RemoteJudger) Judge(ctx context.Context, req *JudgeRequest) (*JudgeResponse, error) {
	if err := r.ensureConn(); err != nil {
		return nil, err
	}

	// Translate language name (e.g. "cpp" -> "cpp17").
	rjLang := req.Language
	if mapped, ok := languageMap[req.Language]; ok {
		rjLang = mapped
	}

	// Build remote_judge-compatible request.
	rjReq := &rjJudgeRequest{
		SubmissionID:  int64(req.SubmissionID),
		ProblemID:     int64(req.ProblemID),
		TraceID:       req.TraceID,
		Language:      rjLang,
		Code:          req.Code,
		TimeLimitMs:   req.TimeLimitMS,
		MemoryLimitMB: req.MemoryLimitMB,
		OutputLimitKB: req.OutputLimitKB,
		RunMode:       req.RunMode,
		TestCases:     make([]rjTestCase, len(req.TestCases)),
	}
	for i, tc := range req.TestCases {
		rjReq.TestCases[i] = rjTestCase{
			CaseNo:   maxCaseNo(tc.CaseNo, int32(i+1)),
			Input:    tc.Input,
			Expected: tc.Expected,
		}
	}

	// Call remote_judge gRPC server.
	callCtx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	rjResp := new(rjJudgeResponse)
	if err := r.conn.Invoke(callCtx, MethodJudge, rjReq, rjResp); err != nil {
		return nil, err
	}

	// Translate response back to AIOJ format (MemoryKB -> MemoryMB string).
	memoryMB := fmt.Sprintf("%.1f", float64(rjResp.MemoryKB)/1024.0)
	caseResults := make([]CaseResult, len(rjResp.CaseResults))
	for i, item := range rjResp.CaseResults {
		caseResults[i] = CaseResult{
			CaseNo:        item.CaseNo,
			Status:        item.Status,
			RuntimeMS:     item.RuntimeMs,
			MemoryKB:      item.MemoryKB,
			StdoutBytes:   item.StdoutBytes,
			StderrBytes:   item.StderrBytes,
			Signal:        item.Signal,
			StdoutPreview: item.StdoutPreview,
			StderrPreview: item.StderrPreview,
		}
	}

	return &JudgeResponse{
		SubmissionID: uint64(rjResp.SubmissionID),
		Status:       rjResp.Status,
		RuntimeMS:    rjResp.RuntimeMs,
		MemoryMB:     memoryMB,
		MemoryKB:     rjResp.MemoryKB,
		CompileOut:   rjResp.CompileOut,
		ErrorMessage: rjResp.ErrorMessage,
		CaseResults:  caseResults,
	}, nil
}

func maxCaseNo(value, fallback int32) int32 {
	if value > 0 {
		return value
	}
	return fallback
}
