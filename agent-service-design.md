# Agent-Service 设计文档

> 最后更新：2026-06-17

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
                           │ remote_judge │               │  DeepSeek V4  │
                           │  (判题沙箱)   │               │  (OpenAI API) │
                           └──────────────┘               └───────────────┘
```

### 1.2 核心原则

- **用户不可感知**：agent-service 不直接暴露给前端，所有请求经 OJ 后端转发
- **职责分离**：OJ 后端负责数据整理、上下文组装、持久化；agent-service 负责 AI 推理
- **统一标签字典**：知识点（knowledge_points）为唯一源头 → 自动同步到 algorithm_tags 表 → 题目标签强制校验 → agent-service 使用同一套标签（CandidateTagDict）
- **RAG 增强**：agent-service 内置 RAG 检索，自动注入相关知识到 AI prompt
- **掌握度实时计算**：知识图谱节点颜色由 OJ 后端从用户提交记录实时计算，不依赖 AI 或中间表

---

## 二、统一标签字典

### 2.1 设计

知识点表（`knowledge_points`）作为唯一数据源，包含 ~83 个知识点，按 11 个分类分组，通过 parentId 自引用形成树形层级：

| 分类     | 知识点（叶子节点示例）                                                                             |
| -------- | -------------------------------------------------------------------------------------------------- |
| 基础算法 | 枚举、模拟、排序、二分、双指针、前缀和、差分、分治、贪心、递归、离散化                             |
| 数据结构 | 数组、链表、栈、单调栈、队列、单调队列、堆、哈希表、并查集、字典树、线段树、树状数组、平衡树、分块 |
| 动态规划 | 背包DP、区间DP、树形DP、数位DP、状态压缩DP、DP优化、计数DP、概率DP、博弈论DP                       |
| 图论     | 最短路、最小生成树、网络流、二分图、拓扑排序、强连通分量、桥和割点、树上问题、LCA                  |
| 数学     | 质数、GCD/LCM、快速幂、模运算、组合数学、容斥原理、概率期望、矩阵、高斯消元、莫比乌斯反演、博弈论  |
| 字符串   | 字符串处理、KMP、Trie、后缀数组、后缀自动机、AC自动机、Manacher、哈希                              |
| 搜索     | BFS、DFS、迭代加深、IDA*、双向BFS、启发式搜索、折半搜索、回溯                                      |
| 贪心     | 区间贪心、排序贪心、反悔贪心                                                                       |
| 计算几何 | 向量、凸包、半平面交、最近点对、旋转卡壳                                                           |
| 位运算   | 位操作、状态压缩、集合运算                                                                         |

每个分类有一个同名子节点承接题目映射（如"数据结构（分类）"下有叶子"数据结构"），避免非叶子节点被映射题目的情况。

### 2.2 同步链路

```
seed_knowledge.go (83个知识点)
    → knowledge_points 表
    → syncTagsFromKnowledgePoints() → algorithm_tags 表
    → TagHandler.List/Names 提供前端标签选择器
    → ProblemHandler.validateTags() 校验题目标签
    → agent-service CandidateTagDict 硬编码常量
