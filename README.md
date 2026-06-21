# Fused

> **融合仓库** — 将 `AIOJ-main`、`remote_judge`、`agent-service` 整合于同一代码库协同开发。

![Go](https://img.shields.io/badge/Go-1.21%2F1.25%2B-00ADD8?logo=go)
![Vue](https://img.shields.io/badge/Vue-3-4FC08D?logo=vue.js)
![License](https://img.shields.io/badge/license-MIT-blue)

---

## 目录

- [项目概览](#项目概览)
- [整体架构](#整体架构)
- [判题链路](#判题链路)
- [AI 链路 (Tool Calling)](#ai-链路-tool-calling)
- [仓库结构](#仓库结构)
- [技术栈](#技术栈)
- [初始配置](#初始配置)
- [快速启动](#快速启动)
- [端口总览](#端口总览)
- [默认账号](#默认账号)
- [常用命令](#常用命令)
- [相关文档](#相关文档)

---

## 项目概览

当前仓库包含三个核心子项目：

| 项目                    | 职责                           | 技术栈                     |
| ----------------------- | ------------------------------ | -------------------------- |
| **AIOJ-main**     | AI 辅助在线判题平台            | Vue 3 + Go 1.21 (Gin/GORM) |
| **remote_judge**  | 独立判题子系统                 | Go 1.25+ (Docker Sandbox)  |
| **agent-service** | AI 微服务 (Tool Calling Agent) | Go 1.21 (DeepSeek/Ollama)  |

核心特性:

- **AIOJ-main** — 用户系统、题目管理（52 道 + 版本控制）、异步判题、学习计划、知识图谱（200+ 节点硬编码）、每日推荐、Rating 系统、AI 能力转发、管理后台
- **remote_judge** — Docker 沙箱执行、代码黑名单检测、多语言支持（C++17/Go/Python3）、容器池复用、熔断降级、gRPC + HTTP 双协议、70+ 测试
- **agent-service** — AI Agent 工具调用（`query_user_problems` / `submit_code` / `retrieve_knowledge`）、RAG 检索增强（600 文档块向量化）、DeepSeek V4（主）+ Ollama（降级）双模型策略

---

## 整体架构

```mermaid
flowchart TB
    subgraph Fused["<b>fused 融合仓库</b>"]
        direction LR
        AIOJ[("<b>AIOJ-main</b><br/>────<br/>AI 在线判题平台<br/>────<br/>Vue 3 前端<br/>Go 1.21 Gin/GORM 后端<br/>────<br/>用户系统 · 题目管理<br/>异步判题 · 知识图谱<br/>Rating · 管理后台")]
        RJ[("<b>remote_judge</b><br/>────<br/>独立判题子系统<br/>────<br/>Go 1.25+<br/>Docker CLI Sandbox<br/>gRPC + JSON Codec<br/>────<br/>黑名单 · 容器池<br/>熔断 · 70+ 测试")]
        AS[("<b>agent-service</b><br/>────<br/>AI Agent 微服务<br/>────<br/>Go 1.21 Gin<br/>DeepSeek V4 (主)<br/>Ollama (降级/Embedding)<br/>────<br/>Tool Calling · RAG<br/>Agent Loop (max 3 rounds)")]
    end

    subgraph Infra["基础设施"]
        MySQL[("MySQL 8.x")]
        RMQ[("RabbitMQ 3.x")]
        Docker[("Docker Desktop<br/>C++17 / Go / Python3")]
        OllamaSrv[("Ollama Server<br/>nomic-embed-text")]
        DeepSeek[("DeepSeek API<br/>deepseek-chat")]
    end

    AIOJ -->|"GORM 读写业务数据"| MySQL
    AIOJ -->|"发布 JudgeTask / 消费判题结果"| RMQ
    AIOJ -->|"gRPC (JSON Codec) 提交判题请求"| RJ
    AIOJ -->|"HTTP 代理转发 AI 请求"| AS
    RJ -->|"Docker CLI 编译 + 运行 + 测量"| Docker
    AS -->|"Chat + Tool Calling"| DeepSeek
    AS -->|"Embedding / 降级 Chat"| OllamaSrv
    AS -->|"Tool 执行: 判题 / 查题"| AIOJ

    style Fused fill:#1a1a2e,stroke:#16213e,color:#e0e0e0
    style AIOJ fill:#0f3460,stroke:#16213e,color:#c8e6c9
    style RJ fill:#0f3460,stroke:#16213e,color:#bbdefb
    style AS fill:#0f3460,stroke:#16213e,color:#f8bbd0
    style MySQL fill:#e8f5e9,stroke:#4caf50
    style RMQ fill:#fff3e0,stroke:#ff9800
    style Docker fill:#e3f2fd,stroke:#2196f3
    style OllamaSrv fill:#f3e5f5,stroke:#9c27b0
    style DeepSeek fill:#fce4ec,stroke:#e91e63
```

---

## 判题链路

```mermaid
sequenceDiagram
    autonumber

    actor User as 用户
    participant API as AIOJ Backend (Gin :8080)
    participant MQ as RabbitMQ
    participant Worker as MQ Worker
    participant DB as MySQL
    participant Judger as remote_judge (gRPC :9090)
    participant Sandbox as Docker Sandbox

    rect rgb(227, 242, 253)
        Note over User,API: ① 接收提交
        User->>+API: POST /api/submissions {problemId, code, language}
        API->>API: JWT 认证 + 参数校验
        API->>DB: INSERT submission (status=Pending)
        API->>MQ: Publish JudgeTask
        API-->>-User: 201 {submissionId, status: Pending}
    end

    rect rgb(255, 243, 224)
        Note over MQ,DB: ② 异步调度
        MQ->>+Worker: Consume JudgeTask
        Worker->>DB: UPDATE status → Queueing
        Worker->>DB: 查询题目详情 + 测试点
        Worker->>DB: UPDATE status → Judging
        Worker->>-Judger: gRPC JudgeRequest
    end

    rect rgb(232, 245, 233)
        Note over Judger,Sandbox: ③ 沙箱判题
        Judger->>Judger: 代码黑名单检测 (AC 自动机)
        Judger->>+Sandbox: Docker CLI 编译 + 运行
        Sandbox-->>-Judger: 运行输出 + 资源用量
        Judger->>Judger: 逐测试点结果比对
        Judger-->>Worker: gRPC JudgeResponse {status, runtimeMs, memoryKb, caseResults[], traceId}
    end

    rect rgb(243, 229, 245)
        Note over Worker,DB: ④ 结果持久化
        Worker->>DB: 写入 verdict + runtime + memory + caseResults
        Worker->>DB: UPDATE status → Finished
    end

    rect rgb(252, 228, 236)
        Note over User,API: ⑤ 查询结果
        User->>+API: GET /api/submissions/:id
        API->>DB: 查询最新结果
        API-->>-User: 200 {status, runtimeMs, caseResults[], ...}
    end
```

---

## AI 链路 (Tool Calling)

agent-service 采用 **Agent 工具调用** 架构，AI 可主动调用工具完成复杂任务：

```mermaid
sequenceDiagram
    autonumber

    actor User as 前端 (Vue 3)
    participant Backend as AIOJ Backend<br/>/api/ai/*
    participant DB as MySQL
    participant Agent as agent-service<br/>/api/agent/chat :8090
    participant RAG as RAG Engine
    participant LLM as DeepSeek V4
    participant OJ as OJ Tool API<br/>/api/agent/*

    rect rgb(227, 242, 253)
        Note over User,Backend: ① 鉴权 + 上下文
        User->>+Backend: POST /api/ai/chat (JWT + mode + problem_ids)
        Backend->>Backend: JWT 鉴权 + 限流
        Backend->>DB: 查询题目信息 + 用户数据
        DB-->>Backend: 结构化上下文
    end

    rect rgb(232, 245, 233)
        Note over Backend,LLM: ② Agent Loop (max 3 轮)
        Backend->>+Agent: POST /api/agent/chat {mode, messages, problem, extraProblems}
        Agent->>RAG: 检索相关 OI-Wiki 知识
        RAG-->>Agent: Top-5 文档块
        Agent->>Agent: 构建 System Prompt + 工具定义

        loop Agent Loop (≤ maxRounds)
            Agent->>+LLM: ChatWithTools (messages + tools)
            LLM-->>-Agent: 回复 / tool_calls
            alt 有 tool_calls
                Agent->>+OJ: POST /api/agent/problems (query_user_problems)<br/>或 POST /api/agent/judge (submit_code)
                OJ->>DB: 查题 / 判题
                OJ-->>-Agent: JSON 结果
                Agent->>Agent: 工具结果追加到 messages
            else 无 tool_calls
                Agent->>Agent: 退出循环
            end
        end
    end

    rect rgb(243, 229, 245)
        Note over Agent,User: ③ 返回
        Agent-->>-Backend: {reply, roundsUsed, mode, provider}
        Backend-->>-User: JSON Response (透传)
    end
```

### 三种工具

| 工具                    | 功能                                     | Agent →                           |
| ----------------------- | ---------------------------------------- | ---------------------------------- |
| `query_user_problems` | 查询用户做题记录（按标签/状态/难度过滤） | HTTP → OJ `/api/agent/problems` |
| `submit_code`         | 提交代码评测，返回判题结果               | HTTP → OJ `/api/agent/judge`    |
| `retrieve_knowledge`  | 从 OI-Wiki 检索算法知识（600 文档块）    | 本地 RAG 引擎                      |

### Mode 映射

| mode                  | 可用工具                                | 最大轮数 | 典型用途                      |
| --------------------- | --------------------------------------- | -------- | ----------------------------- |
| `chat`              | query_user_problems, retrieve_knowledge | 3        | 自由对话，可主动查题/查知识   |
| `code-diagnosis`    | 无                                      | 0        | 纯代码分析，调一次 LLM        |
| `generate-solution` | 无                                      | 0        | 基于代码生成题解              |
| `knowledge-graph`   | query_user_problems                     | 1        | AI 分析用户薄弱点             |
| `study-plan`        | query_user_problems                     | 2        | AI 创建个性化题单             |
| `solve`             | 全部 3 个工具                           | 3        | 解题辅助（hint/explain/full） |

### 请求格式

```json
POST /api/agent/chat
{
  "mode": "solve",
  "user_id": 1,
  "messages": [{"role": "user", "content": "帮我解这道哈希表的题"}],
  "problem": {"id": 101, "title": "两数之和", "content": "...", "tags": ["哈希表"]},
  "extraProblems": [{"id": 102, "title": "最长回文", "tags": ["动态规划"]}],
  "code": "#include <iostream>...",
  "language": "cpp17"
}
```

---

## 仓库结构

```
fused/
│
├── AIOJ-main/                          # AI 辅助在线判题平台
│   ├── backend/                        #   Go 后端 (Gin/GORM)
│   │   ├── cmd/server/                 #     HTTP API 入口
│   │   ├── cmd/judger/                 #     gRPC 判题服务入口
│   │   ├── docker/                     #     MySQL + RabbitMQ Compose
│   │   ├── internal/
│   │   │   ├── data/                   #     ★ 硬编码知识图谱 (新增)
│   │   │   ├── handler/                #     业务处理器 + agent 内部路由
│   │   │   ├── models/                 #     GORM 模型 (精简后)
│   │   │   ├── database/               #     MySQL + 种子数据 (精简后)
│   │   │   └── ...
│   │   ├── API.md                      #     API 契约文档
│   │   └── config.yaml                 #     运行配置
│   ├── frontend/                       #   Vue 3 前端
│   └── README.md
│
├── remote_judge/                       # 独立判题子系统
│   ├── cmd/                            #   server / judger / smoke / stress
│   ├── docker/                         #   判题镜像 + Compose
│   ├── internal/                       #   domain / sandbox / judger / queue / worker
│   └── README.md
│
├── agent-service/                      # AI 微服务 (Tool Calling Agent)
│   ├── cmd/server/                     #   HTTP API 入口
│   ├── internal/
│   │   ├── ai/                         #   AI Client (ChatWithTools)
│   │   ├── handler/                    #   统一 Chat handler + Agent Loop
│   │   ├── tools/                      #   ★ 工具定义 + 执行器 (新增)
│   │   ├── rag/                        #   RAG 检索引擎
│   │   └── config/                     #   环境变量配置
│   ├── oiwiki_docs/                    #   OI-Wiki 文档 (600 个向量块)
│   ├── .env                            #   配置文件 (需自行创建)
│   └── README.md
│
├── CLAUDE.md                           # 开发指南
├── PROBLEM.md                          # 需求与讨论记录
└── README.md                           # (本文档)
```

---

## 技术栈

### AIOJ-main

| 层级     | 技术                                                                 |
| -------- | -------------------------------------------------------------------- |
| 前端     | Vue 3, Vite, Pinia, Vue Router, Element Plus, Monaco Editor, ECharts |
| 后端     | Go 1.21, Gin, GORM, MySQL, RabbitMQ, gRPC, JWT                       |
| 知识图谱 | 硬编码在 `internal/data/knowledge.go` (200+ 节点, 11 分类)         |

### remote_judge

| 类别 | 工具               |
| ---- | ------------------ |
| 语言 | Go 1.25+           |
| 沙箱 | Docker CLI Sandbox |
| 通信 | gRPC + JSON Codec  |
| 队列 | Memory / RabbitMQ  |
| 仓储 | Memory / MySQL     |

### agent-service

| 类别      | 工具                                                   |
| --------- | ------------------------------------------------------ |
| 语言      | Go 1.21, Gin                                           |
| 主模型    | DeepSeek V4 (OpenAI 兼容 API, 支持 Tool Calling)       |
| 降级模型  | Ollama (qwen2.5-coder:7b, 纯文本, 不支持 Tool Calling) |
| Embedding | Ollama nomic-embed-text:latest                         |
| RAG       | langchaingo 文档分割 + 向量检索                        |
| Agent     | 自定义 Agent Loop (max 3 轮工具调用)                   |

---

## 初始配置

### 1. 前置依赖

确保已安装并运行：

| 依赖           | 安装方式                                                   | 验证命令                        |
| -------------- | ---------------------------------------------------------- | ------------------------------- |
| Go 1.21+       | [go.dev/dl](https://go.dev/dl/)                               | `go version`                  |
| Node.js 18+    | [nodejs.org](https://nodejs.org/)                             | `node --version`              |
| MySQL 8.x      | Docker 或本地安装                                          | `mysql --version`             |
| RabbitMQ 3.x   | Docker 或本地安装                                          | 访问 `http://localhost:15672` |
| Docker Desktop | [docker.com](https://www.docker.com/products/docker-desktop/) | `docker --version`            |
| Ollama         | [ollama.com](https://ollama.com/) (可选)                      | `ollama list`                 |

### 2. 克隆仓库

```bash
git clone https://github.com/Yanxueye/AIOJ.git
cd fused_oj
```

### 3. 配置 agent-service `.env`

在 `agent-service/` 目录下创建 `.env` 文件:

```env
# DeepSeek V4 Flash (OpenAI-compatible)
AI_PROVIDER=openai
OPENAI_API_KEY=sk-your-deepseek-api-key
OPENAI_BASE_URL=https://api.deepseek.com/v1
OPENAI_MODEL=deepseek-chat

# Ollama (降级对话 + RAG embedding)
OLLAMA_URL=http://127.0.0.1:11434
OLLAMA_MODEL=qwen2.5-coder:7b
EMBEDDING_MODEL=nomic-embed-text:latest

# OJ 后端地址 (Tool Calling 用)
OJ_BASE_URL=http://127.0.0.1:8080
```

> **注意**: `OJ_BASE_URL` 是 agent-service 调用 OJ 后端 Tool API 的地址。如果 OJ 后端改了端口，这里也要同步改。

### 4. 配置 AIOJ 后端 `config.yaml`

检查 `AIOJ-main/backend/config.yaml` 关键配置:

```yaml
mysql:
  dsn: "toj:toj_password@tcp(127.0.0.1:3306)/terminaloj?charset=utf8mb4&parseTime=True&loc=Local"
  auto_migrate: true
  seed: true          # 首次启动自动创建种子数据

rabbitmq:
  url: "amqp://guest:guest@127.0.0.1:5672/"
  enabled: true

judger:
  grpc_addr: "127.0.0.1:9090"
  remote: true

ai:
  enabled: true
  endpoint: "http://127.0.0.1:8090/api/agent"
  model: "deepseek-chat"
  timeout_seconds: 180
```

### 5. 拉取 Ollama 模型（可选，用于 RAG）

```bash
ollama pull nomic-embed-text:latest
ollama pull qwen2.5-coder:7b
```

> 如果没有 Ollama 和本地模型，agent-service 会用 DeepSeek API 做对话，RAG 检索降级为关键词匹配。

### 6. 安装前端依赖

```bash
cd AIOJ-main/frontend
npm install
```

---

## 快速启动

按顺序启动各服务（每个命令在独立终端中运行）：

### 1. 基础设施 (MySQL + RabbitMQ)

```cmd
cd AIOJ-main\backend
docker compose -f docker\docker-compose.yml up -d mysql rabbitmq
```

### 2. 构建判题沙箱镜像

```cmd
cd remote_judge
docker build -t remote-judge-cpp17 -f docker\images\cpp17\Dockerfile .
docker build -t remote-judge-go122 -f docker\images\go1.22\Dockerfile .
docker build -t remote-judge-python311 -f docker\images\python3.11\Dockerfile .
```

### 3. remote_judge gRPC 判题服务

```cmd
cd remote_judge
set REMOTE_JUDGE_GRPC_ADDR=127.0.0.1:9090
go run .\cmd\judger
```

### 4. agent-service

```cmd
cd agent-service
go run .\cmd\server
```

### 5. AIOJ 后端

```cmd
cd AIOJ-main\backend
go run .\cmd\server -config config.yaml
```

### 6. AIOJ 前端

```cmd
cd AIOJ-main\frontend
npm run dev
```

前端访问 `http://localhost:5173`，后端 API 在 `http://localhost:8080`。

---

## 端口总览

| 服务              | 默认端口 | 说明                 |
| ----------------- | -------- | -------------------- |
| AIOJ 前端         | :5173    | Vite 开发服务器      |
| AIOJ 后端         | :8080    | Gin HTTP API         |
| remote_judge gRPC | :9090    | 判题服务             |
| agent-service     | :8090    | AI Agent 微服务      |
| MySQL             | :3306    | 数据库               |
| RabbitMQ          | :5672    | 消息队列             |
| RabbitMQ 管理     | :15672   | Web UI (guest/guest) |

---

## 默认账号

| 角色     | 用户名         | 密码       |
| -------- | -------------- | ---------- |
| 普通用户 | `coder_test` | `123456` |
| 管理员   | `admin`      | `123456` |

---

## 常用命令

| 操作                   | 命令                                                |
| ---------------------- | --------------------------------------------------- |
| AIOJ 后端测试          | `cd AIOJ-main\backend && go test ./...`           |
| remote_judge 测试      | `cd remote_judge && go test ./...`                |
| AIOJ 前端构建          | `cd AIOJ-main\frontend && npm run build`          |
| agent-service 构建     | `cd agent-service && go build ./...`              |
| agent-service RAG 状态 | `curl http://127.0.0.1:8090/api/agent/rag-status` |

---

## 相关文档

| 文档                                                  | 说明                        |
| ----------------------------------------------------- | --------------------------- |
| [AIOJ-main/README.md](AIOJ-main/README.md)               | AIOJ 项目说明               |
| [AIOJ-main/backend/API.md](AIOJ-main/backend/API.md)     | 后端 API 契约               |
| [remote_judge/README.md](remote_judge/README.md)         | remote_judge 项目说明       |
| [agent-service/README.md](agent-service/README.md)       | agent-service 项目说明      |
| [CLAUDE.md](CLAUDE.md)                                   | 开发指南 (命令、架构、端口) |
| [remote_judge/docs/check.md](remote_judge/docs/check.md) | remote_judge 测试引导       |
