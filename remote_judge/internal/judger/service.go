package judger

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"remote_judge/internal/domain"
	"remote_judge/internal/logger"
	"remote_judge/internal/sandbox"
)

// Service 提供判题核心逻辑。
type Service struct {
	sandbox sandbox.Sandbox
	pool    *WorkspacePool
}

// NewService 创建一个 Judger 服务。
func NewService(sb sandbox.Sandbox) *Service {
	return &Service{sandbox: sb}
}

// WithWorkspacePool 为服务附加一个工作区对象池。
func (s *Service) WithWorkspacePool(pool *WorkspacePool) *Service {
	s.pool = pool
	return s
}

// Judge 执行一次完整的编译与判题流程。
func (s *Service) Judge(ctx context.Context, req domain.JudgeRequest) (domain.JudgeResult, error) {
	spec, ok := domain.SupportedLanguages[req.Language]
	if !ok {
		return domain.JudgeResult{
			SubmissionID: req.SubmissionID,
			Status:       domain.StatusSystemError,
			ErrorMessage: "language not supported",
		}, nil
	}

	if valid, reason := ValidateCode(req.Code, req.Language); !valid {
		return domain.JudgeResult{
			SubmissionID: req.SubmissionID,
			Status:       domain.StatusCompileError,
			CompileOut:   reason,
			ErrorMessage: "forbidden code pattern",
		}, nil
	}

	workDir, err := s.prepareWorkspace(req, spec.SourceFile)
	if err != nil {
		return domain.JudgeResult{SubmissionID: req.SubmissionID, Status: domain.StatusSystemError, ErrorMessage: err.Error()}, err
	}
	if s.pool != nil {
		defer s.pool.Release(workDir)
	} else {
		defer os.RemoveAll(workDir)
	}

	result := domain.JudgeResult{SubmissionID: req.SubmissionID}

	if spec.Compiled {
		compileTimeout := 15 * time.Second
		if req.TimeLimitMs > 0 {
			scaled := time.Duration(req.TimeLimitMs*5) * time.Millisecond
			if scaled > compileTimeout {
				compileTimeout = scaled
			}
		}
		compileReq := sandbox.ExecRequest{
			Language:      req.Language,
			Image:         spec.DockerImage,
			WorkDir:       workDir,
			Command:       spec.CompileCmd,
			TimeLimit:     compileTimeout,
			MemoryLimitMB: max(req.MemoryLimitMB, 128),
			OutputLimitKB: req.OutputLimitKB,
		}
		compileRes, compileErr := s.sandbox.Compile(ctx, compileReq)
		if compileRes.ExitCode != 0 {
			logger.Error("judge.compile", req.TraceID, "compile failed", map[string]any{
				"submissionId": req.SubmissionID,
				"language":     req.Language,
				"exitCode":     compileRes.ExitCode,
				"stderrBytes":  compileRes.StderrBytes,
			})
			result.Status = domain.StatusCompileError
			result.CompileOut = trimPreview(nonEmpty(compileRes.Stderr, compileRes.Stdout, compileErrString(compileErr)), 500)
			result.ErrorMessage = "compile failed"
			return result, nil
		}
		if compileErr != nil {
			result.Status = domain.StatusSystemError
			result.ErrorMessage = compileErr.Error()
			return result, nil
		}
	}

	finalStatus := domain.StatusAccepted
	maxRuntime := 0
	maxMemory := 0
	caseResults := make([]domain.SubmissionCaseResult, 0, len(req.TestCases))

	for _, tc := range req.TestCases {
		runReq := sandbox.ExecRequest{
			Language:      req.Language,
			Image:         spec.DockerImage,
			WorkDir:       workDir,
			Command:       spec.RunCmd,
			Stdin:         tc.Input,
			TimeLimit:     time.Duration(req.TimeLimitMs) * time.Millisecond,
			MemoryLimitMB: req.MemoryLimitMB,
			OutputLimitKB: req.OutputLimitKB,
		}
		runRes, runErr := s.sandbox.Run(ctx, runReq)
		if runErr != nil && runRes.ExitCode == 0 && !runRes.TimedOut {
			finalStatus = domain.StatusSystemError
			caseResults = append(caseResults, domain.SubmissionCaseResult{
				SubmissionID:  req.SubmissionID,
				CaseNo:        tc.CaseNo,
				Status:        domain.StatusSystemError,
				StderrPreview: trimPreview(runErr.Error(), 200),
			})
			break
		}

		caseStatus := s.compareCase(runRes, tc.Expected, req)
		caseResults = append(caseResults, domain.SubmissionCaseResult{
			SubmissionID:  req.SubmissionID,
			CaseNo:        tc.CaseNo,
			Status:        caseStatus,
			RuntimeMs:     int(runRes.Runtime.Milliseconds()),
			MemoryKB:      runRes.MemoryKB,
			StdoutBytes:   runRes.StdoutBytes,
			StderrBytes:   runRes.StderrBytes,
			Signal:        runRes.Signal,
			StdoutPreview: trimPreview(runRes.Stdout, 200),
			StderrPreview: trimPreview(runRes.Stderr, 200),
		})
		logger.Info("judge.case", req.TraceID, "judge case finished", map[string]any{
			"submissionId": req.SubmissionID,
			"caseNo":       tc.CaseNo,
			"status":       caseStatus,
			"runtimeMs":    int(runRes.Runtime.Milliseconds()),
			"memoryKB":     runRes.MemoryKB,
			"stdoutBytes":  runRes.StdoutBytes,
			"stderrBytes":  runRes.StderrBytes,
			"signal":       runRes.Signal,
		})

		if int(runRes.Runtime.Milliseconds()) > maxRuntime {
			maxRuntime = int(runRes.Runtime.Milliseconds())
		}
		if runRes.MemoryKB > maxMemory {
			maxMemory = runRes.MemoryKB
		}

		if caseStatus != domain.StatusAccepted {
			finalStatus = caseStatus
			break
		}
	}

	if len(caseResults) < len(req.TestCases) && finalStatus != domain.StatusAccepted {
		for _, tc := range req.TestCases[len(caseResults):] {
			caseResults = append(caseResults, domain.SubmissionCaseResult{
				SubmissionID: req.SubmissionID,
				CaseNo:       tc.CaseNo,
				Status:       finalStatus,
			})
		}
	}

	result.Status = finalStatus
	result.RuntimeMs = maxRuntime
	result.MemoryKB = maxMemory
	result.CaseResults = caseResults
	if finalStatus == domain.StatusWrongAnswer {
		result.ErrorMessage = "answer mismatch"
	}
	return result, nil
}

