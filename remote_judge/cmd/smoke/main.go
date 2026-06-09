package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"time"
)

// submitRequest 表示 smoke 测试提交体。
type submitRequest struct {
	ProblemID int64  `json:"problemId"`
	Language  string `json:"language"`
	Code      string `json:"code"`
	UserID    int64  `json:"userId"`
}

// submitResponse 表示提交接口响应。
type submitResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		ID int64 `json:"id"`
	} `json:"data"`
}

// detailResponse 表示提交详情响应。
type detailResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		ID           int64  `json:"id"`
		Status       string `json:"status"`
		RuntimeMs    int    `json:"runtimeMs"`
		MemoryKB     int    `json:"memoryKb"`
		CompileOut   string `json:"compileOutput"`
		ErrorMessage string `json:"errorMessage"`
	} `json:"data"`
}

// main 模拟外部系统提交真实代码并轮询判题结果。
func main() {
	var (
		baseURL  = flag.String("addr", "http://127.0.0.1:8080", "remote_judge base URL")
		language = flag.String("lang", "cpp17", "submission language")
		problem  = flag.Int64("problem", 1001, "problem id")
		userID   = flag.Int64("user", 400001, "user id")
		timeout  = flag.Duration("timeout", 10*time.Second, "poll timeout")
		mode     = flag.String("mode", "ac", "sample mode: ac|wa|ce|py|ole")
	)
	flag.Parse()

	code := sampleCode(*mode, *language)
	reqBody, _ := json.Marshal(submitRequest{
		ProblemID: *problem,
		Language:  *language,
		Code:      code,
		UserID:    *userID,
	})

	client := &http.Client{Timeout: 5 * time.Second}
	submitResp, err := client.Post(*baseURL+"/api/submissions", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		panic(err)
	}
	defer submitResp.Body.Close()

	body, _ := io.ReadAll(submitResp.Body)
	var created submitResponse
	if err := json.Unmarshal(body, &created); err != nil {
		panic(err)
	}
	if created.Code != 0 {
		panic(fmt.Sprintf("submit failed: %s", string(body)))
	}

	deadline := time.Now().Add(*timeout)
	lastStatus := ""
	for time.Now().Before(deadline) {
		time.Sleep(300 * time.Millisecond)
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/submissions/%d?userId=%d", *baseURL, created.Data.ID, *userID), nil)
		req.Header.Set("X-Demo-User-ID", fmt.Sprintf("%d", *userID))
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		payload, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()

		var detail detailResponse
		if json.Unmarshal(payload, &detail) != nil {
			continue
		}
		if detail.Data.Status != lastStatus {
			fmt.Printf("submission=%d status=%s runtime=%dms memory=%dKB\n", detail.Data.ID, detail.Data.Status, detail.Data.RuntimeMs, detail.Data.MemoryKB)
			lastStatus = detail.Data.Status
		}
		if detail.Data.CompileOut != "" {
			fmt.Printf("compileOutput: %s\n", detail.Data.CompileOut)
		}
		if detail.Data.ErrorMessage != "" {
			fmt.Printf("errorMessage: %s\n", detail.Data.ErrorMessage)
		}
		if isTerminal(detail.Data.Status) {
			return
		}
	}

	panic("poll timeout")
}

// sampleCode 返回一段真实可评测代码。
func sampleCode(mode, language string) string {
	switch language {
	case "python3.11":
		if mode == "ole" {
			return "print('x' * 4096)\n"
		}
		if mode == "wa" {
			return "a,b=map(int,input().split())\nprint(a-b)\n"
		}
		return "a,b=map(int,input().split())\nprint(a+b)\n"
	default:
		switch mode {
		case "wa":
			return "#include <iostream>\nint main(){int a,b;std::cin>>a>>b;std::cout<<a-b<<\"\\n\";}"
		case "ce":
			return "#include <iostream>\nint main( { return 0; }"
		case "ole":
			return "#include <iostream>\nint main(){for(int i=0;i<5000;i++)std::cout<<'x';}"
		default:
			return "#include <iostream>\nint main(){int a,b;std::cin>>a>>b;std::cout<<a+b<<\"\\n\";}"
		}
	}
}

// isTerminal 判断是否已经进入终态。
func isTerminal(status string) bool {
	switch status {
	case "Accepted", "Wrong Answer", "Compile Error", "Runtime Error", "Time Limit Exceeded", "Memory Limit Exceeded", "Output Limit Exceeded", "System Error":
		return true
	default:
		return false
	}
}
