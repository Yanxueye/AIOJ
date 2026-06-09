# TerminalOJ 前端 API 接口文档

> 基础路径：`/api`  
> 认证方式：JWT Bearer Token（`Authorization: Bearer <token>`）  
> 响应格式：`{ "code": 0, "message": "ok", "data": {...} }`

---

## 1. 认证模块

### 1.1 用户登录

- **POST** `/api/auth/login`
- **描述**：用户登录，获取 JWT Token
- **请求体**：
  ```json
  {
    "username": "string",
    "password": "string"
  }
  ```
- **响应**：
  ```json
  {
    "code": 0,
    "message": "ok",
    "data": {
      "token": "jwt_token_string",
      "user": {
        "id": 1,
        "username": "coder_test",
        "email": "test@terminaloj.com",
        "avatar": "",
        "bio": "热爱算法的开发者",
        "rating": 1520,
        "rank": 42,
        "solvedCount": 28,
        "totalSubmissions": 65,
        "acceptRate": "43.1",
        "registeredAt": "2026-03-15"
      }
    }
  }
  ```

### 1.2 用户注册

- **POST** `/api/auth/register`
- **描述**：注册新用户
- **请求体**：
  ```json
  {
    "username": "string (3-20字符)",
    "email": "string (合法邮箱)",
    "password": "string (至少6位)"
  }
  ```
- **响应**：
  ```json
  {
    "code": 0,
    "message": "ok",
    "data": {
      "message": "注册成功"
    }
  }
  ```

---

## 2. 用户模块

### 2.1 获取个人信息

- **GET** `/api/user/profile`
- **认证**：需要
- **响应**：
  ```json
  {
    "code": 0,
    "message": "ok",
    "data": {
      "id": 1,
      "username": "coder_test",
      "email": "test@terminaloj.com",
      "avatar": "",
      "bio": "热爱算法的开发者",
      "rating": 1520,
      "rank": 42,
      "solvedCount": 28,
      "totalSubmissions": 65,
      "acceptRate": "43.1",
      "registeredAt": "2026-03-15",
      "solvedByDifficulty": {
        "简单": 15,
        "中等": 10,
        "困难": 3
      },
      "solvedByAlgorithm": {
        "动态规划": 8,
        "贪心": 5,
        "搜索": 4,
        "图论": 3,
        "数学": 3,
        "字符串": 2,
        "数据结构": 2,
        "模拟": 1
      },
      "recentActivity": [
        { "date": "2026-04-06", "count": 3 },
        { "date": "2026-04-05", "count": 5 }
      ]
    }
  }
  ```

### 2.2 更新个人信息

- **PUT** `/api/user/profile`
- **认证**：需要
- **请求体**：
  ```json
  {
    "email": "string (可选)",
    "bio": "string (可选, 最大200字符)"
  }
  ```
- **响应**：返回更新后的用户对象（同 2.1）

---

## 3. 题目模块

### 3.1 获取题目列表

- **GET** `/api/problems`
- **查询参数**：
  | 参数 | 类型 | 必填 | 说明 |
  |------|------|------|------|
  | page | int | 否 | 页码，默认 1 |
  | pageSize | int | 否 | 每页数量，默认 20 |
  | keyword | string | 否 | 搜索关键字（题号或题目名称） |
  | difficulty | string | 否 | 难度筛选：简单/中等/困难 |
  | tag | string | 否 | 算法标签筛选 |
- **响应**：
  ```json
  {
    "code": 0,
    "message": "ok",
    "data": {
      "list": [
        {
          "id": 1001,
          "title": "两数之和",
          "difficulty": "简单",
          "difficultyScore": 800,
          "tags": ["动态规划", "数学"],
          "acceptRate": "72.3",
          "submitCount": 3421,
          "accepted": true
        }
      ],
      "total": 50
    }
  }
  ```

### 3.2 获取题目详情

- **GET** `/api/problems/:id`
- **认证**：需要
- **路径参数**：`id` - 题目ID
- **响应**：
  ```json
  {
    "code": 0,
    "message": "ok",
    "data": {
      "id": 1001,
      "title": "两数之和",
      "difficulty": "简单",
      "difficultyScore": 800,
      "tags": ["动态规划", "数学"],
      "acceptRate": "72.3",
      "submitCount": 3421,
      "accepted": true,
      "content": "Markdown 格式的题目描述（支持 LaTeX）",
      "timeLimit": 1000,
      "memoryLimit": 256,
      "source": "TerminalOJ 原创题目"
    }
  }
  ```

### 3.3 获取公告列表

- **GET** `/api/announcements`
- **响应**：
  ```json
  {
    "code": 0,
    "message": "ok",
    "data": [
      {
        "id": 1,
        "title": "公告标题",
        "content": "公告内容",
        "date": "2026-04-06",
        "type": "success | info | warning | primary"
      }
    ]
  }
  ```

---

## 4. 提交评测模块

### 4.1 提交代码

- **POST** `/api/submissions`
- **认证**：需要
- **请求体**：
  ```json
  {
    "problemId": 1001,
    "language": "cpp | java | python | go",
    "code": "string (源代码)"
  }
  ```
- **响应**：
  ```json
  {
    "code": 0,
    "message": "ok",
    "data": {
      "id": 100001,
      "problemId": 1001,
      "status": "Accepted | Wrong Answer | Time Limit Exceeded | Runtime Error | Compilation Error | Pending",
      "language": "cpp",
      "runtime": 42,
      "memory": "3.2",
      "createdAt": "2026-04-06T10:30:00.000Z"
    }
  }
  ```

### 4.2 获取提交列表

