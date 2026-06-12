# Agent-Service 设计文档

> 最后更新：2026-06-12

---

## 一、架构定位

### 1.1 服务关系

```
┌──────────┐     HTTP      ┌──────────────┐     HTTP      ┌───────────────┐
│  前端     │ ───────────→ │  OJ 后端     │ ───────────→ │  agent-service │
│ (Vue 3)  │ ←─────────── │  (Gin/GORM)  │ ←─────────── │  (Gin)         │
└──────────┘               └──────────────┘               └───────────────┘
                                  │                              │
                                  │ gRPC                         │ HTTP
                                  ↓                              ↓
                           ┌──────────────┐               ┌───────────────┐
                           │ remote_judge  │               │  LLM Provider │
                           │  (判题沙箱)   │               │  (MIMO/Ollama)│
                           └──────────────┘               └───────────────┘
```

### 1.2 核心原则

- **用户不可感知**：agent-service 不直接暴露给前端，所有请求经 OJ 后端转发
- **职责分离**：OJ 后端负责数据整理、上下文组装、持久化；agent-service 负责 AI 推理
- **统一标签字典**：OJ 后端与 agent-service 共用同一套算法标签（AlgorithmTag），确保 AI 输出的标签与题库一致
- **RAG 增强**：agent-service 内置 RAG 检索，自动注入相关知识到 AI prompt

---

## 二、统一标签字典

### 2.1 设计

OJ 后端的 `algorithm_tags` 表作为唯一标签字典，包含 90+ 标签，按分类分组：

| 分类 | 示例标签 |
|------|----------|
| 基础算法 | 模拟、枚举、贪心、排序、二分、双指针、前缀和、位运算、哈希 |
| 动态规划 | 背包、区间DP、树形DP、状压DP、数位DP、LIS、LCS |
| 图论 | BFS、DFS、最短路径、最小生成树、拓扑排序、二分图、并查集 |
| 数据结构 | 栈、队列、堆、链表、树、线段树、树状数组、字典树 |
| 字符串 | KMP、字符串哈希、Manacher、后缀自动机、后缀数组 |
| 数学 | 数论、质数、GCD/LCM、快速幂、组合数学、NTT、博弈论 |
| 搜索 | 回溯、剪枝、迭代加深、A*、双向BFS |
| 计算几何 | 凸包、半平面交、最近点对、旋转卡壳 |

### 2.2 共享方式

- OJ 后端提供 `GET /api/tags` 和 `GET /api/tags/names` API
- agent-service 启动时从 OJ 后端拉取标签列表，缓存到内存
- AI prompt 中注入标签列表，要求 AI 输出的算法标签必须来自该列表
- 前端题目创建时使用下拉多选框选择标签

---

## 三、RAG 系统

### 3.1 数据来源

| 来源 | 存储方式 | 更新频率 | 选取上限 |
|------|----------|----------|----------|
| OI-Wiki 文档 | 本地 markdown 文件，langchaingo 加载分割 | 手动运行爬虫 | top-3 |
| 高赞用户题解 | MySQL `problem_solutions` 表 | 实时 | top-2 |

### 3.2 检索流程

```
用户提问/代码 → 构建查询 → 检索 OI-Wiki (top-3) + 检索题解 (top-2) → 合并注入 prompt
```

### 3.3 标签关联

- OI-Wiki 文档的 metadata 中包含 `category` 字段，与标签字典的分类对齐
- 检索时优先匹配与题目标签相同分类的知识
- 题解检索时按题目标签筛选

---

## 四、AI 功能详细设计

### 4.1 代码诊断（Code Diagnosis）

**触发场景**：用户提交代码判题失败后，点击"获取诊断"

**数据流**：
```
前端 → OJ 后端 → agent-service → LLM → agent-service → OJ 后端 → 前端
```

**OJ 后端组装的信息**：
```json
{
  "problemId": 1001,
  "problemTitle": "两数之和",
  "problemContent": "题面内容...",
  "problemEditorial": "官方题解（如果有）...",
  "samples": [{"input": "...", "expected": "..."}],
  "algorithmTags": ["数组", "哈希表"],
  "language": "cpp",
  "code": "用户代码...",
  "judgeStatus": "Wrong Answer",
  "errorMessage": "期望输出 0 1，实际输出 1 2",
  "recentSubmissions": [
    {"status": "Wrong Answer", "language": "cpp", "code": "完整代码...", "errorMessage": "期望输出 0 1，实际输出 1 2", "createdAt": "2026-06-12T10:00:00Z"},
    {"status": "Wrong Answer", "language": "cpp", "code": "完整代码...", "errorMessage": "期望输出 0 1，实际输出 1 0", "createdAt": "2026-06-12T09:30:00Z"},
    {"status": "Wrong Answer", "language": "cpp", "code": "完整代码...", "errorMessage": "编译错误", "createdAt": "2026-06-12T09:00:00Z"}
  ]
}
```

**滑动窗口**：最近 3 次提交记录

