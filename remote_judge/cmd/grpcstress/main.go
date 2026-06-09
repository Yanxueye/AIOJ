package main

import (
	"context"
	"flag"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"remote_judge/internal/domain"
	grpcclient "remote_judge/internal/transport/grpcclient"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:9090", "gRPC service address")
	total := flag.Int("n", 100, "total request count")
	concurrency := flag.Int("c", 10, "concurrency")
	lang := flag.String("lang", "mixed", "cpp17|go1.22|python3.11|mixed")
	flag.Parse()

	client, err := grpcclient.New(*addr)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	var success int64
	var failure int64
	var firstFailure atomic.Value
	var latMu sync.Mutex
	var latencies []time.Duration
	var compileLatencies []time.Duration
	var runLatencies []time.Duration

	jobs := make(chan int)
	start := time.Now()
	var wg sync.WaitGroup
	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for job := range jobs {
				begin := time.Now()
				req := requestFor(job, *lang)
				ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
				resp, err := client.Judge(ctx, req)
				cancel()
				lat := time.Since(begin)

				latMu.Lock()
				latencies = append(latencies, lat)
				latMu.Unlock()

				if err != nil {
					atomic.AddInt64(&failure, 1)
					firstFailure.CompareAndSwap(nil, "transport: "+err.Error())
					continue
				}
				if resp.Status == "" {
					atomic.AddInt64(&failure, 1)
					firstFailure.CompareAndSwap(nil, "empty status")
					continue
				}
				if resp.Status == domain.StatusSystemError {
					atomic.AddInt64(&failure, 1)
					firstFailure.CompareAndSwap(nil, "business: "+resp.ErrorMessage)
					continue
				}
				latMu.Lock()
				if len(resp.CaseResults) == 0 {
					compileLatencies = append(compileLatencies, lat)
				} else {
					runLatencies = append(runLatencies, lat)
				}
				latMu.Unlock()
				atomic.AddInt64(&success, 1)
			}
		}(i)
	}
	for i := 0; i < *total; i++ {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	elapsed := time.Since(start)

	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
	fmt.Printf("total=%d success=%d failure=%d elapsed=%s qps=%.2f p50=%s p90=%s p99=%s\n",
		*total,
		success,
		failure,
		elapsed,
		float64(*total)/elapsed.Seconds(),
		percentile(latencies, 50),
		percentile(latencies, 90),
		percentile(latencies, 99),
	)
	if v := firstFailure.Load(); v != nil {
		fmt.Printf("first_failure=%s\n", v.(string))
	}
	if len(compileLatencies) > 0 {
		sort.Slice(compileLatencies, func(i, j int) bool { return compileLatencies[i] < compileLatencies[j] })
		fmt.Printf("compile_p50=%s compile_p90=%s compile_p99=%s\n",
			percentile(compileLatencies, 50),
			percentile(compileLatencies, 90),
			percentile(compileLatencies, 99),
		)
	}
	if len(runLatencies) > 0 {
		sort.Slice(runLatencies, func(i, j int) bool { return runLatencies[i] < runLatencies[j] })
		fmt.Printf("run_p50=%s run_p90=%s run_p99=%s\n",
			percentile(runLatencies, 50),
			percentile(runLatencies, 90),
			percentile(runLatencies, 99),
		)
	}
}

func percentile(values []time.Duration, p int) time.Duration {
	if len(values) == 0 {
		return 0
	}
	index := (len(values) - 1) * p / 100
	return values[index]
}

func requestFor(i int, mode string) domain.JudgeRequest {
	lang := mode
	if mode == "mixed" {
		switch i % 3 {
		case 0:
			lang = "cpp17"
		case 1:
			lang = "go1.22"
		default:
			lang = "python3.11"
		}
	}
	code := map[string]string{
		"cpp17":      "#include <iostream>\nint main(){int a,b;std::cin>>a>>b;std::cout<<a+b<<\"\\n\";}",
		"go1.22":     "package main\nimport \"fmt\"\nfunc main(){var a,b int; fmt.Scan(&a,&b); fmt.Println(a+b)}",
		"python3.11": "a,b=map(int,input().split())\nprint(a+b)",
	}[lang]
	return domain.JudgeRequest{
		SubmissionID:  int64(1000000 + i),
		ProblemID:     1001,
		TraceID:       fmt.Sprintf("grpc-stress-%d", i),
		Language:      lang,
		Code:          code,
		TimeLimitMs:   2000,
		MemoryLimitMB: 256,
		OutputLimitKB: 1024,
		TestCases: []domain.TestCase{
			{ProblemID: 1001, CaseNo: 1, Input: "1 2\n", Expected: "3\n"},
		},
	}
}
