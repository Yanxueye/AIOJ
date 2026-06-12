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

## 三、代码质量问题（Frontend）

### 3.1 死代码：MySolutions.vue 和 SolutionDetail.vue

**当前状态：** 两个文件存在于 `src/views/` 但未被 router 引用，无任何组件 import 它们。

**预期处理：** 删除这两个文件，功能已在 ProblemDetail.vue 的统一题解列表中实现。

---

### 3.2 NavBar "题目管理" 路由错误

**当前状态：** 管理员下拉菜单的"题目管理"命令路由到 `/admin/problems/new`（创建页面），而非题目管理列表。

**预期实现：** 新增 AdminProblemList.vue 管理页面（含搜索、筛选、批量操作），或路由到 ProblemList.vue 并默认显示管理操作列。

---

### 3.3 CodeEditor 缺少功能

**当前状态：**
- `legacyDraftKey` prop 定义但未使用，迁移逻辑不完整
- 无全屏/禅模式切换
- 无自动换行切换
- 字体大小变更后未调用 `editor.layout()` 可能导致渲染问题

**预期实现：** 补充上述功能，参考 LeetCode 编辑器体验。

---

### 3.4 AI Store 缺少错误状态

**当前状态：** `stores/ai.js` 无 reactive `error` ref，API 失败时只 re-throw，消费组件无法响应式展示错误。单一 `loading` ref 多操作共用，并发时互相干扰。

**预期实现：** 新增 `error` ref，按操作拆分 loading 状态（`chatLoading`, `diagnoseLoading`, `solveLoading`）。

---

### 3.5 多个页面缺少错误处理

| 页面 | 问题 |
|------|------|
| StudyPlanList.vue | `onMounted` 无 try/catch，API 失败时无反馈 |
| StudyPlanDetail.vue | 无 catch handler，API 失败显示空白页 |
| StatsCharts.vue | 数据为空时 ECharts 渲染空白图表，无 empty 状态 |
| AdminAuditLogs.vue | 无分页，大数据量时性能问题 |

---

### 3.6 Home.vue 数据不准确

**当前状态：** Hero 区域声称"4 编程语言"，实际只支持 3 种（C++、Python、Go）。

**预期处理：** 改为"3 编程语言"或新增 Java/JavaScript 支持。

---

### 3.7 USE_MOCK 硬编码

**当前状态：** `api/index.js` 中 `const USE_MOCK = false` 硬编码，切换 mock 需改源码。

**预期实现：** 改为 `import.meta.env.VITE_USE_MOCK === 'true'`，通过环境变量控制。

---

## 四、代码质量问题（Backend）

### 4.1 题目列表分页 Bug（HIGH）

**当前状态：** `handler/problem.go:202` 使用 `len(list)` 作为 total 返回，且 statusFilter 在数据库分页后应用，导致：
- 分页 total 不准确
- 状态筛选后返回条数少于 pageSize

**预期实现：** 将 statusFilter 下推到 SQL 查询层（WHERE 条件），使用数据库 count 作为 total。

---

### 4.2 种子数据 Rating 缺失（HIGH）

**当前状态：** `seed.go` 创建 Problem 时未设置 Rating 字段，全部默认为 800。"合并K个升序链表"（difficultyScore=1900）的 Rating 也是 800。

**影响：** 推荐系统按 Rating 过滤失效，用户 Rating 更新计算不准确。

**预期实现：** 种子数据中为每个 Problem 显式设置 Rating（与 difficultyScore 对齐）。

---

### 4.3 热力图时区问题

**当前状态：** `Heatmap` handler 使用 MySQL `DATE(created_at)` 返回服务器时区日期，与前端本地时区可能不一致。

**预期实现：** 使用 `DATE(CONVERT_TZ(created_at, '+00:00', @@session.time_zone))` 或统一存储 UTC 日期。

---

### 4.4 AI 端点无速率限制

**当前状态：** `/ai/chat`、`/ai/code-diagnosis` 等端点无 per-user 限流，可被无限调用。

**预期实现：** 复用 `middleware.PerUserRateLimit`，设置 AI 端点限流（如 10/min）。

---

### 4.5 JWT 角色检查使用缓存 Claims

**当前状态：** `RequireAdmin()` 从 JWT claims 读取角色，管理员降权后旧 token 仍有效。