```

- agent-service 使用硬编码的 `CandidateTagDict` 常量（启动时无需联网拉取）
- 前端题目创建/修改时标签必须来自 `algorithm_tags` 表
- 知识图谱按 `problems.tags` 字段（JSON_CONTAINS）直接匹配知识点名称

---

## 三、RAG 系统

### 3.1 数据来源

| 来源         | 存储方式                                      | 更新频率     | 选取数量 |
| ------------ | --------------------------------------------- | ------------ | -------- |
| OI-Wiki 文档 | 本地 markdown → langchaingo 分割 → 嵌入缓存 | 手动运行爬虫 | top-3    |

### 3.2 检索流程

```
用户提问/代码 → 拼接题目标签为查询文本 → 检索 OI-Wiki (top-3) → 注入 System Prompt
```

---

## 四、AI 功能详细设计

### 4.1 代码分析（Code Diagnosis）

**触发场景**：用户提交代码后（无论 AC 与否），点击"AI 分析"

**数据流**：

```
前端 → OJ 后端（整理上下文） → agent-service → LLM → agent-service → OJ 后端 → 前端
```

**OJ 后端传递的信息**：

- 题目信息（标题、题面、题解、样例）
- 用户代码（语言、内容）
- 评测结果（状态、耗时、内存、错误信息）
- 未通过测试点（仅非 AC 时）：输入、预期输出、实际输出
- 最近提交记录（同题目的历史提交）
- 候选算法标签字典（injected in prompt）

**Prompt 区分**：

- **Accepted**：分析代码质量、复杂度、优化空间，不质疑正确性
- **非 Accepted**：分析错误原因、指出具体问题、给出修复建议

**LLM 输出**：

```json
{
  "timeComplexity": "**O(n)**",
  "spaceComplexity": "**O(1)**",
  "algorithmTags": ["哈希表"],
  "suggestions": ["建议内容"]
}
```

**前端渲染**：复杂度可视化曲线 + 算法标签对比（题目标签 vs 代码标签）+ 建议列表

---

### 4.2 题解生成（Generate Solution）

**触发场景**：用户通过题目后，点击"AI 生成题解"

**数据流**：同代码分析

**OJ 后端传递**：题目信息（含题解、标签）+ 用户最近 AC 代码

**agent-service 处理**：

1. RAG 检索（基于题目标签）
2. 构建 prompt（含标签字典约束）
3. 调用 LLM，返回题解草稿 JSON

**LLM 输出**：

```json
{
  "title": "题解标题",
  "content": "题解内容（Markdown）",
  "algorithmTags": ["哈希表"],
  "complexity": {"time": "**O(n)**", "space": "**O(n)**"}
}
```

**前端行为**：自动填充到题解编辑器，用户修改后手动发布

---

### 4.3 解题辅助（Solve）

**触发场景**：用户点击"获取提示"或"获取解法"

**三个级别**：

| 级别    | 传给 agent-service                     | AI 输出                            | 判题验证                |
| ------- | -------------------------------------- | ---------------------------------- | ----------------------- |
| hint    | 编辑器代码 + 题面 + 题解 + 样例 + 标签 | 一句话启发式提示                   | 否                      |
| explain | 同上                                   | 指出最大的一个问题（不给解决方案） | 否                      |
| full    | 同上 + 判题错误（重试时）              | 完整代码                           | 是（状态机，最多 3 次） |

**full 级别状态机**：

```
┌─────────────────────────────────────────────────────────────┐
│                     OJ 后端状态机                            │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ① 发送题目信息给 agent-service，请求生成代码                  │
│                         ↓                                   │
│  ② agent-service 返回代码                                    │
│                         ↓                                   │
│  ③ OJ 后端调用 remote_judge 判题（不保存提交记录）            │
│                         ↓                                   │
│  ④ 判题结果 == Accepted?                                     │
│     ├─ 是 → 返回代码给前端 ✅                                 │
│     └─ 否 → 将判题结果发回 agent-service，请求修正            │
│                         ↓                                   │
│  ⑤ agent-service 修改代码，返回新代码                         │
│                         ↓                                   │
│  ⑥ 重试次数 < 3?                                             │
│     ├─ 是 → 回到 ③                                           │
│     └─ 否 → 返回 "经过 3 次尝试仍无法通过" ❌                  │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**实现位置**：`AIOJ-main/backend/internal/handler/ai.go` → `solveWithRetry()` 方法

**LLM 输出（hint）**：

```json
{ "answer": "一句话启发式提示" }
```

**LLM 输出（explain）**：

```json
{ "answer": "当前代码最大的问题（Markdown）" }
```

**LLM 输出（full）**：

```json
{
  "answer": "解题思路简述",
  "code": "完整代码",
  "language": "cpp",
  "timeComplexity": "**O(n)**",
  "spaceComplexity": "**O(1)**",
  "verifyResult": "✅ 代码已通过验证"
}
```

---

### 4.4 AI 对话（Chat）

**触发场景**：用户在侧边栏或 AI 训练页面对话

**特性**：

- 支持自动注入题目上下文（题面、样例、用户代码）
- 持久化到 `conversations` + `messages` 表
- 同一会话内支持多轮记忆
- 支持对话历史列表和删除
- RAG 自动检索注入（基于题目标签）

**OJ 后端传递**：

```json
{
  "message": "用户消息",
  "conversationId": "uuid",
  "history": [{"role": "user", "content": "..."}, ...],
  "problem": {"id": 1001, "title": "...", "content": "...", "tags": ["哈希表"], "samples": [...]},
  "codeLanguage": "cpp",
  "code": "用户当前编辑器代码"
}
```

**agent-service 处理**：

1. 使用题目标签做 RAG 检索（top-3 OI-Wiki）
2. 构建 system prompt（含题目上下文 + RAG 结果 + 代码）
3. 传入历史消息 + 当前消息
4. 调用 LLM，返回纯文本回复

---

### 4.5 知识图谱（Knowledge Graph）

**触发场景**：用户点击"AI 分析薄弱点"

**数据流**：

```
前端 → OJ 后端（整理做题数据 + 推送上下文获得 Prompt） → agent-service → LLM → agent-service → OJ 后端 → 前端
```

**OJ 后端传递**：

```json
{
  "scope": "recent",
  "problems": [
    {"id": 1001, "title": "两数之和", "tags": ["数组", "哈希表"], "status": "AC", "attempts": 2}
  ],
  "tagStats": {
    "数组": {"solved": 5, "attempted": 8, "acRate": 62.5},
    "哈希表": {"solved": 3, "attempted": 4, "acRate": 75.0}
  }
}
```

**agent-service 输出**：