**agent-service 处理**：
1. RAG 检索相关知识（top-3 OI-Wiki + top-2 题解）
2. 构建 prompt，注入标签字典
3. 调用 LLM，要求返回 JSON：

```json
{
  "summary": "问题总结",
  "timeComplexity": "**O(n)** — 说明",
  "spaceComplexity": "**O(1)** — 说明",
  "algorithmTags": ["哈希表"],
  "issues": [{"line": 10, "severity": "error", "message": "...", "hint": "..."}],
  "suggestions": ["建议1", "建议2"]
}
```

**算法标签约束**：prompt 中注入标签列表，要求 `algorithmTags` 字段的值必须来自该列表

---

### 4.2 题解生成（Generate Solution）

**触发场景**：用户通过题目后，点击"AI 生成题解"

**数据流**：
```
前端 → OJ 后端 → agent-service → LLM → agent-service → OJ 后端 → 前端
```

**OJ 后端组装的信息**：
```json
{
  "problemId": 1001,
  "problemTitle": "两数之和",
  "problemContent": "题面内容...",
  "problemEditorial": "官方题解...",
  "algorithmTags": ["数组", "哈希表"],
  "language": "cpp",
  "code": "用户通过的代码...",
  "recentSubmissions": [
    {"status": "Wrong Answer", "code": "...", "createdAt": "2026-06-10T10:00:00Z"},
    {"status": "Wrong Answer", "code": "...", "createdAt": "2026-06-11T09:00:00Z"},
    {"status": "Accepted", "code": "...", "createdAt": "2026-06-12T08:00:00Z"}
  ]
}
```

**提交历史规则**：
- 最近 3 天内，最多 5 条
- 按时间由近到远
- 必须包含最近一次 AC

**agent-service 处理**：
1. RAG 检索相关知识
2. 构建 prompt，要求生成题解草稿
3. 返回 JSON：

```json
{
  "title": "题解标题",
  "content": "题解内容（Markdown）",
  "algorithmTags": ["哈希表"],
  "highlights": ["亮点1"],
  "pitfalls": ["踩坑点1"],
  "complexity": {"time": "**O(n)**", "space": "**O(n)**"}
}
```

**前端行为**：自动填充到题解编辑器，用户修改后手动发布

---

### 4.3 解题辅助（Solve）

**触发场景**：用户点击"获取提示"或"获取解法"

**三个级别**：

| 级别 | 传给 agent-service 的信息 | AI 输出 | 判题验证 |
|------|--------------------------|---------|----------|
| hint | 编辑器代码 + 题面 + 题解 + 样例 + 标签 | 提示（不给代码） | 否 |
| explain | 同上 | 思路解释（不给代码） | 否 |
| full | 同上 | 完整代码 | 是（状态机） |

**hint/explain 的 prompt 设计**：
- 如果用户编辑器为空或只有模板代码，AI 仅提示相关知识点
- 如果用户有实现代码，AI 分析代码后给出针对性提示

**full 级别的状态机**：

