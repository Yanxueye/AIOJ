package judge

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

type SubmitRequest struct {
	Language string `json:"language"`
	Code     string `json:"code"`
}

type SubmitResponse struct {
	ID     uint64 `json:"id"`
	Status string `json:"status"`
}

type SubmissionResult struct {
	ID          uint64       `json:"id"`
	Status      string       `json:"status"`
	RuntimeMS   int          `json:"runtimeMs"`
	MemoryKB    int          `json:"memoryKb"`
	CompileOut  string       `json:"compileOutput"`
	ErrorMsg    string       `json:"errorMessage"`
	CaseResults []CaseResult `json:"caseResults"`
}

type CaseResult struct {
	CaseNo    int    `json:"caseNo"`
	Status    string `json:"status"`
	RuntimeMS int    `json:"runtimeMs"`
	MemoryKB  int    `json:"memoryKb"`
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// Submit submits code for judging via the async submission endpoint and returns the submission ID
func (c *Client) Submit(problemID uint64, lang, code string) (uint64, error) {
	req := map[string]interface{}{
		"problemId": problemID,
		"language":  lang,
		"code":      code,
	}
	body, err := json.Marshal(req)
	if err != nil {
		return 0, err
	}

	url := fmt.Sprintf("%s/api/submissions", c.baseURL)
	resp, err := c.httpClient.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return 0, fmt.Errorf("judge submit failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("judge returned %d: %s", resp.StatusCode, string(respBody))
	}

	var envelope struct {
		Code int              `json:"code"`
		Data *SubmissionResult `json:"data"`
	}
	if err := json.Unmarshal(respBody, &envelope); err != nil {
		return 0, err
	}
	if envelope.Data == nil {
		return 0, fmt.Errorf("no data in response")
	}
	return envelope.Data.ID, nil
}

// GetResult fetches the result of a submission
func (c *Client) GetResult(submissionID uint64) (*SubmissionResult, error) {
	url := fmt.Sprintf("%s/api/submissions/%d", c.baseURL, submissionID)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var envelope struct {
		Code int              `json:"code"`
		Data *SubmissionResult `json:"data"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, err
	}
	if envelope.Data == nil {
		return nil, fmt.Errorf("no data in response")
	}
	return envelope.Data, nil
}

// RunCode runs code with custom input (synchronous)
func (c *Client) RunCode(lang, code, stdin string) (*SubmissionResult, error) {
	payload := map[string]string{
		"language": lang,
		"code":     code,
		"stdin":    stdin,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/api/problems/0/run", c.baseURL)
	resp, err := c.httpClient.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("run code failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result SubmissionResult
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
