# agent-service

AI 微服务，为 AIOJ 平台提供大模型驱动的智能能力。采用 MIMO API 作为主模型，本地 Ollama 作为降级策略。

## 功能

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/agent/health` | GET | 健康检查（含 AI 服务状态） |
| `/api/agent/hint` | POST | 做题错误时给出提示（不泄露答案） |
| `/api/agent/analyze` | POST | 通过后代码分析（复杂度、知识点、优化建议） |
| `/api/agent/generate-solution` | POST | 根据提交历史辅助生成题解 |
| `/api/agent/chat` | POST | 通用 AI 对话（支持上下文注入） |

## 架构

```
AIOJ 后端 (HTTP) ──转发──▶ agent-service (:8090)
                                │
                    ┌───────────┴───────────┐
                    ▼                       ▼
              MIMO API (主)           Ollama (降级)
        mimo-v2.5-pro             qwen2.5-coder:7b
```

**降级策略**：优先调用 MIMO API → 失败时自动降级到本地 Ollama。

## 环境要求

- Go 1.21+
- MIMO API Key（主模型）
- Ollama（可选，降级模型）

## 快速启动

### 1. 创建配置文件

在 `agent-service/` 目录下创建 `.env` 文件：

```env
# Agent Service Configuration

# HTTP server address
AGENT_HTTP_ADDR=:8090

# Primary AI: MIMO API (OpenAI-compatible)
AI_PROVIDER=openai
OPENAI_API_KEY=your-api-key-here
OPENAI_BASE_URL=https://token-plan-sgp.xiaomimimo.com/v1
OPENAI_MODEL=mimo-v2.5-pro

# Fallback AI: Local Ollama
OLLAMA_URL=http://127.0.0.1:11434
OLLAMA_MODEL=qwen2.5-coder:7b

# Judge service
AIOJ_BACKEND_URL=http://127.0.0.1:8080
```

### 2. 启动服务

```cmd
cd agent-service
go run .\cmd\server
```

默认监听 `http://127.0.0.1:8090`。

### 3. 验证

```bash
curl http://127.0.0.1:8090/api/agent/health
```

## 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `AGENT_HTTP_ADDR` | `:8090` | HTTP 监听地址 |
| `AI_PROVIDER` | `openai` | 优先使用的 AI 提供商（`openai` / `ollama`） |
| `OPENAI_API_KEY` | — | MIMO API Key |
| `OPENAI_BASE_URL` | `https://token-plan-sgp.xiaomimimo.com/v1` | OpenAI 兼容接口地址 |
| `OPENAI_MODEL` | `mimo-v2.5-pro` | 主模型名称 |
| `OLLAMA_URL` | `http://127.0.0.1:11434` | Ollama 服务地址 |
| `OLLAMA_MODEL` | `qwen2.5-coder:7b` | 降级模型名称 |
| `AIOJ_BACKEND_URL` | `http://127.0.0.1:8080` | AIOJ 后端地址（用于调用判题服务） |

## 目录结构

```
agent-service/
├─ cmd/server/          HTTP API 入口
├─ internal/
│  ├─ ai/              AI 客户端
│  │  ├─ client.go     统一客户端（自动降级）
│  │  ├─ openai.go     OpenAI 兼容 API 客户端（MIMO）
│  │  └─ ollama.go     Ollama 客户端
│  ├─ config/          环境变量配置 + .env 加载
│  ├─ handler/         HTTP 处理器
│  ├─ judge/           AIOJ 后端判题客户端
│  └─ rag/             RAG 检索逻辑（待实现）
├─ .env                配置文件（含 API Key，不提交 git）
└─ go.mod
```

## 与 AIOJ 后端的集成

agent-service 的 AI 能力通过 AIOJ 后端代理，前端不直接调用 agent-service。AIOJ 后端负责：

1. 转发 AI 请求到 agent-service
2. 注入用户上下文（用户 ID、题目信息、提交记录）
3. 返回结果给前端

## 开发

```cmd
cd agent-service
go build ./...          # 编译
go run .\cmd\server     # 启动
```