**预期实现：** 关键操作（角色变更、删除）时查询数据库验证当前角色，或引入 token 版本号机制。

---

### 4.6 Worker 判题无重试

**当前状态：** 判题 RPC 调用失败后直接标记 SystemError，无重试机制。瞬时网络错误会导致提交永久失败。

**预期实现：** 添加指数退避重试（最多 3 次），全部失败后才标记 SystemError。

---

### 4.7 Mastery 计算性能问题

**当前状态：** `UpdateMastery()` 每次 AC 都加载全表 `ProblemKnowledgePoint` 记录计算总数。

**预期实现：** 缓存 KP 总数（或在 KP 表中维护 problem_count 字段），避免全表扫描。

---

### 4.8 提交 ID 生成器多实例不安全

**当前状态：** `nextSubmissionID` 使用 `sync.Once` + `MAX(id)` 初始化，多实例部署会 ID 冲突。

**预期实现：** 使用数据库自增 ID，或引入 Redis/UUID 分布式 ID 生成。

---

### 4.9 题目难度未校验

**当前状态：** `Create`/`Update` 不校验 Difficulty 是否为"简单/中等/困难"，可写入任意字符串。

**预期实现：** 添加枚举校验。

---

## 五、代码质量问题（agent-service）

### 5.1 双 Provider 回退逻辑 Bug（CRITICAL）

**当前状态：** `ai/client.go` 的 `Chat` 方法中，当 provider=="ollama" 时，通用 fallback 块先执行并 return，ollama-first 逻辑（lines 59-69）永远不可达。`Embedding` 方法同理。

**预期实现：** 重构为 provider switch：openai → try primary then fallback；ollama → try fallback then primary。

---

### 5.2 双重 Body 读取 Bug（CRITICAL）

**当前状态：** `handler.go` 的 `CodeDiagnosis` 和 `Solve` 中，第一次 `ShouldBindJSON` 失败后 body 已消费，第二次绑定必然失败。

**预期实现：** 使用 `c.ShouldBindBodyWith` 或先 `io.ReadAll` 缓存 body 再分别解析。

---

### 5.3 Judge Client RunCode 硬编码 Problem ID 0

**当前状态：** `judge/client.go:123` 使用 `problems/0/run`，AIOJ 后端会 404。

**预期实现：** RunCode 接受 problemID 参数，或使用 AIOJ 后端的自定义输入端点。

---

### 5.4 AI 错误信息泄露内部细节

**当前状态：** `Hint`、`Analyze`、`GenerateSolution` handler 返回 `err.Error()` 给客户端，暴露内部 URL、超时等信息。

**预期实现：** 返回通用错误消息，内部错误记录到日志。

---

### 5.5 无请求体大小限制

**当前状态：** 所有 handler 无 `http.MaxBytesReader`，大 payload 可导致 OOM。

**预期实现：** 添加 Gin 中间件限制请求体大小（如 1MB）。

---

### 5.6 RAG 包完全未使用

**当前状态：** `internal/rag/store.go` 和 `retriever.go` 存在但从未被 import。struct tag `vector:"-"` 应为 `json:"-"`。

**预期处理：** 实现 RAG 功能或暂时删除避免误导。

---

### 5.7 无重试逻辑、无优雅关闭

**当前状态：** AI 调用无重试，`r.Run()` 不处理 SIGINT/SIGTERM。

**预期实现：** 添加指数退避重试，使用 `http.Server.Shutdown` 实现优雅关闭。

---

## 六、优先级排序（更新）

