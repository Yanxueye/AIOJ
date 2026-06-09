# TerminalOJ 后端 API 文档

> 基础路径：`/api`  
> 监听端口：`8080`（可在 `config.yaml` 中修改）  
> 认证方式：`Authorization: Bearer <jwt>`  
> 统一响应信封：`{ "code": 0, "message": "ok", "data": {} }`，`code != 0` 表示业务失败

---

## 1. 认证

### `POST /api/auth/register`

注册新用户。密码服务端使用 `bcrypt` 存储。

请求体：
```json
{ "username": "alice", "email": "alice@toj.com", "password": "secret123" }
```
字段规则：用户名 3–20 字符；邮箱符合 `RFC 5322` 简化正则；密码 ≥ 6 位。

### `POST /api/auth/login`

返回 `token` + 完整 `Profile`（与 `GET /user/profile` 兼容字段集）。

```json
{
  "code": 0, "message": "ok",
  "data": {
    "token": "eyJhbGciOi...",
    "user": {
      "id": 1, "username": "coder_test", "rating": 1520,
      "solvedCount": 28, "totalSubmissions": 65, "acceptRate": "43.1",
      "rank": 42, "registeredAt": "2026-03-15"
    }
  }
}
```

登录失败返回 HTTP 400 + `{"code":-1,"message":"用户名或密码错误"}`。

---

## 2. 用户

| 方法 | 路径 | 认证 | 说明 |
|------|------|------|------|
| GET | `/api/user/profile` | ✅ | 返回扩展 Profile（含难度 / 算法统计、近 14 天做题活跃度） |
| PUT | `/api/user/profile` | ✅ | 仅接收 `email` / `bio`，bio 最长 200 字 |

扩展字段：
```json
{
  "solvedByDifficulty": { "简单": 15, "中等": 10, "困难": 3 },
  "solvedByAlgorithm":  { "动态规划": 8, "贪心": 5, "图论": 3 },
  "recentActivity":     [{ "date": "2026-04-06", "count": 3 }]
}
```

---

## 3. 题目 & 公告

| 方法 | 路径 | 认证 | 说明 |
|------|------|------|------|
| GET | `/api/problems` | 可选 | 列表 + 分页 + 搜索。带 Token 时返回 `accepted` 标记 |
| GET | `/api/problems/:id` | ✅ | 题目详情（含 Markdown 正文） |
| GET | `/api/announcements` | ❌ | 返回最近 20 条公告 |

`GET /api/problems` 查询参数：

| 参数 | 类型 | 说明 |
|------|------|------|
| `page` | int | 页码，默认 1 |
| `pageSize` | int | 每页数量，默认 20，上限 100 |
| `keyword` | string | 题号或标题模糊匹配 |
| `difficulty` | string | `简单` / `中等` / `困难` |
| `tag` | string | 算法标签；通过 MySQL `JSON_CONTAINS` 过滤 |

---

## 4. 提交评测

### `POST /api/submissions`  ✅

流程（见 `PROGRESS.md` 数据通路图）：

1. 校验请求体 + 语言枚举 + 题目存在性
2. 命中 **令牌桶限流**（默认 12 次 / 分钟，突发 3）
3. 生成 `submissionId` 并 **发布** 到 RabbitMQ `toj.submit` 队列
4. 立即返回 `Pending`，前端轮询 `/api/submissions/:id` 获取最终结果

请求体：
```json
{ "problemId": 1001, "language": "cpp", "code": "int main(){...}" }
```

响应：
```json
{
  "code": 0, "message": "ok",
  "data": {
    "id": 100123,
    "problemId": 1001,
    "status": "Pending",
    "language": "cpp",
    "runtime": 0, "memory": "0.0",
    "codeLength": 512,
    "createdAt": "2026-04-20T14:05:12.345Z"
  }
}
```

返回 HTTP 429 + `{"code":-1,"message":"提交过于频繁，请稍后再试"}` 表示限流触发。

### `GET /api/submissions` ✅

查询参数：

| 参数 | 说明 |
|------|------|
| `page`, `pageSize` | 分页 |
| `problemId` | 按题号筛选 |
| `status` | 按评测状态精确匹配 |
| `sortBy` | `time`（默认，按 `id DESC`）或 `problemId` |

响应：`{ list: [...], total: number }`，结构与前端 `SubmissionStatus.vue` 对齐。

### `GET /api/submissions/:id` ✅

返回单条提交；越权访问返回 404。

---

## 5. AI 模块

AI 接口分为两层：

- TerminalOJ 对前端暴露 `/api/ai/*`，统一走 JWT 鉴权和 `{code,message,data}` 响应信封。
- TerminalOJ 对外部封装好的 AI 服务发起 Pipeline 调用。`config.yaml` 中 `ai.enabled=false` 时使用本地 Mock；开启时请求 `ai.endpoint`。

后端调用外部 AI 服务时使用以下 envelope：

```json
{
  "task": "chat | code_diagnosis | knowledge_graph | solve",
  "model": "gpt-4o-mini",
  "payload": {}
}
```

外部服务可直接返回业务 JSON，也可返回 `{ "code": 0, "message": "ok", "data": {...} }`。

