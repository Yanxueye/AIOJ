# 项目不足与改进计划

## 一、缺失功能

### 1.1 RAG 系统

**当前状态：** 未实现。`agent-service/internal/rag/` 目录为空，AI client 的 `Embedding()` 方法存在但无调用方。

**需要改进：** 构建基于 OI-Wiki 文档的向量检索系统，支持语义搜索 + 知识图谱关联检索。

**预期实现：**
- 知识来源：OI-Wiki 文档（向量化存储）+ 用户代码/题解上下文
- 检索方式：向量相似度搜索 + 知识图谱关联检索，综合打分排序
- 向量数据库：本地 Ollama embedding 模型 + 轻量向量存储（如 SQLite-vss 或内存索引）
- 联动：知识图谱提供结构化关联，RAG 提供语义检索，两者融合后注入 AI prompt

---

### 1.2 Rating 历史曲线

**当前状态：** 后端有 Rating 计算逻辑（Elo 公式，每次 AC 后更新），但只存储当前值（`users.rating`），无历史记录。前端 Profile 页只显示当前 Rating 数字。

**需要改进：** 记录每次 Rating 变化，前端展示折线图。

**预期实现：**
- 后端新增 `rating_history` 表：`id, user_id, old_rating, new_rating, delta, problem_id, reason, created_at`
- Worker 中 `updateUserRating()` 写入历史记录
- 新增 API：`GET /api/user/rating-history` 返回最近 N 条记录
- 前端 Profile 页新增 ECharts 折线图，X 轴为时间，Y 轴为 Rating 值
- 可按时间范围筛选（近 1 月 / 近 3 月 / 全部）

---

## 二、需要完善的部分

### 2.1 AI 服务 JSON 结构化输出

**当前状态：** agent-service 的所有 AI 端点（hint/analyze/generate-solution/chat）直接返回大模型的原始文本（Markdown 字符串），前端用 MarkdownRenderer 渲染。大模型输出格式不可控，可能返回散文式文本，前端无法提取结构化字段。

**需要改进：** 所有 AI 端点返回结构化 JSON，前端按字段渲染。

**预期实现：**

**提交分析（analyze/diagnose）返回结构：**
```json
{
  "status": "success",
  "summary": "代码整体良好，存在一处边界问题",
  "timeComplexity": "O(n log n)",
  "spaceComplexity": "O(n)",
  "algorithmTags": ["二分查找", "排序"],
  "codeStyle": "代码结构清晰，变量命名规范",
  "issues": [
    {"line": 15, "severity": "warning", "message": "未处理空数组边界", "hint": "添加 if len(nums)==0 检查"}
  ],
  "suggestions": ["可使用双指针优化空间复杂度"],
  "rawMarkdown": "..." // 保留原始文本用于兜底渲染
}
```

**对话（chat）返回结构：**
```json
{
  "status": "success",
  "reply": "这道题可以使用动态规划...",
  "relatedTopics": ["动态规划", "背包问题"],
  "codeSnippet": "dp[i] = max(dp[i-1], ...)",
  "rawMarkdown": "..."
}
```

**Prompt 优化方向：**
- 要求大模型以 JSON 格式回复，给出明确的 schema 示例
- 算法标签需与知识图谱的 73+ 节点对齐（动态规划、图论、数据结构等）
- 一次分析同时输出：时空间复杂度、算法标签、改进建议、代码风格评价
- agent-service 端解析 JSON，失败时降级返回 rawMarkdown

---

### 2.2 AI 沙箱调用（AI 自检）

**当前状态：** agent-service 的 `judge.Client` 已实现三个方法（Submit/GetResult/RunCode），已注入 Handler，但 **没有任何 handler 调用它**。当前 AI 生成的代码无法自动验证正确性。

**需要改进：** AI 生成代码后自动调用沙箱验证。

