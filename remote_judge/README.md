# remote_judge

> **独立判题子系统** — 面向 OJ 平台的 Docker 沙箱判题引擎。HTTP API 接收代码提交，异步队列驱动 Worker 消费，Docker 容器编译运行，返回标准化判题结果（含精确内存峰值测量）。

![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)
![Docker](https://img.shields.io/badge/Docker-Sandbox-2496ED?logo=docker)
![gRPC](https://img.shields.io/badge/gRPC-JSON_Codec-4285F4)
![License](https://img.shields.io/badge/license-MIT-blue)

---

## 目录

- [架构](#架构)
  - [当前集成方式 (fused)](#当前集成方式-fused)
  - [独立部署 (Standalone)](#独立部署-standalone)
  - [Judger 判题步骤](#judger-判题步骤)
  - [Sandbox 安全与性能](#sandbox-安全与性能)
  - [三种部署模式](#三种部署模式)
- [核心能力](#核心能力)
- [判题流程](#判题流程)
- [快速启动](#快速启动)
- [环境变量](#环境变量)
- [测试](#测试)
- [目录结构](#目录结构)
- [相关文档](#相关文档)

---

## 架构

### 当前集成方式 (fused)

remote_judge 在 fused 融合仓库中的实际角色：AIOJ Backend 通过 **gRPC 直连 `cmd/judger`**，将判题作为一个纯函数调用嵌入 AIOJ 的提交流程。队列调度和状态管理由 AIOJ 侧负责，remote_judge 专注代码执行与判定。

```mermaid
flowchart LR
    subgraph AIOJ["AIOJ 侧 (平台业务)"]
        direction TB
        Frontend["Vue 3 前端"]
        Backend["AIOJ Backend<br/>Gin :8080"]
        RMQ[("RabbitMQ")]
        W["MQ Worker"]

        Frontend -->|"POST /api/submissions"| Backend
        Backend -->|"Publish JudgeTask"| RMQ
        RMQ -->|"Consume"| W
        W -->|"gRPC<br/>JudgeRequest"| Judger
    end

    subgraph remote["remote_judge 侧 (纯判题)"]
        Judger["cmd/judger<br/>gRPC :9090<br/>────<br/>黑名单 → 编译<br/>Docker 运行 → 比对"]
        Sandbox[("Docker<br/>Sandbox")]
        Judger -->|"Docker CLI"| Sandbox
    end

    Judger -->|"JudgeResponse<br/>{status, runtimeMs,<br/>memoryKb, caseResults[],<br/>traceId, ...}"| W

    style AIOJ fill:#e3f2fd,stroke:#2196f3
    style Backend fill:#1565c0,stroke:#0d47a1,color:#fff
    style RMQ fill:#fff3e0,stroke:#ff9800
    style W fill:#1565c0,stroke:#0d47a1,color:#fff
    style remote fill:#e8f5e9,stroke:#4caf50
    style Judger fill:#2e7d32,stroke:#1b5e20,color:#fff
    style Sandbox fill:#c8e6c9,stroke:#4caf50
```

> AIOJ Backend 通过 `config.yaml` 中 `judger.grpc_addr` 指向 remote_judge Judger。请求为 gRPC + 自定义 JSON Codec，无需 protoc 编译。

### 独立部署 (Standalone)

remote_judge 同时保留完整的独立判题后端能力，`cmd/server` 内置 HTTP API、队列调度、Worker 消费和 Judger 调用，可脱离 AIOJ 作为通用的 OJ 判题服务运行。

```mermaid
flowchart LR
    OJ["OJ Backend<br/>(AIOJ / 第三方)"] -->|"HTTP REST"| Server["cmd/server<br/>HTTP API :8080<br/>chi Router"]
    Server -->|"enqueue"| Queue["Queue<br/>Memory ｜ RabbitMQ"]
    Queue -->|"dequeue"| Worker["Worker Pool<br/>令牌池 · 默认 4 并发"]
    Worker -->|"调用"| Judger["Judger<br/>Embedded ｜ Remote gRPC"]
    Judger -->|"Docker CLI"| Docker[("Docker<br/>Sandbox")]

    style Server fill:#0f3460,stroke:#16213e,color:#fff
    style Queue fill:#0f3460,stroke:#16213e,color:#fff
    style Worker fill:#0f3460,stroke:#16213e,color:#fff
    style Judger fill:#0f3460,stroke:#16213e,color:#fff
    style Docker fill:#e3f2fd,stroke:#2196f3
```

### Judger 判题步骤

```mermaid
flowchart LR
    Input["gRPC JudgeRequest<br/>{code, language,<br/>timeLimit, testCases}"] --> BL["代码黑名单<br/>AC 自动机<br/>拦截 system/fork/socket"]
    BL -->|"通过"| Compile["编译<br/>按语言选编译器<br/>时限内完成"]
    Compile -->|"成功"| Run["Docker 运行<br/>容器池复用<br/>cgroup 内存峰值测量"]
    Compile -->|"失败"| CE["Compile Error"]
    Run -->|"正常退出"| Compare["结果比对<br/>逐测试点<br/>优先级判决链"]
    Run -->|"异常"| RE["Runtime Error"]
    Compare -->|"全部通过"| AC["Accepted"]
    Compare -->|"输出不一致"| WA["Wrong Answer"]
    Compare -->|"超时"| TLE["Time Limit Exceeded"]
    Compare -->|"超内存"| MLE["Memory Limit Exceeded"]
    Compare -->|"输出超限"| OLE["Output Limit Exceeded"]

    style Input fill:#f5f5f5,stroke:#999
    style BL fill:#fff3e0,stroke:#ff9800
    style Compile fill:#fff3e0,stroke:#ff9800
    style Run fill:#e3f2fd,stroke:#2196f3
    style Compare fill:#f3e5f5,stroke:#9c27b0,color:#333
    style AC fill:#c8e6c9,stroke:#4caf50
    style WA fill:#ffcdd2,stroke:#f44336
    style CE fill:#ffcdd2,stroke:#f44336
    style RE fill:#ffcdd2,stroke:#f44336
    style TLE fill:#ffcdd2,stroke:#f44336
    style MLE fill:#ffcdd2,stroke:#f44336
    style OLE fill:#ffcdd2,stroke:#f44336
```

### Sandbox 安全与性能

```mermaid
flowchart LR
    subgraph Security["安全加固"]
        direction TB
        S1["cap-drop ALL<br/>移除所有内核能力"]
        S2["Seccomp 白名单<br/>仅允许安全系统调用"]
        S3["禁网 · 非 root<br/>pids-limit"]
    end

    subgraph Performance["性能优化"]
        direction TB
        P1["容器池复用<br/>延迟降低 ~2.5x<br/>每镜像最大 8 个"]
        P2["工作目录对象池<br/>(可选开启)"]
    end

    subgraph Resilience["韧性保障"]
        direction TB
        R1["熔断器<br/>3 次失败打开<br/>30s 半开探测<br/>1 次成功恢复"]
        R2["镜像预热<br/>启动时预拉取"]
    end

    style Security fill:#ffebee,stroke:#f44336
    style Performance fill:#e8f5e9,stroke:#4caf50
    style Resilience fill:#fff3e0,stroke:#ff9800
```

### 三种部署模式

```mermaid
flowchart TB
    subgraph Embedded["Embedded — 单进程"]
        direction LR
        E_Flow["HTTP API → Queue → Worker → Judger<br/>全部在 cmd/server 同进程内"]
    end

    subgraph Remote["Remote — 双进程"]
        direction LR
        R_Server["cmd/server<br/>HTTP API :8080"] -->|"gRPC :9090"| R_Judger["cmd/judger<br/>Judger 独立进程"]
    end

    subgraph Compose["Compose — 容器化"]
        direction LR
        C_Server["Server 容器"] --> C_RMQ["RabbitMQ 容器"]
        C_Server --> C_MySQL["MySQL 容器"]
        C_Server --> C_Judger["Judger 容器"]
    end

    Embedded --> Remote --> Compose

    style Embedded fill:#e8f5e9,stroke:#4caf50
    style Remote fill:#fff3e0,stroke:#ff9800
    style Compose fill:#e3f2fd,stroke:#2196f3
```

---

## 核心能力

| 模块 | 实现 |
|------|------|
| HTTP API | 7 个 REST 接口: 提交、查询、测试点详情、语言列表、健康检查、系统统计、Judger 健康 |
| 异步判题 | Worker 令牌池并发控制 (Buffered Channel, 默认 4 并发) |
| 代码黑名单 | AC 自动机多模式匹配, 编译前拦截危险代码 (system / fork / socket / exec 等) |
| Docker 沙箱 | CLI 驱动, 容器池复用 (默认开启), cgroup 内存峰值精确测量 |
| 多语言 | C++17 / Go 1.22 / Python 3.11, 配置化扩展, 按语言分别黑名单 |
| 判题结果 | 4 中间态 + 7 终态 + System Error, 优先级判定链 |
| 熔断降级 | 3 次失败打开 → 30s 半开 → 1 次探测恢复 |
| 安全加固 | cap-drop ALL + Seccomp 白名单 + 禁网 + 非 root + pids-limit |
| 队列 | Memory Queue / RabbitMQ, 环境变量切换 |
| 仓储 | Memory Repository / MySQL, 环境变量切换 |
| Judger 模式 | Embedded (进程内) / Remote (独立 gRPC 进程) |
| 部署 | Docker Compose 四服务一键部署 (mysql + rabbitmq + judger + server) |
| 可观测性 | 结构化日志 + TraceId 全链路追踪 + Stats 统计接口 |
| 测试 | 70+ 测试函数, 10 个包, Mock + Docker 集成 + Remote Smoke 三层 |

---

## 判题流程

### 状态机

```mermaid
stateDiagram-v2
    [*] --> Pending : 接收提交
    Pending --> Queueing : 入队
    Queueing --> Compiling : Worker 领取任务

    state Compiling {
        [*] --> 编译中
        编译中 --> 编译通过 : success
        编译中 --> 编译失败 : error
    }

    Compiling --> Running : 编译通过
    Compiling --> CompileError : 编译失败

    state Running {
        [*] --> 执行中
        执行中 --> 正常退出 : exit=0
        执行中 --> 运行异常 : signal/非0
        执行中 --> 超时 : TLE
        执行中 --> 超内存 : MLE
    }

    Running --> Comparing : 正常退出
    Running --> RuntimeError : 运行异常
    Running --> TimeLimitExceeded : 超时
    Running --> MemoryLimitExceeded : 超内存

    Comparing --> Accepted : 全部通过
    Comparing --> WrongAnswer : 输出不一致
    Comparing --> OutputLimitExceeded : 输出超限

    CompileError --> [*]
    RuntimeError --> [*]
    TimeLimitExceeded --> [*]
    MemoryLimitExceeded --> [*]
    Accepted --> [*]
    WrongAnswer --> [*]
    OutputLimitExceeded --> [*]
```

### 判题结果结构

```
JudgeResult
├── status              状态: Pending → Queueing → Compiling → Running → 终态
├── traceId             全链路追踪 ID
├── runtimeMs           运行耗时 (毫秒)
├── memoryKb            内存峰值 (KB)
├── compileOutput       编译器输出
├── errorMessage        错误描述
├── signal              终止信号
├── caseResults[]       每个测试点的判定结果
│   ├── caseIndex       测试点编号
│   ├── status          该点状态
│   ├── runtimeMs       该点耗时
│   ├── memoryKb        该点内存
│   └── judgeOutput     输出 (截断)
├── stdoutBytes         标准输出
├── stderrBytes         标准错误
├── queueStartedAt      入队时间
├── judgeStartedAt      开始判题时间
└── finishedAt          完成时间
```

---

## 快速启动

### 环境要求

| 依赖 | 版本要求 | 用途 |
|------|---------|------|
| Go | 1.25+ | 编译运行服务端 |
| Docker Desktop | 24+ | 判题沙箱容器 (Mock 模式不需要) |
| 判题镜像 | 需自行构建 | `remote-judge-cpp17` / `remote-judge-go122` / `remote-judge-python311` |
| RabbitMQ | 3.x | 异步队列 (使用 Memory Queue 可不需要) |
| MySQL | 8.x | 持久化仓储 (使用 Memory Repository 可不需要) |

> **Docker Desktop**: Windows 下使用 Docker 沙箱需要安装 Docker Desktop 并确保其运行中。Mock 模式下不需要 Docker。
>
> **端口冲突**: 默认 HTTP 端口 `:8080` 可能与 Docker Desktop 冲突, 建议设为 `:8081`。

### 判题镜像构建

```cmd
cd remote_judge
docker build -t remote-judge-cpp17    -f docker\images\cpp17\Dockerfile .
docker build -t remote-judge-go122    -f docker\images\go1.22\Dockerfile .
docker build -t remote-judge-python311 -f docker\images\python3.11\Dockerfile .
```

### Embedded 模式 (单进程, 需 Docker + MySQL)

```cmd
cd remote_judge
set REMOTE_JUDGE_HTTP_ADDR=:8081
set REMOTE_JUDGE_GRPC_ADDR=127.0.0.1:9090
set REMOTE_JUDGE_QUEUE=memory
set REMOTE_JUDGE_MYSQL_DSN=root:root@tcp(127.0.0.1:3306)/remote_judge?parseTime=true&charset=utf8mb4
go run .\cmd\server
```

> 快速测试无需 MySQL 时: `set REMOTE_JUDGE_REPOSITORY=memory` 使用内存仓储。
>
> 首次启动自动预拉取判题镜像 (需提前构建)。

### Remote 双进程 (需两个终端)

**终端 1 — Judger:**
```cmd
set REMOTE_JUDGE_GRPC_ADDR=127.0.0.1:9090
go run .\cmd\judger
```

**终端 2 — Server:**
```cmd
set REMOTE_JUDGE_JUDGER_MODE=remote
set REMOTE_JUDGE_HTTP_ADDR=:8081
set REMOTE_JUDGE_GRPC_ADDR=127.0.0.1:9090
go run .\cmd\server
```

### Compose 四服务部署

```cmd
docker compose up -d
docker compose ps
```

---

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
| `REMOTE_JUDGE_SECCOMP_PROFILE` | `embedded` | Seccomp 白名单 (`embedded` / 空禁用) |
| `REMOTE_JUDGE_ENABLE_WORKSPACE_POOL` | `false` | 工作目录对象池 |
| `REMOTE_JUDGE_ENABLE_CONTAINER_POOL` | `true` | 容器池复用 (降低延迟 ~2.5x) |
| `REMOTE_JUDGE_CONTAINER_POOL_MAX_SIZE` | `8` | 每镜像最大池化容器数 |
| `REMOTE_JUDGE_RABBITMQ_URL` | `amqp://guest:guest@127.0.0.1:5672/` | RabbitMQ 连接串 |
| `REMOTE_JUDGE_MYSQL_DSN` | `root:root@tcp(127.0.0.1:3306)/remote_judge?parseTime=true&charset=utf8mb4` | MySQL DSN |
| `REMOTE_JUDGE_ENABLE_DEMO_USER_ID` | `true` | 演示用户模式 |

---

## 测试

完整测试引导见 [`docs/check.md`](docs/check.md)。

| 类型 | 命令 | 说明 |
|------|------|------|
| **Mock 单元** | `go test ./internal/api/... ./internal/service/... ./internal/worker/... ./internal/transport/... -count=1 -timeout 60s -v` | 无需 Docker |
| **Docker 集成** | `go test -v -count=1 -timeout 120s ./internal/judger/ -run Docker` | 需 Docker Desktop + 判题镜像 |
| **全量测试** | `go test ./internal/... -count=1 -timeout 120s` | 全部 70+ 测试 |
| **熔断器** | `go test -v -count=1 -timeout 120s ./internal/sandbox/ -run TestCircuitBreaker` | 熔断器专项 |
| **Benchmark** | `go test ./internal/service -bench . -benchmem -count=1 -timeout 60s` | 性能基准 |

### Smoke 端到端 (需启动 Remote 模式服务)

```cmd
go run .\cmd\smoke -addr http://127.0.0.1:8081 -lang cpp17 -mode ac
go run .\cmd\smoke -addr http://127.0.0.1:8081 -lang cpp17 -mode wa
go run .\cmd\smoke -addr http://127.0.0.1:8081 -lang cpp17 -mode ce
go run .\cmd\smoke -addr http://127.0.0.1:8081 -lang python3.11 -mode ac
```

### 压测

```cmd
go run .\cmd\stress     -n 100 -c 10 -addr http://127.0.0.1:8081
go run .\cmd\grpcstress -n 30 -c 5 -addr 127.0.0.1:9090 -lang mixed
```

---

## 目录结构

```
remote_judge/
│
├── cmd/
│   ├── server/                         # HTTP API + Embedded Judger 入口
│   ├── judger/                         # 独立 gRPC Judger 入口
│   ├── smoke/                          # Smoke 端到端验证工具
│   ├── stress/                         # HTTP 压测工具
│   └── grpcstress/                     # gRPC 压测工具
│
├── internal/
│   ├── api/                            # HTTP Handler (chi Router)
│   ├── app/                            # 依赖注入与组件组装
│   ├── config/                         # 环境变量配置加载
│   ├── domain/                         # 核心领域模型
│   │                                   #   Submission / Problem / JudgeStatus / LanguageConfig
│   ├── judger/                         # 判题主逻辑
│   │                                   #   代码黑名单检测 · 编译 · 运行 · 比对 · 工作目录对象池
│   ├── queue/                          # 队列接口 + Memory / RabbitMQ 实现
│   ├── repository/                     # 仓储接口 + Memory / MySQL 实现
│   ├── sandbox/                        # Sandbox 接口
│   │                                   #   DockerCLI / Mock 实现 · 熔断器 · Seccomp · 镜像预热
│   ├── service/                        # 提交服务 (校验 + 限流) + 查询服务
│   ├── stats/                          # 运行时统计采集器
│   ├── transport/                      # gRPC JSON Codec + Server + Client
│   └── worker/                         # Judge Worker (队列消费 + 令牌管理)
│
├── pkg/pb/                             # gRPC 请求/响应类型定义
│
├── docker/
│   ├── images/                         # 判题镜像 Dockerfile
│   │   ├── cpp17/Dockerfile
│   │   ├── go1.22/Dockerfile
│   │   └── python3.11/Dockerfile
│   ├── Dockerfile.server               # Server 服务镜像
│   └── Dockerfile.judger               # Judger 服务镜像
│
├── docs/                               # 开发周记 (week01–08) + 测试引导 + 截图
├── scripts/                            # PowerShell 启动脚本
└── README.md
```

---

## 相关文档

| 文档 | 说明 |
|------|------|
| [docs/check.md](docs/check.md) | 完整测试引导 |
| [AIOJ-main/README.md](../AIOJ-main/README.md) | AIOJ 主项目说明 |
| [agent-service/README.md](../agent-service/README.md) | AI 微服务说明 |
| [../README.md](../README.md) | 融合仓库总览 |
| [../CLAUDE.md](../CLAUDE.md) | 开发指南 |
| [../PROJECT_GAPS.md](../PROJECT_GAPS.md) | 项目缺陷与改进计划 |
