package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"remote_judge/internal/config"
	"remote_judge/internal/queue"
	"remote_judge/internal/repository"
	"remote_judge/internal/service"
	"remote_judge/internal/stats"
)

// TestHTTPServerCreateAndQuery verifies submission and system endpoints.
func TestHTTPServerCreateAndQuery(t *testing.T) {
	t.Logf(">>> HTTP API: create -> list -> health -> stats")
	subRepo := repository.NewInMemorySubmissionRepository()
	problemRepo := repository.NewInMemoryProblemRepository()
	q := queue.NewMemoryQueue(16)
	collector := stats.NewCollector()
	collector.SetMaxWorkers(4)
	subSvc := service.NewSubmissionService(subRepo, problemRepo, q, collector)
	querySvc := service.NewQueryService(subRepo)
	server := NewHTTPServer(subSvc, querySvc, collector, config.Config{
		QueueMode:         "memory",
		Repository:        "memory",
		JudgerMode:        "embedded",
		WorkerConcurrency: 4,
	})

	handler := FakeAuthMiddleware(server.Routes())

	req := httptest.NewRequest(http.MethodPost, "/api/submissions", strings.NewReader(`{"problemId":1001,"language":"cpp17","code":"int main(){}","userId":1}`))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	if resp.Code != http.StatusOK {
		t.Fatalf("POST /api/submissions code = %d, body = %s", resp.Code, resp.Body.String())
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/submissions?page=1&pageSize=10", nil)
	listResp := httptest.NewRecorder()
	handler.ServeHTTP(listResp, listReq)
	if listResp.Code != http.StatusOK {
		t.Fatalf("GET /api/submissions code = %d", listResp.Code)
	}

	healthReq := httptest.NewRequest(http.MethodGet, "/api/system/health", nil)
	healthResp := httptest.NewRecorder()
	handler.ServeHTTP(healthResp, healthReq)
	if healthResp.Code != http.StatusOK {
		t.Fatalf("GET /api/system/health code = %d", healthResp.Code)
	}

	statsReq := httptest.NewRequest(http.MethodGet, "/api/system/stats", nil)
	statsResp := httptest.NewRecorder()
	handler.ServeHTTP(statsResp, statsReq)
	if statsResp.Code != http.StatusOK {
		t.Fatalf("GET /api/system/stats code = %d", statsResp.Code)
	}

	var body struct {
		Code int `json:"code"`
		Data struct {
			TraceID string `json:"traceId"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal create response failed: %v", err)
	}
		t.Logf("    create: traceId=%s", body.Data.TraceID)
	if body.Data.TraceID == "" {
		t.Fatal("expected trace id in create response")
	}

	var statsBody struct {
		Code int `json:"code"`
		Data struct {
			TotalSubmissions int `json:"totalSubmissions"`
			MaxWorkers       int `json:"maxWorkers"`
		} `json:"data"`
	}
	if err := json.Unmarshal(statsResp.Body.Bytes(), &statsBody); err != nil {
		t.Fatalf("unmarshal stats response failed: %v", err)
	}
	if statsBody.Data.TotalSubmissions != 1 {
		t.Fatalf("unexpected total submissions: %d", statsBody.Data.TotalSubmissions)
	}
		t.Logf("    totalSubmissions=%d | maxWorkers=%d", statsBody.Data.TotalSubmissions, statsBody.Data.MaxWorkers)
	if statsBody.Data.MaxWorkers != 4 {
		t.Fatalf("unexpected max workers: %d", statsBody.Data.MaxWorkers)
	}
}

// TestHTTPServerRejectsBlankCode verifies request validation surfaces as 400.
func TestHTTPServerRejectsBlankCode(t *testing.T) {
	t.Logf(">>> HTTP API: blank code -> 400 Bad Request")
	subRepo := repository.NewInMemorySubmissionRepository()
	problemRepo := repository.NewInMemoryProblemRepository()
	q := queue.NewMemoryQueue(16)
	collector := stats.NewCollector()
	subSvc := service.NewSubmissionService(subRepo, problemRepo, q, collector)
	querySvc := service.NewQueryService(subRepo)
	server := NewHTTPServer(subSvc, querySvc, collector, config.Config{})

	handler := FakeAuthMiddleware(server.Routes())
	req := httptest.NewRequest(http.MethodPost, "/api/submissions", strings.NewReader(`{"problemId":1001,"language":"cpp17","code":"   ","userId":1}`))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d, body=%s", resp.Code, resp.Body.String())
	}
}