| 优先级 | 改进项 | 工作量 | 影响 |
|--------|--------|--------|------|
| **P0** | 题目列表分页 Bug（statusFilter 下推 SQL） | 小 | 分页功能完全失效 |
| **P0** | 种子数据 Rating 缺失 | 小 | 推荐和 Rating 系统基础数据错误 |
| **P0** | AI JSON 结构化输出 + Prompt 优化 | 中 | AI 功能质量核心 |
| **P0** | agent-service 双重 Body 读取 Bug | 小 | 非 envelope 请求全部失败 |
| **P0** | agent-service Provider 回退逻辑 Bug | 小 | ollama 模式完全不可用 |
| **P1** | AI 沙箱调用集成 | 小 | AI 自检能力闭环 |
| **P1** | Rating 历史曲线 | 小 | 用户体验提升明显 |
| **P1** | AI 侧边对话优化（流式+代码上下文） | 中 | 日常使用体验 |
| **P1** | AI 端点速率限制 | 小 | 防止 API 成本失控 |
| **P1** | Worker 判题重试机制 | 小 | 提交成功率提升 |
| **P2** | 知识图谱层级化展示 | 中 | 新手友好度 |
| **P2** | NavBar 路由修正 + AdminProblemList | 中 | 管理员体验 |
| **P2** | CodeEditor 功能补充 | 小 | 编辑器体验 |
| **P2** | 多页面错误处理补充 | 小 | 用户体验 |
| **P2** | JWT 角色实时校验 | 中 | 安全性 |
| **P2** | RAG 系统 | 大 | AI 能力质变 |
| **P3** | 提交 ID 多实例安全 | 中 | 部署扩展性 |
| **P3** | Mastery 计算性能优化 | 小 | 大数据量性能 |
| **P3** | 死代码清理 | 小 | 代码可维护性 |

---

## 七、数据库缺陷

### 7.1 SolutionLike 表未加入 AutoMigrate（HIGH）

**当前状态：** `models/problem.go` 定义了 `SolutionLike` 结构体，`handler/problem.go` 中点赞功能使用了该表，但 `mysql.go` 的 AutoMigrate 列表中 **遗漏了该表**。运行时点赞会报 SQL 错误。

**预期修复：** 在 `mysql.go` 的 AutoMigrate 中添加 `&models.SolutionLike{}`。

---

### 7.2 种子数据 Problem.Rating 未设置（HIGH）

**当前状态：** `seed.go` 创建 Problem 时未设置 Rating 字段，全部默认为 800。导致：
- "合并K个升序链表"（difficultyScore=1900）的 Rating 为 800
- 推荐系统按 Rating 过滤全部失效
- 用户 Rating 更新计算使用错误的题目 Rating

**预期修复：** 在 `seededProblem` 结构体中添加 Rating 字段，种子数据中为每个 Problem 设置与 difficultyScore 对齐的 Rating。

---

### 7.3 知识点映射名称不匹配（HIGH）

**当前状态：** `seed.go` 中 `seedProblemKnowledgeMappings` 使用了 `"二分查找"` 和 `"字符串哈希"` 两个知识点名称，但 `seed_knowledge.go` 中不存在这两个名称（实际为 `"二分"` 和 `"哈希"`）。`getKPID()` 找不到匹配返回 0，映射被静默跳过。

**预期修复：** 将映射名称改为 `"二分"` 和 `"哈希"`。

---

### 7.4 Favorite 表缺少联合唯一索引（MEDIUM）

**当前状态：** `Favorite` 表的 `UserID` 和 `ProblemID` 各有单独索引，但无联合唯一索引。用户可重复收藏同一题目。

**预期修复：** 添加 `gorm:"uniqueIndex:idx_user_problem"` 标签。

---

### 7.5 StudyCheckin 表缺少联合唯一索引（MEDIUM）

**当前状态：** `StudyCheckin` 表无 `(UserID, Date)` 联合唯一索引，用户可同一天多次打卡。

**预期修复：** 添加联合唯一索引。

---

### 7.6 UserPlanProgress 表缺少联合唯一索引（MEDIUM）

**当前状态：** `UserPlanProgress` 表无 `(UserID, PlanID)` 联合唯一索引，可产生重复进度记录。

**预期修复：** 添加联合唯一索引。

---

### 7.7 rating_history 表缺失（MEDIUM）

**当前状态：** Worker 的 `updateUserRating()` 直接覆盖 `users.rating`，无历史记录表。Rating 曲线、审计追踪、错误恢复均不可能。

**预期修复：** 新增 `rating_history` 表（已在 1.2 节描述），Worker 更新 Rating 时同步写入历史。

---

### 7.8 Submission.ID 缺少 autoIncrement（LOW-MEDIUM）

**当前状态：** `Submission.ID` 只有 `primaryKey` 无 `autoIncrement`，与其他所有模型不一致。ID 由 Worker 的 `nextSubmissionID()` 手动分配，多实例部署会冲突。

**预期处理：** 确认是否为设计意图。如需自增，添加 `autoIncrement` 标签。

---

### 7.9 Conversation/Message 缺少 not null 约束（LOW）

**当前状态：** `Conversation.UserID` 和 `Message.ConversationID` 缺少 `not null` 标签，可能产生孤立记录。

