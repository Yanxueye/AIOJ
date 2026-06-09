package grpcclient

import (
	"context"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"

	"remote_judge/internal/domain"
	"remote_judge/internal/judger"
	"remote_judge/internal/sandbox"
	grpcserver "remote_judge/internal/transport/grpcserver"
)

// TestClientJudge verifies the remote gRPC client can call a real Judger server.
func TestClientJudge(t *testing.T) {
	t.Logf(">>> gRPC: client -> server (mock) -> Accepted")
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	server := grpc.NewServer()
	grpcserver.Register(server, grpcserver.New(judger.NewService(&sandbox.MockSandbox{})))
	go func() {
		_ = server.Serve(listener)
	}()
	defer server.GracefulStop()

	client, err := New(listener.Addr().String())
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	result, err := client.Judge(ctx, domain.JudgeRequest{
		SubmissionID:  910001,
		ProblemID:     1001,
		Language:      "cpp17",
		Code:          "#include <iostream>\nint main(){int a,b;std::cin>>a>>b;std::cout<<a+b<<\"\\n\";}",
		TimeLimitMs:   1000,
		MemoryLimitMB: 128,
		OutputLimitKB: 1024,
		TestCases: []domain.TestCase{
			{ProblemID: 1001, CaseNo: 1, Input: "1 2\n", Expected: "3\n"},
		},
	})
	if err != nil {
		t.Fatalf("Judge() error = %v", err)
	}
		t.Logf("    status=%s | cases=%d | case#1=%s", result.Status, len(result.CaseResults), result.CaseResults[0].Status)
	if result.Status != domain.StatusAccepted {
		t.Fatalf("unexpected status: %s", result.Status)
	}
	if len(result.CaseResults) != 1 || result.CaseResults[0].Status != domain.StatusAccepted {
		t.Fatalf("unexpected case results: %+v", result.CaseResults)
	}
}

// TestClientHealth verifies the remote gRPC health method is callable.
func TestClientHealth(t *testing.T) {
	t.Logf(">>> gRPC: health check -> ok")
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	server := grpc.NewServer()
	grpcserver.Register(server, grpcserver.New(judger.NewService(&sandbox.MockSandbox{})))
	go func() {
		_ = server.Serve(listener)
	}()
	defer server.GracefulStop()

	client, err := New(listener.Addr().String())
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := client.Health(ctx); err != nil {
		t.Fatalf("Health() error = %v", err)
	}
}
