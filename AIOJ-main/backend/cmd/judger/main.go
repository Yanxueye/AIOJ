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
// When --remote is provided, this process acts as a gRPC proxy that forwards
// Judge requests to remote_judge's real Docker-based sandbox. Without the
// flag, it runs the built-in MockSandbox (deterministic pseudo-judger).
func main() {
	addr := flag.String("addr", "0.0.0.0:9090", "gRPC listen address")
	remoteAddr := flag.String("remote", "", "remote_judge gRPC address (proxy mode, e.g. 127.0.0.1:9090)")
	flag.Parse()

	var handler judger.Handler
	if *remoteAddr != "" {
		handler = judger.NewRemoteJudger(*remoteAddr, 60)
		log.Printf("[judger] proxy mode -> remote_judge at %s", *remoteAddr)
	} else {
		handler = judger.MockSandbox{}
		log.Println("[judger] mock mode (no --remote flag)")
	}

	lis, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("listen %s: %v", *addr, err)
	}

	s := grpc.NewServer()
	judger.Register(s, handler)
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