**预期修复：** 添加 `not null` 约束。

---

### 7.10 日期字段使用 varchar(16) 而非 DATE 类型（LOW）

**当前状态：** `Announcement.Date`、`DailyChallenge.Date`、`StudyCheckin.Date` 均使用 `varchar(16)` 存储日期字符串，无法利用数据库的日期函数和索引。

**预期处理：** 现有数据量小影响不大，后续可迁移为 MySQL DATE 类型。

---

## 八、数据源与内容缺失

### 8.1 OI-Wiki 文档未整理

**当前状态：** 知识图谱的 73+ 节点有 OI-Wiki URL 链接，但 OI-Wiki 文档内容未向量化存储。RAG 系统的 `internal/rag/` 为空，无任何文档索引。

**预期实现：**
- 爬取 OI-Wiki 核心页面（动态规划、图论、数据结构等 10 大类）
- 文档分块（按标题层级切分，每块 500-1000 token）
- 使用 Ollama embedding 模型向量化
- 存储到轻量向量数据库（SQLite-vss 或内存索引）
- 与知识图谱节点关联（每个知识点对应若干文档块）

---

### 8.2 题目数量不足

**当前状态：** 种子数据仅 5 道题目（1001-1005），覆盖的算法标签有限。知识图谱有 73+ 节点但大部分无关联题目，推荐系统可选题目池过小。

**预期处理：**
- 补充种子题目至 50+ 道，覆盖核心算法分类
- 每个知识点至少关联 2-3 道题目
- 或提供批量导入工具支持从外部 OJ 导入题目

---

### 8.3 ProblemKnowledgePoint 映射不完整

**当前状态：** 5 道种子题目中，仅部分有知识点映射（且有 2 条因名称不匹配被跳过）。大部分知识图谱节点无关联题目。

**预期修复：** 修复名称匹配问题后，为每道种子题目添加 2-3 个知识点映射。

---

## 九、优先级排序（最终）

| 优先级 | 改进项 | 工作量 | 影响 |
|--------|--------|--------|------|
| **P0** | SolutionLike 表加入 AutoMigrate | 极小 | 点赞功能完全不可用 |
| **P0** | 种子数据 Problem.Rating 补全 | 极小 | 推荐和 Rating 基础数据错误 |
| **P0** | 知识点映射名称修正 | 极小 | 题目-知识点关联丢失 |
| **P0** | 题目列表分页 Bug | 小 | 分页功能完全失效 |
| **P0** | agent-service 双重 Body 读取 Bug | 小 | 非 envelope 请求全部失败 |
| **P0** | agent-service Provider 回退逻辑 Bug | 小 | ollama 模式完全不可用 |
| **P1** | AI JSON 结构化输出 + Prompt 优化 | 中 | AI 功能质量核心 |
| **P1** | AI 沙箱调用集成 | 小 | AI 自检能力闭环 |
| **P1** | Rating 历史曲线 + rating_history 表 | 小 | 用户体验提升明显 |
| **P1** | AI 侧边对话优化（流式+代码上下文） | 中 | 日常使用体验 |
| **P1** | AI 端点速率限制 | 小 | 防止 API 成本失控 |
| **P1** | Worker 判题重试机制 | 小 | 提交成功率提升 |
| **P1** | 联合唯一索引补全（Favorite/Checkin/Progress） | 小 | 数据完整性 |
| **P2** | 知识图谱层级化展示 | 中 | 新手友好度 |
| **P2** | NavBar 路由修正 + AdminProblemList | 中 | 管理员体验 |
| **P2** | CodeEditor 功能补充 | 小 | 编辑器体验 |
| **P2** | 多页面错误处理补充 | 小 | 用户体验 |
| **P2** | JWT 角色实时校验 | 中 | 安全性 |
| **P2** | RAG 系统（OI-Wiki 向量化） | 大 | AI 能力质变 |
| **P2** | 题目数量扩充至 50+ | 中 | 推荐和知识图谱效果 |
| **P3** | 提交 ID 多实例安全 | 中 | 部署扩展性 |
| **P3** | Mastery 计算性能优化 | 小 | 大数据量性能 |
| **P3** | 死代码清理 | 小 | 代码可维护性 |
| **P3** | 日期字段类型优化 | 小 | 数据库规范性 |