// compileErrString 提取编译错误文本。
func compileErrString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// nonEmpty 返回第一个非空字符串。
func nonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

// Health 检查底层沙箱是否健康。
func (s *Service) Health(ctx context.Context) error {
	return s.sandbox.Health(ctx)
}

// prepareWorkspace 为本次判题创建临时目录与源文件。
func (s *Service) prepareWorkspace(req domain.JudgeRequest, sourceFile string) (string, error) {
	var (
		workDir string
		err     error
	)
	if s.pool != nil {
		workDir, err = s.pool.Acquire()
	} else {
		workDir, err = os.MkdirTemp("", fmt.Sprintf("submission_%d_", req.SubmissionID))
	}
	if err != nil {
		return "", err
	}
	codePath := filepath.Join(workDir, sourceFile)
	if err := os.WriteFile(codePath, []byte(req.Code), 0o644); err != nil {
		return "", err
	}
	return workDir, nil
}

// compareCase 比较一次运行结果与预期输出。
func (s *Service) compareCase(res sandbox.ExecResult, expected string, req domain.JudgeRequest) domain.SubmissionStatus {
	if res.TimedOut || int(res.Runtime.Milliseconds()) > req.TimeLimitMs {
		return domain.StatusTimeLimitExceeded
	}
	if res.OOMKilled || res.ExitCode == 137 {
		return domain.StatusMemoryLimitExceeded
	}
	if req.MemoryLimitMB > 0 && res.MemoryKB > req.MemoryLimitMB*1024 {
		return domain.StatusMemoryLimitExceeded
	}
	if req.OutputLimitKB > 0 && len(res.Stdout) > req.OutputLimitKB*1024 {
		return domain.StatusOutputLimitExceeded
	}
	if res.ExitCode != 0 {
		return domain.StatusRuntimeError
	}
	if !compareOutput(res.Stdout, expected) {
		return domain.StatusWrongAnswer
	}
	return domain.StatusAccepted
}

