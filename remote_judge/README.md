# remote_judge

面向 OJ 平台的独立判题子系统。提供 HTTP API 接收代码提交，通过异步队列驱动 Worker 消费判题任务，编译前进行代码黑名单检测，在 Docker 沙箱（容器池复用）中完成编译与运行，返回标准化判题结果（含精确内存峰值测量）。

## 架构

```
OJ Backend → HTTP API → Queue (memory/RabbitMQ) → Worker Pool → Judger (embedded/remote gRPC)
                                                                  │
                                                                  ▼
                                                   Docker CLI Sandbox (C++17 / Go / Python)
```

- **对外**：HTTP REST API（7 个接口），异步提交 + 轮询获取结果
- **对内**：gRPC + 自定义 JSON Codec，Embedded / Remote 双模式
- **沙箱**：Docker CLI 驱动，Bind / Copy 双文件传输模式

## 核心能力

| 模块 | 实现 |
|------|------|
| HTTP API | 提交、查询、测试点详情、语言列表、健康检查、系统统计 |
| 异步判题 | Worker 令牌池并发控制（Buffered Channel，默认 4 并发） |
| 代码黑名单 | AC 自动机多模式匹配，编译前拦截危险代码（system/fork/socket 等） |
| Docker 沙箱 | CLI 驱动，容器池复用（默认开启），cgroup 内存峰值精确测量 |
| 多语言 | C++17、Go 1.22、Python 3.11，配置化扩展，按语言分别黑名单 |
| 判题结果 | 4 中间态 + 7 终态 + System Error，优先级判定链 |
| 熔断降级 | 3 次失败打开 → 30s 半开 → 1 次探测恢复 |
| 安全加固 | cap-drop ALL + Seccomp 白名单 + 禁网 + 非 root + pids-limit |
| 队列 | Memory Queue / RabbitMQ，环境变量切换 |
| 仓储 | Memory Repository / MySQL，环境变量切换 |
| Judger 模式 | Embedded（进程内）/ Remote（独立 gRPC 进程） |
| 部署 | Docker Compose 四服务一键部署（mysql + rabbitmq + judger + server） |
| 可观测性 | 结构化日志 + TraceId 全链路追踪 + Stats 统计接口 |
| 测试 | 70+ 个测试函数，10 个包，Mock + Docker 集成 + Remote Smoke 三层 |

## 环境要求

| 依赖 | 版本要求 | 用途 |
|------|---------|------|
| Go | 1.25+ | 编译运行服务端 |
| Docker Desktop | 24+ | 判题沙箱容器（Mock 模式不需要） |
| 判题镜像 | 需自行构建 | `remote-judge-cpp17`、`remote-judge-go122`、`remote-judge-python311` |
| RabbitMQ | 3.x | 异步队列（使用 Memory Queue 可不需要） |
| MySQL | 8.x | 持久化仓储（使用 Memory Repository 可不需要） |

> **Docker Desktop**：Windows 下使用 Docker 沙箱需要安装 Docker Desktop 并确保其运行中。Mock 模式下不需要 Docker。
>
> **判题镜像**：Docker / Remote / Compose 模式需要提前构建镜像，命令见下方"快速启动"。
>
> **端口冲突**：默认端口 `8080`（HTTP）可能被 Docker Desktop 占用，建议设为 `:8081`；默认 gRPC 端口 `9090` 可能被 Clash 等代理占用，建议设为 `:9091`。

## 目录

```
cmd/
  server/         HTTP API + Embedded Judger 入口
  judger/         独立 gRPC Judger 入口
  smoke/          Smoke 端到端验证工具
  stress/         HTTP 压测工具
  grpcstress/     gRPC 压测工具
internal/
  api/            HTTP Handler（chi Router）
  app/            依赖注入与组件组装
  config/         环境变量配置加载
  domain/         核心领域模型（提交、题目、判题状态、语言配置）
  judger/         判题主逻辑（代码黑名单检测、编译、运行、比对）+ 工作目录对象池
  queue/          队列接口 + Memory / RabbitMQ 实现
  repository/     仓储接口 + Memory / MySQL 实现
  sandbox/        Sandbox 接口 + DockerCLI / Mock 实现 + 熔断器 + Seccomp + 镜像预热
  service/        提交服务（校验、限流）+ 查询服务
  stats/          运行时统计采集器
  transport/      gRPC JSON Codec + gRPC Server + gRPC Client
  worker/         Judge Worker（队列消费、令牌管理）
pkg/pb/           gRPC 请求/响应类型定义
docker/
  images/         判题镜像 Dockerfile（cpp17 / go1.22 / python3.11）
  Dockerfile.server / Dockerfile.judger  服务镜像
docs/             开发周记（week01–08）、测试引导文档（check.md）、截图（photos/）
scripts/          PowerShell 启动脚本
```

## 快速启动

### 嵌入式模式（需 Docker Desktop + MySQL + 判题镜像）

服务直接启动，Judger 与 Server 同进程。默认使用 MySQL 仓储，需确保 MySQL 已运行（参见下方 Compose 部署或自行启动）。