```json
{
  "nodes": [{"id": "哈希表", "label": "哈希表", "mastery": "proficient", "category": "数据结构"}],
  "edges": [{"source": "数组", "target": "哈希表", "type": "related"}],
  "suggestions": ["建议加强动态规划练习"],
  "rawMarkdown": "分析文本..."
}
```

**前端行为**：在右侧面板展示 AI 分析文本和建议，不覆盖节点颜色（节点颜色由后端从提交记录实时计算）。

---

### 4.6 AI 创建题单（Create Study Plan）**新增**

**触发场景**：知识图谱页点击"AI 创建题单"按钮

**数据流**：

```
前端 → OJ 后端 → agent-service（两次 LLM 调用 + 题目检索） → OJ 后端（创建题单） → 前端
```

**OJ 后端处理**：

1. 整理用户做题记录（problems + tagStats）
2. 识别薄弱知识点（通过率 < 50% 或解决数 < 3）
3. 对每个薄弱标签，按 `JSON_CONTAINS(tags, ...)` 检索未做过的题目作为候选
4. 将用户数据 + 候选题目发送给 agent-service

**agent-service 处理**：

1. 分析用户的薄弱知识点
2. 从候选题目中选择 5~15 道，按难度递进排列
3. 生成题单标题和描述

**LLM 输出**：

```json
{
  "title": "题单标题",
  "description": "题单描述",
  "problemIDs": [1001, 1003, 1005]
}
```

**OJ 后端**：根据返回的 problemIDs 创建 StudyPlan（含 UserID）

---

## 五、API 接口汇总

### 5.1 OJ 后端 → agent-service

| 端点                             | 方法 | 说明                          |
| -------------------------------- | ---- | ----------------------------- |
| `/api/agent/chat`              | POST | AI 对话                       |
| `/api/agent/code-diagnosis`    | POST | 代码诊断                      |
| `/api/agent/solve`             | POST | 解题辅助（hint/explain/full） |
| `/api/agent/generate-solution` | POST | 题解生成                      |
| `/api/agent/knowledge-graph`   | POST | 知识图谱薄弱点分析            |
| `/api/agent/create-study-plan` | POST | AI 创建题单                   |
| `/api/agent/health`            | GET  | 健康检查                      |
| `/api/agent/rag-status`        | GET  | RAG 状态                      |

### 5.2 前端 → OJ 后端（AI 相关）

| 端点                                   | 方法   | 说明         |
| -------------------------------------- | ------ | ------------ |
| `/api/ai/chat`                       | POST   | AI 对话      |
| `/api/ai/code-diagnosis`             | POST   | 代码诊断     |
| `/api/ai/solve`                      | POST   | 解题辅助     |
| `/api/ai/generate-solution`          | POST   | 题解生成     |
| `/api/ai/knowledge-graph`            | POST   | 知识图谱分析 |
| `/api/ai/create-study-plan`          | POST   | AI 创建题单  |
| `/api/ai/history`                    | GET    | 对话历史列表 |
| `/api/ai/conversations/:id/messages` | GET    | 会话消息     |
| `/api/ai/conversations/:id`          | DELETE | 删除会话     |

---

## 六、配置

### 6.1 OJ 后端 config.yaml

```yaml
ai:
  enabled: true
  endpoint: "http://127.0.0.1:8090/api/agent"
  api_key: ""
  model: "deepseek-chat"
  timeout_seconds: 180
```

### 6.2 agent-service .env

```env
AI_PROVIDER=openai
OPENAI_API_KEY=sk-xxxxxxxx
OPENAI_BASE_URL=https://api.deepseek.com/v1
OPENAI_MODEL=deepseek-chat
AI_THINKING=false
OLLAMA_URL=http://127.0.0.1:11434
OLLAMA_MODEL=qwen2.5-coder:7b
AGENT_HTTP_ADDR=:8090
EMBEDDING_MODEL=nomic-embed-text:latest
```

> `.env` 在 `agent-service/` 目录下，已加入 `.gitignore`

---

## 七、关键设计决策

| 决策           | 说明                                                                               |
| -------------- | ---------------------------------------------------------------------------------- |
| 知识图谱掌握度 | 实时从 submissions 表计算（用户做过该标签的题数 / 该标签总题数 × 100），不依赖 AI |
| 题目检索方式   | 通过 `JSON_CONTAINS(tags, '"标签名"')` 直接匹配 problems.tags，不使用桥接表      |
| 标签字典同步   | 知识点表 → 同步到 algorithm_tags 表 → 题目标签校验 → agent-service 硬编码常量   |
| 分类节点架构   | 每个分类节点（如"图论（分类）"）下有一个同名叶子节点（"图论"）承接题目映射         |
| 题单权限       | 创建者可以编辑/删除，其他用户只能收藏/查看                                         |
| reasoning 模式 | 已移除（DeepSeek 不支持 `thinking` 字段）                                        |
