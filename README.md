# Fused

`fused` 是一个融合仓库，用来把 `AIOJ-main`、`remote_judge` 和 `agent-service` 放在同一个代码库中协同开发。

当前仓库里包含三个核心子项目：

- `AIOJ-main`
  一个 AI 辅助在线判题平台，包含 Vue 3 前端和 Go 后端，负责用户系统、题目管理、提交入口、提交记录、学习统计、知识图谱和个性化推荐。
- `remote_judge`
  一个独立的判题服务，负责代码编译、沙箱运行、测试点判定、状态汇总，以及更细粒度的判题结果输出。
- `agent-service`
  一个 AI 微服务，提供大模型对话、代码分析、提示生成、题解辅助等能力，支持 MIMO API（主）+ 本地 Ollama（降级）双模型策略。

## 当前集成方式

当前仓库采用的主集成方案是：

- `AIOJ-main/backend` 负责接收提交请求
- 提交任务进入 RabbitMQ
- `AIOJ-main/backend` 的 worker 从 MySQL 读取题目与测试点
- worker 通过 gRPC 调用 `remote_judge/cmd/judger`
- `remote_judge` 返回完整判题结果
- `AIOJ-main/backend` 把结果写回自己的提交记录
- AI 相关请求通过 `AIOJ-main/backend` 转发给 `agent-service`

这条链路保留了各服务的职责边界：

- `AIOJ-main` 负责平台业务和提交流转
- `remote_judge` 负责纯判题执行
- `agent-service` 负责 AI 能力（对话、分析、RAG）

同时，仓库里也保留了 `remote_judge/cmd/server` 这一完整判题后端入口，便于后续尝试"由 remote_judge 独立托管提交队列和判题流程"的另一种方案。

## 已对齐的判题结果信息

当前融合后的主链路已经尽量向 `remote_judge` 的结果模型对齐，AIOJ 侧已接入这些 richer 字段：

- `status`
  包含 `Pending`、`Queueing`、`Compiling`、`Running`、`Accepted`、`Wrong Answer`、`Compile Error`、`Runtime Error`、`Time Limit Exceeded`、`Memory Limit Exceeded`、`Output Limit Exceeded`、`System Error`
- `traceId`
- `runtimeMs`
- `memoryKb`
- `compileOutput`
- `errorMessage`
- `caseResults`
- `stdoutBytes`
- `stderrBytes`
- `signal`
- `queueStartedAt`
- `judgeStartedAt`
- `finishedAt`

也就是说，当前不再只保留"是否 Accepted"这种扁平信息，而是尽量保留 `remote_judge` 的完整判题细节。

## 仓库结构

```text
fused/
├─ AIOJ-main/
│  ├─ backend/
│  │  ├─ cmd/
│  │  ├─ docker/
│  │  ├─ internal/
│  │  ├─ proto/
│  │  ├─ API.md
│  │  └─ config.yaml
│  ├─ frontend/
│  │  ├─ src/
│  │  └─ package.json
│  ├─ README.md
│  ├─ PROGRESS.md
│  └─ WORK_SUMMARY.md
├─ remote_judge/
│  ├─ cmd/
│  ├─ docker/
│  ├─ internal/
│  ├─ pkg/
│  ├─ proto/
│  ├─ scripts/
│  └─ README.md
├─ agent-service/
│  ├─ cmd/server/
│  ├─ internal/
│  ├─ .env            (需自行创建，含 API Key)
│  ├─ go.mod
│  └─ README.md
├─ GOAL.md
└─ CLAUDE.md
```

## 技术栈

### AIOJ-main

- Frontend: Vue 3, Vite, Pinia, Vue Router, Element Plus, Monaco Editor, ECharts
- Backend: Go 1.21, Gin, GORM, MySQL, RabbitMQ, gRPC, JWT

### remote_judge

- Go 1.25+
- Docker CLI sandbox
- gRPC + JSON codec
- Memory / RabbitMQ queue
- Memory / MySQL repository

### agent-service

- Go 1.21, Gin
- MIMO API（OpenAI 兼容，主模型）
- Ollama（本地，降级模型）
- HTTP 通信，经 AIOJ 后端转发

## 环境要求

至少需要以下环境：

- Go
  `AIOJ-main/backend` 和 `agent-service` 使用 Go 1.21，`remote_judge` 使用 Go 1.25+
- Node.js / npm
  用于 `AIOJ-main/frontend`
- MySQL 8.x
  AIOJ 后端使用
- RabbitMQ 3.x
  AIOJ 提交队列使用
- Docker Desktop
  `remote_judge` 运行真实沙箱时使用
- Ollama（可选）
  `agent-service` 降级模型使用

## 项目配置

### 1. 环境依赖安装

| 依赖 | 版本要求 | 用途 |
|------|---------|------|
| Go | 1.21+（AIOJ、agent-service）/ 1.25+（remote_judge） | 后端编译运行 |
| Node.js / npm | 18+ | 前端编译运行 |
| MySQL | 8.x | 数据存储 |
| RabbitMQ | 3.x | 提交消息队列 |
| Docker Desktop | 最新版 | 判题沙箱运行 |

### 2. Docker 镜像拉取/构建

启动前需确保以下镜像可用：

```cmd
REM 启动 MySQL + RabbitMQ（AIOJ 依赖）
cd AIOJ-main\backend
docker compose -f docker\docker-compose.yml up -d mysql rabbitmq

REM 构建判题沙箱镜像（remote_judge 需要）
cd remote_judge
docker build -t remote-judge-cpp17 -f docker\images\cpp17\Dockerfile .
docker build -t remote-judge-go122 -f docker\images\go1.22\Dockerfile .
docker build -t remote-judge-python311 -f docker\images\python3.11\Dockerfile .
```

