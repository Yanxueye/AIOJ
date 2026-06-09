package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// submissionPayload 定义压测提交体。
type submissionPayload struct {
	ProblemID int64  `json:"problemId"`
	Language  string `json:"language"`
	Code      string `json:"code"`
	UserID    int64  `json:"userId,omitempty"`
}

// main 对 HTTP 提交接口做简单并发压测。
func main() {
	var (
		total       = flag.Int("n", 100, "total requests")
		concurrency = flag.Int("c", 10, "concurrency")
		addr        = flag.String("addr", "http://127.0.0.1:8080/api/submissions", "target address")
	)
	flag.Parse()

	payload, _ := json.Marshal(submissionPayload{
		ProblemID: 1001,
		Language:  "cpp17",
		Code:      "#include <iostream>\nint main(){int a,b;std::cin>>a>>b;std::cout<<a+b<<\"\\n\";}",
	})

	client := &http.Client{Timeout: 5 * time.Second}
	start := time.Now()
	var success int64
	var wg sync.WaitGroup
	sem := make(chan struct{}, *concurrency)

	for i := 0; i < *total; i++ {
		wg.Add(1)
		sem <- struct{}{}
		go func(i int) {
			defer wg.Done()
			defer func() { <-sem }()
			var body []byte
			if *total > 12 {
				dynamicPayload, _ := json.Marshal(submissionPayload{
					ProblemID: 1001,
					Language:  "cpp17",
					Code:      "#include <iostream>\nint main(){int a,b;std::cin>>a>>b;std::cout<<a+b<<\"\\n\";}",
					UserID:    int64(i + 1),
				})
				body = dynamicPayload
			} else {
				body = payload
			}
			resp, err := client.Post(*addr, "application/json", bytes.NewReader(body))
			if err == nil && resp.StatusCode == http.StatusOK {
				atomic.AddInt64(&success, 1)
			}
			if resp != nil {
				_ = resp.Body.Close()
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)
	fmt.Printf("total=%d success=%d duration=%s qps=%.2f\n", *total, success, duration, float64(*total)/duration.Seconds())
}