```cmd
cd remote_judge
set REMOTE_JUDGE_HTTP_ADDR=:8081
set REMOTE_JUDGE_GRPC_ADDR=127.0.0.1:9091
set REMOTE_JUDGE_QUEUE=memory
set REMOTE_JUDGE_MYSQL_DSN=root:root@tcp(127.0.0.1:3306)/remote_judge?parseTime=true&charset=utf8mb4
go run .\cmd\server
```

> 快速测试无需 MySQL 时，可设为 `set REMOTE_JUDGE_REPOSITORY=memory` 使用内存仓储。

> 首次启动自动预拉取判题镜像，需提前构建（参见下方"判题镜像构建"）。

```cmd
cd remote_judge
docker build -t remote-judge-cpp17 -f docker/images/cpp17/Dockerfile .
docker build -t remote-judge-go122 -f docker/images/go1.22/Dockerfile .
docker build -t remote-judge-python311 -f docker/images/python3.11/Dockerfile .
```

### Remote 双进程（需两个终端）

终端 1 — Judger：
```cmd
set REMOTE_JUDGE_GRPC_ADDR=127.0.0.1:9091
go run .\cmd\judger
```

终端 2 — Server：
```cmd
set REMOTE_JUDGE_JUDGER_MODE=remote
set REMOTE_JUDGE_HTTP_ADDR=:8081
set REMOTE_JUDGE_GRPC_ADDR=127.0.0.1:9091
go run .\cmd\server
```

### Compose 四服务部署

```cmd
docker compose up -d
docker compose ps
```

## 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `REMOTE_JUDGE_HTTP_ADDR` | `:8080` | HTTP 监听地址 |
| `REMOTE_JUDGE_GRPC_ADDR` | `127.0.0.1:9090` | gRPC 地址 |
| `REMOTE_JUDGE_JUDGER_MODE` | `embedded` | `embedded` / `remote` |
| `REMOTE_JUDGE_JUDGER_ADDR` | — | Remote 模式下的 Judger 地址 |
| `REMOTE_JUDGE_DOCKER_TRANSFER` | `bind` | `bind` / `copy` |
| `REMOTE_JUDGE_QUEUE` | `memory` | `memory` / `rabbitmq` |
| `REMOTE_JUDGE_REPOSITORY` | `mysql` | `mysql` / `memory` |
| `REMOTE_JUDGE_WORKER_CONCURRENCY` | `4` | Worker 并发数 |
| `REMOTE_JUDGE_PREWARM_IMAGES` | `true` | 启动时预拉取判题镜像 |
| `REMOTE_JUDGE_SECCOMP_PROFILE` | `embedded` | Seccomp 白名单（`embedded` / 空禁用） |
| `REMOTE_JUDGE_ENABLE_WORKSPACE_POOL` | `false` | 工作目录对象池 |
| `REMOTE_JUDGE_ENABLE_CONTAINER_POOL` | `true` | 容器池复用（降低延迟 ~2.5x） |
| `REMOTE_JUDGE_CONTAINER_POOL_MAX_SIZE` | `8` | 每镜像最大池化容器数 |
| `REMOTE_JUDGE_RABBITMQ_URL` | `amqp://guest:guest@127.0.0.1:5672/` | RabbitMQ 连接串 |
| `REMOTE_JUDGE_MYSQL_DSN` | `root:root@tcp(127.0.0.1:3306)/remote_judge?parseTime=true&charset=utf8mb4` | MySQL DSN |
| `REMOTE_JUDGE_ENABLE_DEMO_USER_ID` | `true` | 演示用户模式 |

## 测试

完整测试引导见 [`docs/check.md`](docs/check.md)。

### Mock 单元测试（无需 Docker）

```cmd
go test ./internal/api/... ./internal/service/... ./internal/worker/... ./internal/transport/... -count=1 -timeout 60s -v
```

### Docker 集成测试（需 Docker Desktop + 判题镜像）

```cmd
go test -v -count=1 -timeout 120s ./internal/judger/ -run Docker
```

### 全量测试

```cmd
go test ./internal/... -count=1 -timeout 120s
```

### 熔断器测试

```cmd
go test -v -count=1 -timeout 120s ./internal/sandbox/ -run TestCircuitBreaker
```

### Benchmark

```cmd
go test ./internal/service -bench . -benchmem -count=1 -timeout 60s
```

### Smoke 端到端（需启动 Remote 模式服务）

```cmd
go run .\cmd\smoke -addr http://127.0.0.1:8081 -lang cpp17 -mode ac
go run .\cmd\smoke -addr http://127.0.0.1:8081 -lang cpp17 -mode wa
go run .\cmd\smoke -addr http://127.0.0.1:8081 -lang cpp17 -mode ce
go run .\cmd\smoke -addr http://127.0.0.1:8081 -lang python3.11 -mode ac
```

### 压测

```cmd
go run .\cmd\stress -n 100 -c 10 -addr http://127.0.0.1:8081
go run .\cmd\grpcstress -n 30 -c 5 -addr 127.0.0.1:9091 -lang mixed
```