### 3. API Key / URL 配置

#### agent-service 配置

在 `agent-service/` 目录下创建 `.env` 文件：

```env
# AI 服务提供商：openai（使用 MIMO API）或 ollama（本地模型）
AI_PROVIDER=openai

# MIMO API 配置（主模型）
OPENAI_API_KEY=your-mimo-api-key-here
OPENAI_BASE_URL=https://token-plan-sgp.xiaomimimo.com/v1
OPENAI_MODEL=mimo-v2.5-pro

# Ollama 配置（本地降级模型，可选）
OLLAMA_URL=http://127.0.0.1:11434
OLLAMA_MODEL=qwen2.5-coder:7b

# 服务地址
AGENT_HTTP_ADDR=:8090
AIOJ_BACKEND_URL=http://127.0.0.1:8080
```

> **说明**：如不配置 `OPENAI_API_KEY`，agent-service 将使用内置 Mock 回复（仅用于开发调试）。如需使用 Ollama 本地模型，先安装 Ollama 并拉取模型：`ollama pull qwen2.5-coder:7b`。

#### AIOJ 后端配置

编辑 `AIOJ-main/backend/config.yaml`，关键配置项：

```yaml
# MySQL 连接
mysql:
  dsn: "toj:toj_password@tcp(127.0.0.1:3306)/terminaloj?charset=utf8mb4&parseTime=True&loc=Local"

# RabbitMQ 连接
rabbitmq:
  url: "amqp://guest:guest@127.0.0.1:5672/"
  enabled: true

# 判题服务（remote_judge gRPC 地址）
judger:
  grpc_addr: "127.0.0.1:9090"
  remote: true

# AI 服务（agent-service 地址，可选）
ai:
  enabled: false                                    # 设为 true 启用 AI 功能
  endpoint: "http://127.0.0.1:8090/api/agent"       # 指向 agent-service
  api_key: ""
  model: "mimo-v2.5-pro"
  timeout_seconds: 20
```

> **说明**：`ai.enabled` 默认为 `false`，此时 AI 功能使用内置 Mock。设为 `true` 并确保 agent-service 运行后，即可使用完整 AI 功能。

### 4. 端口总览

| 服务 | 默认端口 | 说明 |
|------|---------|------|
| AIOJ 前端 | :5173 | Vite 开发服务器 |
| AIOJ 后端 | :8080 | Gin HTTP API |
| remote_judge | :9090 | gRPC 判题服务 |
| agent-service | :8090 | AI 微服务 |
| MySQL | :3306 | 数据库 |
| RabbitMQ | :5672 | 消息队列 |
| RabbitMQ 管理面板 | :15672 | Web UI（guest/guest） |

## 快速启动

下面的步骤以当前默认集成方式为准，即：

- `AIOJ-main/backend` -> RabbitMQ -> worker -> gRPC -> `remote_judge/cmd/judger`
- AI 请求 -> `AIOJ-main/backend` -> `agent-service`

### 1. 启动 AIOJ 依赖

```cmd
cd AIOJ-main\backend
docker compose -f docker/docker-compose.yml up -d mysql rabbitmq
```

### 2. 构建 remote_judge 判题镜像

```cmd
cd remote_judge
docker build -t remote-judge-cpp17 -f docker/images/cpp17/Dockerfile .
docker build -t remote-judge-go122 -f docker/images/go1.22/Dockerfile .
docker build -t remote-judge-python311 -f docker/images/python3.11/Dockerfile .
```

### 3. 启动 remote_judge gRPC 判题服务

当前 AIOJ 默认配置指向 `127.0.0.1:9090`。

```cmd
cd remote_judge
set REMOTE_JUDGE_GRPC_ADDR=127.0.0.1:9090
go run .\cmd\judger
```

### 4. 启动 agent-service

需要先创建配置文件（参考 `agent-service/README.md`）：

```cmd
cd agent-service
go run .\cmd\server
```

默认监听：

- API: `http://127.0.0.1:8090`

### 5. 启动 AIOJ 后端

```cmd
cd AIOJ-main\backend
go run .\cmd\server -config config.yaml
```

默认监听：

- API: `http://127.0.0.1:8080`

### 6. 启动 AIOJ 前端

```cmd
cd AIOJ-main\frontend
npm install
npm run dev
```

通常前端开发服务器地址是：

- `http://127.0.0.1:5173`

## 默认账号

当前 AIOJ 默认种子账号：

```text
普通用户:
username: coder_test
password: 123456

管理员:
username: admin
password: 123456
```

## 另一种保留方案

仓库中仍然保留了另一种集成方向：

- AIOJ 后端只做薄代理
- 提交通过 HTTP 转发给 `remote_judge/cmd/server`
- 判题队列、Worker、状态推进更多由 `remote_judge` 自己负责

## 常用命令

### AIOJ backend tests

```cmd
cd AIOJ-main\backend
go test ./...
```

### remote_judge tests

```cmd
cd remote_judge
go test ./...
```

### AIOJ frontend build

```cmd
cd AIOJ-main\frontend
npm run build
```

### agent-service build

```cmd
cd agent-service
go build ./...
```

## 相关文档

- `AIOJ-main/backend/API.md`
  AIOJ 后端 API 文档
- `AIOJ-main/frontend/API.md`
  AIOJ 前端接口约定
- `AIOJ-main/README.md`
  AIOJ 项目原始说明
- `remote_judge/README.md`
  remote_judge 项目原始说明
- `agent-service/README.md`
  agent-service AI 微服务说明