// normalizeOutput 统一输出比较规则。归一化 CRLF、去除末尾空白。
func normalizeOutput(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.TrimRight(s, " \t\n\r")
}

// normalizeOutputStrict 合并连续空白并 trim 两端。
func normalizeOutputStrict(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.TrimSpace(s)
	// 合并连续的空格和制表符。
	var b strings.Builder
	inSpace := false
	for _, r := range s {
		if r == ' ' || r == '\t' {
			if !inSpace {
				b.WriteByte(' ')
				inSpace = true
			}
		} else {
			b.WriteRune(r)
			inSpace = false
		}
	}
	return b.String()
}

// compareOutput 检查输出相等性，支持逐行浮点容忍。
func compareOutput(actual, expected string) bool {
	if normalizeOutput(actual) == normalizeOutput(expected) {
		return true
	}
	if normalizeOutputStrict(actual) == normalizeOutputStrict(expected) {
		return true
	}
	actualLines := strings.Split(normalizeOutput(actual), "\n")
	expectedLines := strings.Split(normalizeOutput(expected), "\n")
	if len(actualLines) != len(expectedLines) {
		return false
	}
	for i := range actualLines {
		if !lineMatch(strings.TrimSpace(actualLines[i]), strings.TrimSpace(expectedLines[i])) {
			return false
		}
	}
	return true
}

// lineMatch 检查单行输出与预期是否匹配，支持浮点容忍。
func lineMatch(actual, expected string) bool {
	if actual == expected {
		return true
	}
	aVal, aErr := strconvParseFloat(actual)
	eVal, eErr := strconvParseFloat(expected)
	if aErr != nil || eErr != nil {
		return false
	}
	diff := aVal - eVal
	if diff < 0 {
		diff = -diff
	}
	return diff < 1e-6
}

func strconvParseFloat(s string) (float64, error) {
	n := len(s)
	i := 0
	for i < n && (s[i] == ' ' || s[i] == '\t') {
		i++
	}
	if i >= n {
		return 0, errors.New("empty")
	}
	neg := false
	if s[i] == '+' || s[i] == '-' {
		neg = s[i] == '-'
		i++
	}
	var intPart int64
	for i < n && s[i] >= '0' && s[i] <= '9' {
		intPart = intPart*10 + int64(s[i]-'0')
		i++
	}
	fracPart := 0.0
	if i < n && s[i] == '.' {
		i++
		pow := 0.1
		for i < n && s[i] >= '0' && s[i] <= '9' {
			fracPart += float64(s[i]-'0') * pow
			pow *= 0.1
			i++
		}
	}
	if i == 0 || (intPart == 0 && fracPart == 0 && i == 0) {
		return 0, errors.New("no digits")
	}
	result := float64(intPart) + fracPart
	if neg {
		result = -result
	}
	return result, nil
}

// trimPreview 截断长文本以便展示。
func trimPreview(s string, limit int) string {
	if len(s) <= limit {
		return s
	}
	return s[:limit]
}

// max 返回两个整数中的较大值。
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
