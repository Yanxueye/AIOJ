# TerminalOJ Backend

Go + Gin + GORM + RabbitMQ + gRPC 版在线判题后端，匹配 `frontend/API.md` 契约。

## 快速开始

```bash
# 基础设施
docker compose -f docker/docker-compose.yml up -d mysql rabbitmq

# 依赖
go mod tidy

# 判题容器（可独立进程）
go run ./cmd/judger

# API 服务
go run ./cmd/server -config config.yaml
```

默认账号：`coder_test` / `123456`。详见 [API.md](API.md) 与 [PROGRESS.md](PROGRESS.md)。

## 目录速览

- `cmd/server/` — HTTP API 入口
- `cmd/judger/` — gRPC 判题服务入口（Docker 镜像 `Dockerfile.judger`）
- `internal/handler/` — Gin 路由与业务 Handler
- `internal/mq/` — RabbitMQ 生产者 + Worker 消费者（带内存降级）
- `internal/judger/` — gRPC 客户端、JSON Codec、MockSandbox
- `internal/models/` — GORM 实体与 DTO
- `proto/judger.proto` — 跨语言判题契约