| 方法 | 路径 | 认证 | 说明 |
|------|------|------|------|
| POST | `/api/ai/chat` | ✅ | 多轮 AI 对话；支持题目上下文，会落库用户和助手消息 |
| GET | `/api/ai/history` | ✅ | 最近 50 个会话元数据 |
| GET | `/api/ai/conversations/:id/messages` | ✅ | 指定会话的完整消息流 |
| POST | `/api/ai/code-diagnosis` | ✅ | AI 辅助代码错误诊断 |
| POST | `/api/ai/knowledge-graph` | ✅ | 基于题目和最近提交生成学习知识图谱 |
| POST | `/api/ai/solve` | ✅ | 针对题目的提示 / 讲解 / 完整解法入口 |

### `POST /api/ai/chat`

请求体：

```json
{
  "message": "讲讲动态规划",
  "history": [{ "role": "user", "content": "..." }],
  "problem_id": 1001,
  "conversation_id": ""
}
```

响应：

```json
{
  "reply": "Markdown 格式回复",
  "conversationId": "uuid",
  "provider": "mock | external",
  "metadata": {}
}
```

### `GET /api/ai/history`

响应：

```json
{
  "conversations": [
    {
      "id": "uuid",
      "title": "关于动态规划的讨论",
      "problemId": 1001,
      "createdAt": "2026-05-04T10:00:00.000Z",
      "messageCount": 6
    }
  ]
}
```

### `GET /api/ai/conversations/:id/messages`

响应：

```json
{
  "conversation": { "id": "uuid", "title": "...", "problemId": 1001, "createdAt": "..." },
  "messages": [
    { "id": 1, "role": "user", "content": "...", "createdAt": "..." },
    { "id": 2, "role": "assistant", "content": "...", "createdAt": "..." }
  ]
}
```

### `POST /api/ai/code-diagnosis`

支持直接传代码，也支持传 `submissionId` 后由后端读取用户自己的提交代码。

请求体：

```json
{
  "problemId": 1001,
  "submissionId": 100123,
  "language": "cpp",
  "code": "int main(){...}",
  "judgeStatus": "Wrong Answer",
  "errorMessage": "case 3 failed"
}
```

响应：

```json
{
  "summary": "诊断摘要",
  "issues": [
    { "line": 12, "severity": "error", "message": "数组越界", "hint": "检查循环边界" }
  ],
  "suggestions": ["补充边界用例"],
  "fixedCode": "可选修正代码",
  "rawMarkdown": "可直接渲染的 Markdown",
  "provider": "mock | external"
}
```

### `POST /api/ai/knowledge-graph`

请求体：

```json
{ "problemId": 1001, "scope": "recent" }
```

`problemId` 可空；为空时按用户最近提交整理。后端默认带最近 50 条提交摘要给 AI 管线。

响应：

```json
{
  "summary": "图谱摘要",
  "nodes": [{ "id": "tag:动态规划", "label": "动态规划", "type": "algorithm", "weight": 8 }],
  "edges": [{ "source": "user", "target": "tag:动态规划", "type": "strong_at", "weight": 8 }],
  "rawMarkdown": "可直接渲染的 Markdown",
  "provider": "mock | external"
}
```

### `POST /api/ai/solve`

请求体：

```json
{ "problemId": 1001, "question": "我不理解状态转移", "level": "hint" }
```

`level` 取值：`hint` / `explain` / `full`。

响应：

```json
{
  "answer": "Markdown 格式解题说明",
  "hints": ["先手算样例"],
  "complexity": "O(n log n)",
  "provider": "mock | external"
}
```

---
## 6. 错误码约定

| HTTP | code | 含义 |
|------|------|------|
| 200 | 0 | 成功 |
| 400 | -1 | 业务校验失败 |
| 401 | -1 | 未登录 / Token 过期 |
| 403 | -1 | 权限不足 |
| 404 | -1 | 资源不存在 |
| 429 | -1 | 限流触发（仅 `POST /submissions`） |
| 500 | -1 | 服务端异常 |

---

## 7. 判题 gRPC 契约

见 `proto/judger.proto`。服务路径：`/judger.Judger/Judge`。请求 / 响应结构：

```text
JudgeRequest  = { submission_id, problem_id, language, code, time_limit_ms, memory_limit_mb, test_cases[] }
JudgeResponse = { submission_id, status, runtime_ms, memory_mb, error_message }
```

后端 → 判题容器调用示意：

```go
resp, err := judger.NewClient("127.0.0.1:9090", 15).Judge(ctx, &JudgeRequest{...})
```

传输层使用 `google.golang.org/grpc` + 自定义 `json` Codec（`internal/judger/codec.go`），无需 `protoc-gen-go`/`protoc-gen-go-grpc`。其他语言客户端可直接按 `judger.proto` 生成标准 protobuf 绑定，然后连上同一端口（Go 服务端的 JSON Codec 不妨碍 proto 客户端连入，只需额外注册 proto codec）。

---

## 8. 联调指南

1. `docker compose -f backend/docker/docker-compose.yml up -d mysql rabbitmq`
2. `cd backend && go mod tidy && go run ./cmd/judger` （另开一个终端）
3. `go run ./cmd/server -config config.yaml`
4. 默认账号：`coder_test` / `123456`；题库已自动 Seed 5 道题
5. 前端将 `src/api/index.js` 中的 `USE_MOCK` 设为 `false`，`baseURL` 指向 `http://localhost:8080`

