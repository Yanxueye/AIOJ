# TerminalOJ AI 接口实现记录

> 最后更新：2026-05-04  
> 目标：按 `worker.skill` 的 AI 业务要求，完成 AI 能力的调用层、Pipeline 层、前后端入口和文档。

## 完成内容

| 能力 | 状态 | 落地位置 |
|------|------|----------|
| AI 聊天 / 问题 / 解题 | 已完成 | `backend/internal/handler/ai.go`、`backend/internal/ai/client.go`、`frontend/src/components/AIChat.vue` |
| AI 辅助代码鉴别错误 | 已完成 | `POST /api/ai/code-diagnosis`，题目页「诊断代码」按钮 |
| AI 整理做题信息 / 知识图谱构建 | 已完成 | `POST /api/ai/knowledge-graph`，AI 训练页「整理我的知识图谱」按钮 |
| 会话记录与消息流 | 已完成 | `conversations/messages` 表复用；新增 `GET /api/ai/conversations/:id/messages` |
| 外部 AI 服务 Pipeline 调用 | 已完成 | `internal/ai.Client` 按 `task + model + payload` envelope 调外部服务 |
| Mock 降级 | 已完成 | `ai.enabled=false` 时本地确定性 Mock，前端 `USE_MOCK=true` 时本地 Mock |
| API 文档 | 已完成 | `backend/API.md`、`frontend/API.md` |

## 后端实现

新增 `backend/internal/ai/client.go`，封装四类 AI 任务：

- `Chat`：多轮对话，支持题目上下文。
- `DiagnoseCode`：代码诊断，传入题目、语言、代码、评测状态和错误信息。
- `BuildKnowledgeGraph`：读取最近提交摘要，生成节点 / 边形式的知识图谱。
- `Solve`：题目提示、讲解、完整解法的统一入口。

配置项位于 `backend/config.yaml`：

```yaml
ai:
  enabled: false
  endpoint: "http://127.0.0.1:18080"
  api_key: ""
  model: "gpt-4o-mini"
  timeout_seconds: 20
```

`enabled=false` 或 `endpoint` 为空时，不访问外部服务，直接走 Mock。开启外部服务后，后端会请求：

- `POST {endpoint}/chat`
- `POST {endpoint}/code-diagnosis`
- `POST {endpoint}/knowledge-graph`
- `POST {endpoint}/solve`

请求体统一为：

```json
{
  "task": "chat | code_diagnosis | knowledge_graph | solve",
  "model": "gpt-4o-mini",
  "payload": {}
}
```

外部服务响应可直接返回业务对象，也可返回 `{ "code": 0, "message": "ok", "data": {...} }`。

## 前端接线

新增或扩展的前端调用：

- `frontend/src/api/ai.js`：新增 `diagnoseCode`、`buildKnowledgeGraph`、`solveProblem`、`getMessages`。
- `frontend/src/stores/ai.js`：维护 `conversationId`，封装诊断 / 解题 / 图谱动作，并把结果写入当前消息流。
- `frontend/src/components/AIChat.vue`：题目上下文模式下新增「解题提示」「诊断代码」按钮。
- `frontend/src/views/ProblemDetail.vue`：把当前代码和语言传给 AI 面板。
- `frontend/src/views/AITraining.vue`：新增「整理我的知识图谱」入口。
- `frontend/src/api/mock.js`：补齐新增 AI 接口的前端 Mock。

## 接口清单

| 方法 | 路径 | 说明 |
|------|------|------|
| `POST` | `/api/ai/chat` | 聊天 / 问题 |
| `GET` | `/api/ai/history` | 会话元数据 |
| `GET` | `/api/ai/conversations/:id/messages` | 会话消息流 |
| `POST` | `/api/ai/code-diagnosis` | 代码错误诊断 |
| `POST` | `/api/ai/knowledge-graph` | 学习知识图谱 |
| `POST` | `/api/ai/solve` | 解题辅助 |

详细字段见 `backend/API.md` 和 `frontend/API.md`。

## 调试结果

已执行：

```powershell
cd backend && go mod tidy && go test ./...
cd frontend && npm run build
```

结果：

- 后端 `go test ./...` 通过。
- 前端 `npm run build` 通过。
- Vite 仍提示 Monaco / Element Plus / AIChat chunk 超过 1000KB，这是项目已有的大依赖分包体积提示，不影响构建产物生成。

## 后续对接点

- 外部 AI 服务只需要实现四个 HTTP 端点，并按上面的 envelope 读取 `task/model/payload`。
- 若后续要做流式输出，可以在现有 `Chat` 旁新增 SSE 路由，不破坏当前 JSON 接口。
- 知识图谱当前返回节点 / 边数据和 Markdown 摘要，前端后续可接 ECharts graph 或 D3 做可视化。
