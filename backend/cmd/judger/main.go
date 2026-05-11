package main

import (
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/terminaloj/backend/internal/judger"
	"google.golang.org/grpc"
)

// This binary is intended to run inside a Docker container isolated from
// the main API. See docker/Dockerfile.judger for the reference image.
//
// In this MVP the sandbox is mocked (judger.MockSandbox); the real image
// would shell out to `isolate` / `nsjail` / `runc` after compiling the
// submitted source.
func main() {
	addr := flag.String("addr", "0.0.0.0:9090", "gRPC listen address")
	flag.Parse()

	lis, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("listen %s: %v", *addr, err)
	}

	s := grpc.NewServer()
	judger.Register(s, judger.MockSandbox{})
	log.Printf("[judger] gRPC listening on %s", *addr)

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("serve: %v", err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Println("[judger] shutting down...")
	s.GracefulStop()
}
