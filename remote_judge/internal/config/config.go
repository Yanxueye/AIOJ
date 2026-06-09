package config

import "os"

// Config 描述服务的运行时配置。
type Config struct {
	HTTPAddr          string
	GRPCAddr          string
	JudgerMode        string
	JudgerAddr        string
	QueueMode         string
	Repository        string
	DockerTransfer    string
	RabbitMQURL       string
	MySQLDSN          string
	EnableDemoID      bool
	WorkerConcurrency int
	PrewarmImages     bool
	SeccompProfile      string
	EnableWorkspacePool  bool
	EnableContainerPool  bool
	ContainerPoolMaxSize int
}

// Load 从环境变量读取配置。
func Load() Config {
	return Config{
		HTTPAddr:          valueOrDefault("REMOTE_JUDGE_HTTP_ADDR", ":8080"),
		GRPCAddr:          valueOrDefault("REMOTE_JUDGE_GRPC_ADDR", "127.0.0.1:9090"),
		JudgerMode:        valueOrDefault("REMOTE_JUDGE_JUDGER_MODE", "embedded"),
		JudgerAddr:        valueOrDefault("REMOTE_JUDGE_JUDGER_ADDR", "127.0.0.1:9090"),
		QueueMode:         valueOrDefault("REMOTE_JUDGE_QUEUE", "memory"),
		Repository:        valueOrDefault("REMOTE_JUDGE_REPOSITORY", "mysql"),
		DockerTransfer:    valueOrDefault("REMOTE_JUDGE_DOCKER_TRANSFER", "bind"),
		RabbitMQURL:       valueOrDefault("REMOTE_JUDGE_RABBITMQ_URL", "amqp://guest:guest@127.0.0.1:5672/"),
		MySQLDSN:          valueOrDefault("REMOTE_JUDGE_MYSQL_DSN", "root:root@tcp(127.0.0.1:3306)/remote_judge?parseTime=true&charset=utf8mb4"),
		EnableDemoID:      valueOrDefault("REMOTE_JUDGE_ENABLE_DEMO_USER_ID", "true") == "true",
		WorkerConcurrency: intValueOrDefault("REMOTE_JUDGE_WORKER_CONCURRENCY", 4),
		PrewarmImages:     valueOrDefault("REMOTE_JUDGE_PREWARM_IMAGES", "true") == "true",
		SeccompProfile:      valueOrDefault("REMOTE_JUDGE_SECCOMP_PROFILE", "embedded"),
		EnableWorkspacePool:  valueOrDefault("REMOTE_JUDGE_ENABLE_WORKSPACE_POOL", "false") == "true",
		EnableContainerPool:  valueOrDefault("REMOTE_JUDGE_ENABLE_CONTAINER_POOL", "true") == "true",
		ContainerPoolMaxSize: intValueOrDefault("REMOTE_JUDGE_CONTAINER_POOL_MAX_SIZE", 8),
	}
}

// valueOrDefault reads an environment variable with fallback.
func valueOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// intValueOrDefault reads a positive integer environment variable with fallback.
func intValueOrDefault(key string, fallback int) int {
	v := valueOrDefault(key, "")
	if v == "" {
		return fallback
	}
	n := 0
	for _, ch := range v {
		if ch < '0' || ch > '9' {
			return fallback
		}
		n = n*10 + int(ch-'0')
	}
	if n <= 0 {
		return fallback
	}
	return n
}
