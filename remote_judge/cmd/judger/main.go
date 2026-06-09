package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"google.golang.org/grpc"

	"remote_judge/internal/config"
	"remote_judge/internal/judger"
	"remote_judge/internal/sandbox"
	grpcserver "remote_judge/internal/transport/grpcserver"
)

// main starts a standalone gRPC Judger process.
func main() {
	cfg := config.Load()
	sb := sandbox.Build(cfg)

	listener, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		log.Fatalf("listen gRPC failed: %v", err)
	}

	server := grpc.NewServer()
	grpcserver.Register(server, grpcserver.New(judger.NewService(sb)))

	go func() {
		stopCh := make(chan os.Signal, 1)
		signal.Notify(stopCh, syscall.SIGINT, syscall.SIGTERM)
		<-stopCh
		server.GracefulStop()
	}()

	if err := sb.Health(context.Background()); err != nil {
		log.Printf("sandbox health check failed: %v", err)
	}
	log.Printf("standalone Judger gRPC listening on %s", cfg.GRPCAddr)
	if err := server.Serve(listener); err != nil {
		log.Fatalf("serve gRPC failed: %v", err)
	}
}
