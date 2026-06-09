package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"remote_judge/internal/config"
	"remote_judge/internal/domain"
	"remote_judge/internal/repository"
	"remote_judge/internal/service"
	"remote_judge/internal/stats"
)

// HTTPServer 提供提交与查询 HTTP API。
type HTTPServer struct {
	submissionService *service.SubmissionService
	queryService      *service.QueryService
	stats             *stats.Collector
	cfg               config.Config
}

// NewHTTPServer 创建一个新的 HTTP API 服务。
func NewHTTPServer(submissionService *service.SubmissionService, queryService *service.QueryService, collector *stats.Collector, cfg config.Config) *HTTPServer {
	return &HTTPServer{
		submissionService: submissionService,
		queryService:      queryService,
		stats:             collector,
		cfg:               cfg,
	}
}

// Routes 注册所有 HTTP 路由。
func (s *HTTPServer) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/submissions", s.handleCreateSubmission)
	mux.HandleFunc("GET /api/submissions", s.handleListSubmissions)
	mux.HandleFunc("GET /api/submissions/", s.handleSubmissionDetail)
	mux.HandleFunc("GET /api/judge/languages", s.handleLanguages)
	mux.HandleFunc("GET /api/system/health", s.handleHealth)
	mux.HandleFunc("GET /api/system/stats", s.handleStats)
	return mux
}

// handleCreateSubmission 处理提交创建请求。
func (s *HTTPServer) handleCreateSubmission(w http.ResponseWriter, r *http.Request) {
	var req service.CreateSubmissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, responseEnvelope{-1, "参数错误", nil})
		return
	}
	req.UserID = currentUserID(r.Context())
	sub, err := s.submissionService.Create(r.Context(), req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrBadRequest):
			writeJSON(w, http.StatusBadRequest, responseEnvelope{-1, "参数错误", nil})
		case errors.Is(err, service.ErrRateLimited):
			writeJSON(w, http.StatusTooManyRequests, responseEnvelope{-1, "提交过于频繁，请稍后再试", nil})
		case errors.Is(err, repository.ErrNotFound):
			writeJSON(w, http.StatusBadRequest, responseEnvelope{-1, "题目不存在", nil})
		default:
			writeJSON(w, http.StatusInternalServerError, responseEnvelope{-1, "创建提交失败", nil})
		}
		return
	}
	writeJSON(w, http.StatusOK, responseEnvelope{0, "ok", sub})
}

// handleListSubmissions 处理提交列表查询。
func (s *HTTPServer) handleListSubmissions(w http.ResponseWriter, r *http.Request) {
	filter := domain.SubmissionFilter{
		UserID:    currentUserID(r.Context()),
		Page:      atoiDefault(r.URL.Query().Get("page"), 1),
		PageSize:  atoiDefault(r.URL.Query().Get("pageSize"), 20),
		ProblemID: int64(atoiDefault(r.URL.Query().Get("problemId"), 0)),
		Status:    r.URL.Query().Get("status"),
		Language:  r.URL.Query().Get("language"),
		SortBy:    r.URL.Query().Get("sortBy"),
	}
	items, total, err := s.queryService.List(r.Context(), filter)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, responseEnvelope{-1, "查询失败", nil})
		return
	}
	writeJSON(w, http.StatusOK, responseEnvelope{0, "ok", map[string]any{
		"list":     items,
		"total":    total,
		"page":     filter.Page,
		"pageSize": filter.PageSize,
	}})
}

// handleSubmissionDetail 分发详情与测试点结果请求。
func (s *HTTPServer) handleSubmissionDetail(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if strings.HasSuffix(path, "/cases") {
		s.handleSubmissionCases(w, r)
		return
	}
	s.handleGetSubmission(w, r)
}

// handleGetSubmission 返回单条提交记录。
func (s *HTTPServer) handleGetSubmission(w http.ResponseWriter, r *http.Request) {
	id, err := parseSubmissionID(r.URL.Path)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, responseEnvelope{-1, "参数错误", nil})
		return
	}
	sub, err := s.queryService.Get(r.Context(), id)
	if err != nil || sub.UserID != currentUserID(r.Context()) {
		writeJSON(w, http.StatusNotFound, responseEnvelope{-1, "提交不存在", nil})
		return
	}
	writeJSON(w, http.StatusOK, responseEnvelope{0, "ok", sub})
}