```
┌─────────────────────────────────────────────────────────────┐
│                     OJ 后端状态机                            │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ① 发送题目信息给 agent-service，请求生成代码                  │
│                         ↓                                   │
│  ② agent-service 返回代码                                    │
│                         ↓                                   │
│  ③ OJ 后端调用判题服务（不保存记录）                           │
│                         ↓                                   │
│  ④ 判题结果 == AC?                                           │
│     ├─ 是 → 返回代码给前端                                    │
│     └─ 否 → 将判题结果+代码发回 agent-service                 │
│                         ↓                                   │
│  ⑤ agent-service 修改代码，返回新代码                         │
│                         ↓                                   │
│  ⑥ 重试次数 < 3?                                             │
│     ├─ 是 → 回到 ③                                           │
│     └─ 否 → 返回 "抱歉，我也无法通过此题"                      │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**agent-service 返回格式**：
```json
{
  "code": "用户代码...",
  "language": "cpp",
  "explanation": "解题思路说明",
  "algorithmTags": ["动态规划", "背包"],
  "timeComplexity": "**O(n*m)**",
  "spaceComplexity": "**O(m)**"
}
```

---

### 4.4 AI 对话（Chat）

**触发场景**：用户在侧边栏或 AI 训练页面对话

**特性**：
- 支持自动注入题目上下文（题面、样例、用户代码）
- 持久化到 `conversations` + `messages` 表
- 同一会话内支持多轮记忆（最近 N 轮作为上下文）
- 不同会话之间独立
- RAG 自动检索注入

**OJ 后端组装的信息**：
```json
{
  "message": "用户消息",
  "conversationId": "uuid",
  "problemId": 1001,
  "problemContext": {
    "title": "两数之和",
    "content": "题面...",
    "samples": [...],
    "algorithmTags": ["数组", "哈希表"]
  },
  "codeContext": {
    "language": "cpp",
    "code": "用户当前编辑器代码"
  },
  "history": [
    {"role": "user", "content": "之前的提问"},
    {"role": "assistant", "content": "之前的回答"}
  ]
}
```

**agent-service 处理**：
1. RAG 检索相关知识（top-3）
2. 构建 system prompt，注入题目上下文 + RAG 结果
3. 调用 LLM，返回回复

---

### 4.5 知识图谱生成（Knowledge Graph）

**触发场景**：用户点击"整理我的知识图谱"

**数据流**：
```
前端 → OJ 后端（整理用户做题数据） → agent-service → LLM → agent-service → OJ 后端 → 前端
```

**OJ 后端整理的数据**：
```json
{
  "timeRange": "1month",
  "problems": [
    {"id": 1001, "title": "两数之和", "tags": ["数组", "哈希表"], "status": "AC", "attempts": 2},
    {"id": 1002, "title": "最长回文子串", "tags": ["字符串", "动态规划"], "status": "WA", "attempts": 3}
  ],
  "tagStats": {
    "数组": {"solved": 5, "attempted": 8, "acRate": 62.5},
    "哈希表": {"solved": 3, "attempted": 4, "acRate": 75.0},
    "动态规划": {"solved": 1, "attempted": 5, "acRate": 20.0}
  }
}
```

**时间范围**：用户可选 1 周或 1 月（默认 1 月），不超过 1 个月

**agent-service 处理**：
1. 分析用户的算法掌握情况
2. 生成知识图谱节点和边
3. 返回 JSON：

```json
{
  "nodes": [
    {"id": "数组", "label": "数组", "mastery": 80, "category": "基础算法"},
    {"id": "动态规划", "label": "动态规划", "mastery": 20, "category": "动态规划"}
  ],
  "edges": [
    {"source": "数组", "target": "哈希表", "type": "related"},
    {"source": "动态规划", "target": "背包", "type": "contains"}
  ],
  "suggestions": ["建议加强动态规划练习"]
}
```

**标签约束**：节点的 `id` 和 `label` 必须来自统一标签字典

**持久化**：OJ 后端将结果存入数据库，支持后续个性化推荐

---

## 五、API 接口汇总

### 5.1 OJ 后端 → agent-service

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/agent/chat` | POST | AI 对话 |
| `/api/agent/code-diagnosis` | POST | 代码诊断 |
| `/api/agent/solve` | POST | 解题辅助（hint/explain/full） |
| `/api/agent/generate-solution` | POST | 题解生成 |
| `/api/agent/knowledge-graph` | POST | 知识图谱生成 |
| `/api/agent/health` | GET | 健康检查 |
| `/api/agent/rag-status` | GET | RAG 状态 |

### 5.2 agent-service → OJ 后端

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/problems/:id/run` | POST | 运行代码（判题验证，不保存记录） |
| `/api/tags/names` | GET | 获取标签字典 |

### 5.3 前端 → OJ 后端

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/ai/chat` | POST | AI 对话 |
| `/api/ai/code-diagnosis` | POST | 代码诊断 |
| `/api/ai/solve` | POST | 解题辅助 |
| `/api/ai/generate-solution` | POST | 题解生成 |
| `/api/ai/knowledge-graph` | POST | 知识图谱生成 |
| `/api/ai/history` | GET | 对话历史 |
| `/api/ai/conversations/:id/messages` | GET | 会话消息 |
| `/api/tags` | GET | 标签字典（按分类分组） |
| `/api/tags/names` | GET | 标签名列表 |

---

## 六、配置

### 6.1 OJ 后端 config.yaml

```yaml
ai:
  enabled: true
  endpoint: "http://127.0.0.1:8090/api/agent"
  api_key: ""
  model: "mimo-v2.5-pro"
  timeout_seconds: 60
```

### 6.2 agent-service .env

```env
# AI Provider
AI_PROVIDER=openai
OPENAI_API_KEY=your-api-key
OPENAI_BASE_URL=https://token-plan-sgp.xiaomimimo.com/v1
OPENAI_MODEL=mimo-v2.5-pro

# Ollama (fallback)
OLLAMA_URL=http://127.0.0.1:11434
OLLAMA_MODEL=qwen2.5-coder:7b

# Service
AGENT_HTTP_ADDR=:8090
AIOJ_BACKEND_URL=http://127.0.0.1:8080
```

---

## 七、待实现清单

| 优先级 | 功能 | 状态 |
|--------|------|------|
| P0 | 统一标签字典（AlgorithmTag + KnowledgePoint 合并） | 待实现 |
| P0 | OJ 后端 AI 端点配置修正（enabled=true, port=8090） | 待实现 |
| P0 | OJ 后端组装上下文转发给 agent-service | 待实现 |
| P1 | 代码诊断：滑动窗口 + 统一标签 | 待实现 |
| P1 | 题解生成：提交历史 + 自动填充 | 待实现 |
| P1 | 解题辅助：hint/explain/full + 状态机 | 待实现 |
| P1 | AI 对话：持久化 + 多轮记忆 + RAG | 待实现 |
| P2 | 知识图谱：用户数据整理 + 持久化 | 待实现 |
| P2 | RAG：题解索引 + 标签关联检索 | 待实现 |