**预期实现：**
- `generate-solution` handler：AI 生成题解代码后，调用 `judge.Submit()` 提交验证，将判题结果（AC/WA/TLE）附在返回数据中
- `solve` handler（level=full）：AI 给出完整解法后，自动运行验证
- 新增 `verify` 端点：接收代码+题目ID，AI 分析 + 沙箱验证，返回综合结果
- 调用链路：agent-service → AIOJ backend `/api/problems/{id}/run` → remote_judge gRPC

**需确认的问题：**
- AIOJ backend 的 `/api/problems/{id}/run` 需要 JWT 认证，agent-service 如何获取 token？（方案：agent-service 使用固定 admin 账号登录获取 token，或 AIOJ 后端开放内部 API）

---

### 2.3 知识图谱展现优化

**当前状态：** 知识图谱使用 ECharts force-directed graph（力导向图），10 个大类 + 73 个子节点平铺展示，节点间连线为父子关系。问题是：
- 大类之间没有展现相关关系（如"动态规划"和"贪心"的关联）
- 力导向图节点随机散布，新手难以理解难度递进
- 个性化推荐与知识图谱分离，没有联动

**需要改进：** 采用有向层级图，体现难度递进和知识点关联。

**预期实现：**
- **层级布局：** 按难度分层（基础 → 进阶 → 高级），节点从上到下排列
- **有向边：** 父→子为"包含"关系，同层节点间添加"关联"边（如 DP→贪心、图论→搜索）
- **节点样式：** 大小反映关联题目数量，颜色反映用户掌握度（红→黄→绿渐变）
- **交互：** 悬停显示掌握度和推荐题目，点击跳转题库筛选
- **个性化联动：** 未掌握的节点高亮闪烁，推荐学习路径用虚线标注
- 布局算法：使用 ECharts 的 `dagre` 布局（有向无环图）或手动分层坐标

---

### 2.4 AI 侧边对话优化

**当前状态：** 侧边栏 AI 对话功能基本可用，但存在以下问题：
- 无流式输出，等待时间长
- 未传递编辑器代码作为上下文
- 多轮对话记忆依赖前端 history 数组，无持久化
- UI 简陋，无 Markdown 渲染

**需要改进：** 全面优化侧边栏 AI 对话体验。

**预期实现：**

**A. 流式输出（SSE）：**
- agent-service 新增 SSE 端点 `POST /api/agent/chat/stream`
- AIOJ 后端新增转发端点 `POST /api/ai/chat/stream`
- 前端使用 `EventSource` 或 `fetch + ReadableStream` 接收
- 打字机效果逐字渲染

**B. 编辑器代码上下文：**
- 前端发送对话时自动附带当前编辑器代码和语言
- Prompt 中注入：`当前用户代码（{language}）:\n{code}`
- 支持"分析当前代码"快捷按钮

**C. 多轮对话记忆：**
- 后端 `conversations` 表已有，但需优化上下文窗口管理
- 保留最近 N 轮对话作为上下文（避免 token 超限）
- 支持切换历史对话

**D. UI/UX 优化：**
- 对话消息支持 Markdown 渲染（代码高亮、公式）
- 代码块一键复制
- 对话列表侧边栏（历史对话切换）
- 输入框支持 Shift+Enter 换行
- 加载状态动画优化

---

## 三、优先级排序

| 优先级 | 改进项 | 工作量 | 影响 |
|--------|--------|--------|------|
| P0 | AI JSON 结构化输出 + Prompt 优化 | 中 | AI 功能质量核心 |
| P0 | AI 沙箱调用集成 | 小 | AI 自检能力闭环 |
| P1 | Rating 历史曲线 | 小 | 用户体验提升明显 |
| P1 | AI 侧边对话优化（流式+代码上下文） | 中 | 日常使用体验 |
| P2 | 知识图谱层级化展示 | 中 | 新手友好度 |
| P2 | RAG 系统 | 大 | AI 能力质变，但依赖向量数据库 |
