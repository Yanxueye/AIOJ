package app

import (
	"context"
	"database/sql"
	"net"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"google.golang.org/grpc"

	"remote_judge/internal/api"
	"remote_judge/internal/config"
	"remote_judge/internal/judger"
	"remote_judge/internal/queue"
	"remote_judge/internal/repository"
	"remote_judge/internal/sandbox"
	"remote_judge/internal/service"
	"remote_judge/internal/stats"
	grpcclient "remote_judge/internal/transport/grpcclient"
	grpcserver "remote_judge/internal/transport/grpcserver"
	"remote_judge/internal/worker"
)

// App 表示整个 remote_judge 系统。
type App struct {
	cfg        config.Config
	httpServer *http.Server
	grpcServer *grpc.Server
	grpcLn     net.Listener
	queue      queue.Queue
	grpcClient *grpcclient.Client
	db         *sql.DB
}

// New 构建完整应用。
func New() (*App, error) {
	cfg := config.Load()
	collector := stats.NewCollector()

	var submissions repository.SubmissionRepository
	var problems repository.ProblemRepository
	var db *sql.DB
	if cfg.Repository == "mysql" {
		conn, err := sql.Open("mysql", cfg.MySQLDSN)
		if err != nil {
			return nil, err
		}
		if err := conn.Ping(); err != nil {
			return nil, err
		}
		db = conn
		submissions = repository.NewMySQLSubmissionRepository(conn)
		problems = repository.NewMySQLProblemRepository(conn)
	} else {
		submissions = repository.NewInMemorySubmissionRepository()
		problems = repository.NewInMemoryProblemRepository()
	}

	var q queue.Queue = queue.NewMemoryQueue(256)
	if cfg.QueueMode == "rabbitmq" {
		rmq, err := queue.NewRabbitMQQueue(cfg.RabbitMQURL, "toj.submit")
		if err != nil {
			return nil, err
		}
		q = rmq
	}

	sb := sandbox.Build(cfg)

	localJudger := judger.NewService(sb)
	if cfg.EnableWorkspacePool {
		localJudger = localJudger.WithWorkspacePool(judger.NewWorkspacePool(4, 16))
	}
	var judgeExecutor judger.Executor = localJudger
	var remoteClient *grpcclient.Client
	if cfg.JudgerMode == "remote" {
		client, err := grpcclient.New(cfg.JudgerAddr)
		if err != nil {
			return nil, err
		}
		remoteClient = client
		judgeExecutor = client
	}
	submissionSvc := service.NewSubmissionService(submissions, problems, q, collector)
	querySvc := service.NewQueryService(submissions)
	httpAPI := api.NewHTTPServer(submissionSvc, querySvc, collector, cfg)
	worker := worker.NewJudgeWorker(q, submissions, problems, judgeExecutor, collector, cfg.WorkerConcurrency)

	ctx := context.Background()
	if err := worker.Start(ctx); err != nil {
		return nil, err
	}

	httpServer := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           api.FakeAuthMiddleware(httpAPI.Routes()),
		ReadHeaderTimeout: 5 * time.Second,
	}

	var grpcSrv *grpc.Server
	var grpcLn net.Listener
	if cfg.JudgerMode == "embedded" {
		grpcSrv = grpc.NewServer()
		grpcserver.Register(grpcSrv, grpcserver.New(localJudger))
		var err error
		grpcLn, err = net.Listen("tcp", cfg.GRPCAddr)
		if err != nil {
			return nil, err
		}
	}

	return &App{
		cfg:        cfg,
		httpServer: httpServer,
		grpcServer: grpcSrv,
		grpcLn:     grpcLn,
		queue:      q,
		grpcClient: remoteClient,
		db:         db,
	}, nil
}

// Run 启动 HTTP 与 gRPC 服务。
func (a *App) Run() error {
	if a.grpcServer != nil && a.grpcLn != nil {
		go func() {
			_ = a.grpcServer.Serve(a.grpcLn)
		}()
	}
	return a.httpServer.ListenAndServe()
}

// Shutdown 优雅关闭系统。
func (a *App) Shutdown(ctx context.Context) error {
	if a.grpcServer != nil {
		a.grpcServer.GracefulStop()
	}
	if a.grpcClient != nil {
		_ = a.grpcClient.Close()
	}
	_ = a.queue.Close()
	if a.db != nil {
		_ = a.db.Close()
	}
	return a.httpServer.Shutdown(ctx)
}