// handleSubmissionCases 返回测试点结果与汇总。
func (s *HTTPServer) handleSubmissionCases(w http.ResponseWriter, r *http.Request) {
	id, err := parseSubmissionID(pathWithoutSuffix(r.URL.Path, "/cases"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, responseEnvelope{-1, "参数错误", nil})
		return
	}
	sub, err := s.queryService.Get(r.Context(), id)
	if err != nil || sub.UserID != currentUserID(r.Context()) {
		writeJSON(w, http.StatusNotFound, responseEnvelope{-1, "提交不存在", nil})
		return
	}
	cases, err := s.queryService.Cases(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, responseEnvelope{-1, "查询失败", nil})
		return
	}
	passed, failed, skipped := 0, 0, 0
	for _, item := range cases {
		switch item.Status {
		case domain.StatusAccepted:
			passed++
		case domain.StatusPending, domain.StatusQueueing, domain.StatusCompiling, domain.StatusRunning:
			skipped++
		default:
			failed++
		}
	}
	writeJSON(w, http.StatusOK, responseEnvelope{0, "ok", map[string]any{
		"submissionId": id,
		"traceId":      sub.TraceID,
		"summary": map[string]int{
			"total":   len(cases),
			"passed":  passed,
			"failed":  failed,
			"skipped": skipped,
		},
		"cases": cases,
	}})
}

// handleLanguages 返回支持的语言列表。
func (s *HTTPServer) handleLanguages(w http.ResponseWriter, _ *http.Request) {
	list := make([]map[string]any, 0, len(domain.SupportedLanguages))
	for _, lang := range domain.SupportedLanguages {
		list = append(list, map[string]any{
			"id":       lang.ID,
			"label":    lang.Label,
			"enabled":  true,
			"compiled": lang.Compiled,
			"version":  lang.Version,
		})
	}
	writeJSON(w, http.StatusOK, responseEnvelope{0, "ok", map[string]any{"list": list}})
}

// handleHealth 返回服务健康状态。
func (s *HTTPServer) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, responseEnvelope{0, "ok", map[string]any{
		"status":            "SERVING",
		"queue":             s.cfg.QueueMode,
		"sandbox":           "docker",
		"repository":        s.cfg.Repository,
		"judgerMode":        s.cfg.JudgerMode,
		"workerConcurrency": s.cfg.WorkerConcurrency,
	}})
}

// handleStats 返回当前系统指标。
func (s *HTTPServer) handleStats(w http.ResponseWriter, _ *http.Request) {
	snapshot := stats.Snapshot{}
	if s.stats != nil {
		snapshot = s.stats.Snapshot()
	}
	writeJSON(w, http.StatusOK, responseEnvelope{0, "ok", map[string]any{
		"totalSubmissions": snapshot.TotalSubmissions,
		"activeWorkers":    snapshot.ActiveWorkers,
		"maxWorkers":       snapshot.MaxWorkers,
		"busyRejects":      snapshot.BusyRejects,
		"statusCounts":     snapshot.StatusCounts,
	}})
}

type responseEnvelope struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

type ctxKey string

const userIDKey ctxKey = "user_id"

// UserIDCarrier 在请求体中携带可选的演示用户 ID。
type UserIDCarrier struct {
	UserID int64 `json:"userId"`
}

// WithUserID 将当前用户 ID 注入请求上下文。
func WithUserID(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// currentUserID 从上下文中读取当前用户 ID。
func currentUserID(ctx context.Context) int64 {
	v, ok := ctx.Value(userIDKey).(int64)
	if !ok || v <= 0 {
		return 300241
	}
	return v
}

// writeJSON 写入 JSON 响应。
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// atoiDefault 解析整数，解析失败时返回默认值。
func atoiDefault(s string, fallback int) int {
	if s == "" {
		return fallback
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return fallback
	}
	return v
}

// parseSubmissionID 从请求路径中解析提交 ID。
func parseSubmissionID(path string) (int64, error) {
	part := strings.TrimPrefix(path, "/api/submissions/")
	id, err := strconv.ParseInt(part, 10, 64)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// pathWithoutSuffix 从路径中移除指定后缀。
func pathWithoutSuffix(path, suffix string) string {
	return path[:len(path)-len(suffix)]
}