- **GET** `/api/submissions`
- **认证**：需要
- **查询参数**：
  | 参数 | 类型 | 必填 | 说明 |
  |------|------|------|------|
  | page | int | 否 | 页码，默认 1 |
  | pageSize | int | 否 | 每页数量，默认 20 |
  | problemId | int | 否 | 按题号筛选 |
  | status | string | 否 | 按评测状态筛选 |
  | sortBy | string | 否 | 排序方式：time(默认) / problemId |
- **响应**：
  ```json
  {
    "code": 0,
    "message": "ok",
    "data": {
      "list": [
        {
          "id": 100001,
          "problemId": 1001,
          "problemTitle": "两数之和",
          "status": "Accepted",
          "language": "cpp",
          "runtime": 42,
          "memory": "3.2",
          "createdAt": "2026-04-06T10:30:00.000Z",
          "codeLength": 512
        }
      ],
      "total": 80
    }
  }
  ```

### 4.3 获取提交详情

- **GET** `/api/submissions/:id`
- **认证**：需要
- **路径参数**：`id` - 提交ID
- **响应**：返回单个提交对象（同列表中的单项）

---

## 5. AI 模块

### 5.1 AI 对话

- **POST** `/api/ai/chat`
- **认证**：需要
- **描述**：发送消息给 AI 助手，可选择关联题目上下文；后端会返回并维护 `conversationId`
- **请求体**：
  ```json
  {
    "message": "string (用户消息)",
    "history": [
      { "role": "user | assistant | system", "content": "string" }
    ],
    "problem_id": 1001,
    "conversation_id": "uuid 或空字符串"
  }
  ```
- **响应**：
  ```json
  {
    "code": 0,
    "message": "ok",
    "data": {
      "reply": "Markdown 格式的 AI 回复（支持 LaTeX 和代码块）",
      "conversationId": "uuid",
      "provider": "mock | external",
      "metadata": {}
    }
  }
  ```

### 5.2 获取历史对话

- **GET** `/api/ai/history`
- **认证**：需要
- **响应**：
  ```json
  {
    "code": 0,
    "message": "ok",
    "data": {
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
  }
  ```

### 5.3 获取会话消息

- **GET** `/api/ai/conversations/:id/messages`
- **认证**：需要
- **响应**：
  ```json
  {
    "code": 0,
    "message": "ok",
    "data": {
      "conversation": { "id": "uuid", "title": "...", "problemId": 1001, "createdAt": "..." },
      "messages": [
        { "id": 1, "role": "user", "content": "...", "createdAt": "..." },
        { "id": 2, "role": "assistant", "content": "...", "createdAt": "..." }
      ]
    }
  }
  ```

### 5.4 代码诊断

- **POST** `/api/ai/code-diagnosis`
- **认证**：需要
- **描述**：在题目页点击「诊断代码」时调用，后端会把题目上下文、语言、代码和可选评测错误传给 AI 管线
- **请求体**：
  ```json
  {
    "problemId": 1001,
    "submissionId": 100123,
    "language": "cpp | java | python | go",
    "code": "string",
    "judgeStatus": "Wrong Answer",
    "errorMessage": "case 3 failed"
  }
  ```
- **响应**：
  ```json
  {
    "code": 0,
    "message": "ok",
    "data": {
      "summary": "诊断摘要",
      "issues": [
        { "line": 12, "severity": "error", "message": "数组越界", "hint": "检查循环边界" }
      ],
      "suggestions": ["补充边界用例"],
      "fixedCode": "可选修正代码",
      "rawMarkdown": "可直接渲染的 Markdown",
      "provider": "mock | external"
    }
  }
  ```

### 5.5 学习知识图谱

- **POST** `/api/ai/knowledge-graph`
- **认证**：需要
- **描述**：AI 训练页点击「整理我的知识图谱」时调用，后端会附带最近提交摘要
- **请求体**：
  ```json
  { "problemId": 1001, "scope": "recent" }
  ```
- **响应**：
  ```json
  {
    "code": 0,
    "message": "ok",
    "data": {
      "summary": "图谱摘要",
      "nodes": [{ "id": "tag:动态规划", "label": "动态规划", "type": "algorithm", "weight": 8 }],
      "edges": [{ "source": "user", "target": "tag:动态规划", "type": "strong_at", "weight": 8 }],
      "rawMarkdown": "可直接渲染的 Markdown",
      "provider": "mock | external"
    }
  }
  ```

### 5.6 解题辅助

- **POST** `/api/ai/solve`
- **认证**：需要
- **描述**：题目页点击「解题提示」时调用；也可传具体问题给 AI
- **请求体**：
  ```json
  { "problemId": 1001, "question": "我不理解状态转移", "level": "hint" }
  ```
  > `level`：`hint` / `explain` / `full`
- **响应**：
  ```json
  {
    "code": 0,
    "message": "ok",
    "data": {
      "answer": "Markdown 格式解题说明",
      "hints": ["先手算样例"],
      "complexity": "O(n log n)",
      "provider": "mock | external"
    }
  }
  ```

---
## 错误响应格式

```json
{
  "code": -1,
  "message": "错误描述",
  "data": null
}
```

| HTTP 状态码 | 说明 |
|-------------|------|
| 200 | 成功 |
| 400 | 请求参数错误 |
| 401 | 未认证或 Token 过期 |
| 403 | 权限不足 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

---

## Mock 模式

前端内置了完整的 Mock 数据层，在 `src/api/index.js` 中设置 `USE_MOCK = true` 即可启用。Mock 模式下所有 API 请求由前端本地拦截处理，无需启动后端服务。

切换到真实后端时，将 `USE_MOCK` 设为 `false`，确保后端服务运行在 `http://localhost:8080` 即可。

