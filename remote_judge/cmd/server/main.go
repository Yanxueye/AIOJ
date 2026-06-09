package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"remote_judge/internal/app"
)

// main starts the HTTP API, worker, and optional embedded gRPC Judger server.
func main() {
	application, err := app.New()
	if err != nil {
		log.Fatalf("build app failed: %v", err)
	}

	go func() {
		stopCh := make(chan os.Signal, 1)
		signal.Notify(stopCh, syscall.SIGINT, syscall.SIGTERM)
		<-stopCh
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = application.Shutdown(ctx)
	}()

	log.Println("remote_judge service starting")
	if err := application.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("run app failed: %v", err)
	}
}
