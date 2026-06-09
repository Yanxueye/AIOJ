package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/terminaloj/backend/internal/config"
	"github.com/terminaloj/backend/internal/database"
	"github.com/terminaloj/backend/internal/handler"
	"github.com/terminaloj/backend/internal/judger"
	"github.com/terminaloj/backend/internal/models"
	"github.com/terminaloj/backend/internal/mq"
	"github.com/terminaloj/backend/internal/utils"
	"gorm.io/gorm"
)

func main() {
	cfgPath := flag.String("config", "config.yaml", "path to config YAML")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := database.Init(cfg.MySQL)
	if err != nil {
		log.Fatalf("db init: %v", err)
	}

	broker, err := mq.NewBroker(cfg.RabbitMQ.URL, cfg.RabbitMQ.SubmitQueue, cfg.RabbitMQ.Enabled)
	if err != nil {
		log.Fatalf("mq init: %v", err)
	}
	defer broker.Close()

	var jdClient judger.JudgerClient
	if cfg.Judger.Remote {
		jdClient = judger.NewRemoteJudger(cfg.Judger.GRPCAddr, cfg.Judger.TimeoutSeconds)
	} else {
		jdClient = judger.NewClient(cfg.Judger.GRPCAddr, cfg.Judger.TimeoutSeconds)
	}
	defer jdClient.Close()

	worker := &mq.Worker{Broker: broker, DB: db, Judger: jdClient, Concurrency: 4}
	workerCtx, cancelWorker := context.WithCancel(context.Background())
	go func() {
		if err := worker.Start(workerCtx); err != nil && err != context.Canceled {
			log.Printf("[worker] stopped: %v", err)
		}
	}()
	go startRejudgeLoop(workerCtx, worker, db)

	jwtMgr := utils.NewJWTManager(cfg.JWT.Secret, cfg.JWT.ExpireHours)
	router := handler.BuildRouter(db, broker, jdClient, jwtMgr, cfg)

	srv := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		log.Printf("[api] listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Println("[api] shutting down...")
	cancelWorker()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}

func startRejudgeLoop(ctx context.Context, worker *mq.Worker, db *gorm.DB) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			var jobs []models.RejudgeJob
			if err := db.Where("status = ?", "pending").Order("id ASC").Limit(2).Find(&jobs).Error; err != nil {
				log.Printf("[rejudge] list jobs failed: %v", err)
				continue
			}
			for _, job := range jobs {
				if err := worker.ProcessRejudgeJob(ctx, job.ID); err != nil {
					log.Printf("[rejudge] process job %d failed: %v", job.ID, err)
				}
			}
		}
	}
}
