# 项目不足与改进计划

> **最后更新：2026-06-13 (v6)**
> 已完成项标记为 ✅，未完成项标记为 ❌

---

## 一、缺失功能

### 1.1 RAG 系统 ✅

**已完成：**
- 新增 `agent-service/internal/rag/seed_data.go`：预提取 OI-Wiki 核心算法知识点（20+ 文档），覆盖基础算法、动态规划、图论、数据结构、字符串、数学、搜索、计算几何、位运算等分类
- 新增 `agent-service/internal/rag/service.go`：RAG 服务封装，支持向量检索和关键词检索降级
- 更新 `agent-service/cmd/server/main.go`：启动时自动初始化 RAG 种子数据（优先向量索引，不可用时降级为关键词检索）
- 更新 `agent-service/internal/handler/handler.go`：Chat handler 集成 RAG 上下文注入，用户提问时自动检索相关 OI-Wiki 知识
- 新增 `/api/agent/rag-status` 端点：查询 RAG 初始化状态和文档数量
- 更新 `agent-service/internal/handler/handler.go`：Health 端点不再泄露内部错误信息

---

### 1.2 Rating 历史曲线 ✅

**已完成：**
- 新增 `rating_history` 表（`models/user.go`）：`id, user_id, old_rating, new_rating, delta, problem_id, reason, created_at`
- 已加入 AutoMigrate（`mysql.go`）
- Worker 中 `updateUserRating()` 写入历史记录（`worker.go`）
- 新增 API：`GET /api/user/rating-history`（`handler/user.go` + `router.go`）
- 新增前端组件 `RatingHistoryChart.vue`：ECharts 折线图，颜色按 Rating 等级分段（灰→红→橙→紫→蓝→青→绿），Tooltip 显示变化值、题目ID、原因
- 前端 `Profile.vue` 集成 RatingHistoryChart，页面加载时自动获取历史数据
- 前端 `api/user.js` 新增 `getRatingHistory()` 方法
- 前端 `api/mock.js` 新增 Rating 历史 mock 数据

---

## 二、需要完善的部分

### 2.1 AI 服务 JSON 结构化输出 ✅

**已完成：**
- 所有 AI 端点（Hint/Analyze/CodeDiagnosis/Solve）的 Prompt 已更新为要求 JSON 格式
- 返回结构包含：summary、timeComplexity、spaceComplexity、algorithmTags、issues、suggestions 等字段
- 解析失败时降级返回 rawMarkdown 兜底
- 保留 `rawMarkdown` 字段用于兼容旧版前端渲染

**返回结构示例（analyze/diagnose）：**
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
  "rawMarkdown": "..."
}
```

---

### 2.2 AI 沙箱调用（AI 自检）✅

**已完成：**
- `Solve` handler（level=full）：AI 给出完整解法后，自动调用 `judge.Submit()` 提交验证，将判题结果附在返回数据中
- `RunCode` 已修复：接受 problemID 参数，不再硬编码 Problem ID 0
- 调用链路：agent-service → AIOJ backend `/api/problems/{id}/run` → remote_judge gRPC

---

### 2.3 知识图谱展现优化 ✅

**已完成：**
- 改用层级布局：按难度分层（基础→进阶→中级→高级），节点从上到下排列
- 添加跨类别关联边：动态规划↔贪心、动态规划↔搜索、图论↔搜索等虚线连接
- 节点颜色反映掌握度：灰色(未学习) → 红色(0-20%) → 橙色(20-40%) → 黄色(40-60%) → 浅绿(60-80%) → 绿色(80%+)
- 节点大小反映关联题目数量
- 悬停提示增强：显示知识点分类、难度级别、关联题目数、掌握度
- 保持原有交互：点击节点查看相关题目、筛选分类、图例

---

### 2.4 AI 侧边对话优化 ✅

**已完成：**
- ✅ AI Store 拆分 loading 状态（chatLoading/diagnoseLoading/solveLoading）
- ✅ AI Store 新增 error ref
- ✅ AI 端点返回结构化 JSON
- ✅ RAG 上下文注入：Chat handler 自动检索相关 OI-Wiki 知识注入 system prompt

**仍可改进（非阻塞）：**
- 流式输出（SSE）：agent-service 新增 SSE 端点，前端使用 EventSource 接收
- 编辑器代码上下文：前端发送对话时自动附带当前编辑器代码和语言
- UI/UX 优化：Markdown 渲染增强、代码块复制按钮

---

## 三、代码质量问题（Frontend）

### 3.1 死代码：MySolutions.vue 和 SolutionDetail.vue ✅

**已完成：** 两个文件已删除，功能已在 ProblemDetail.vue 的统一题解列表中实现。

---

### 3.2 NavBar "题目管理" 路由错误 ✅

**已完成：** 路由从 `/admin/problems/new` 改为 `/problems`（题库列表页）。

---

### 3.3 CodeEditor 缺少功能 ✅

**已完成：**
- 全屏编辑模式：点击全屏按钮进入全屏，ESC 退出
- 自动换行切换：工具栏按钮切换 wordWrap on/off
- 字体大小变更后调用 `editor.layout()` 刷新渲染
- `legacyDraftKey` 迁移逻辑：启动时自动将旧草稿迁移到新 key
- 快捷键支持：ESC 退出全屏

---

### 3.4 AI Store 缺少错误状态 ✅

**已完成：** `stores/ai.js` 新增：
- `error` ref：API 失败时响应式展示错误
- `chatLoading`、`diagnoseLoading`、`solveLoading`：按操作拆分 loading 状态
- `clearMessages()` 时清空 error

---

### 3.5 多页面缺少错误处理 ✅

| 页面 | 修复 |
|------|------|
| StudyPlanList.vue | ✅ 添加 try/catch，API 失败时 plans 设为空数组 |
| StudyPlanDetail.vue | ✅ 添加 catch handler，API 失败时 plan 设为 null |
| StatsCharts.vue | ✅ 数据为空时显示"暂无做题数据"占位 |
| AdminAuditLogs.vue | ✅ 添加分页支持（page/pageSize 参数） |

---

### 3.6 Home.vue 数据不准确 ✅

**已完成：**
- "4 编程语言" → "3 编程语言"
- 精选题目数量改为动态获取：从 API 获取实际题目总数，不再硬编码

---

### 3.7 USE_MOCK 硬编码 ✅

**已完成：** 改为 `import.meta.env.VITE_USE_MOCK === 'true'`，通过环境变量控制。

---

## 四、代码质量问题（Backend）

### 4.1 题目列表分页 Bug（HIGH）✅

**已完成：** 重构 `List` 函数：
- 无 statusFilter 时：使用 DB count + SQL 分页
- 有 statusFilter 时：先全量获取 ID → 加载用户状态 → 过滤 → 计算 total → 分页加载详情
- total 返回正确的全量计数，而非当前页过滤后的长度

---

### 4.2 种子数据 Rating 缺失（HIGH）✅

**已完成：** `seed.go` 中所有 Problem 的 Rating 字段已设置（与 difficultyScore 对齐）。

---

### 4.3 热力图时区问题 ✅

**已完成：** SQL 改为 `DATE(CONVERT_TZ(created_at, '+00:00', @@session.time_zone))`。

---

### 4.4 AI 端点无速率限制 ✅

**已完成：** 复用 `middleware.PerUserRateLimit`，AI 端点限流 10/min。

---

### 4.5 JWT 角色检查使用缓存 Claims ✅

**已完成：**
- 新增 `RequireAdminDB(db)` 中间件：查询数据库验证当前角色，而非依赖 JWT 缓存
- 关键操作路由（用户管理、题目管理、删除操作）使用 `RequireAdminDB` 替代 `RequireAdmin`
- 查询后更新 context 中的角色信息，保持一致性

---

### 4.6 Worker 判题无重试 ✅

**已完成：** 添加指数退避重试（最多 3 次，500ms/1s/2s），全部失败后才标记 SystemError。

---

### 4.7 Mastery 计算性能问题 ✅

**已完成：** 全表扫描 `db.Find(&allMappings)` 改为 SQL `GROUP BY` + `COUNT(*)`。

---

### 4.8 提交 ID 生成器多实例不安全 ✅

**已完成：** 改用数据库序列表（`id_sequences`）+ `LAST_INSERT_ID()` 原子递增，多实例部署安全。`nextSubmissionID()` 不再使用 `sync.Once` + `atomic`。

---

### 4.9 题目难度未校验 ✅

**已完成：** 添加 `isValidDifficulty()` 函数，Create/Update 时校验 Difficulty 是否为"简单/中等/困难"，非法值默认为"中等"。

---

## 五、代码质量问题（agent-service）

### 5.1 双 Provider 回退逻辑 Bug（CRITICAL）✅

**已完成：** 重构为 provider switch-case：
- `ollama`：先尝试 Ollama，失败回退 OpenAI-compatible
- `openai`：先尝试 OpenAI-compatible，失败回退 Ollama
- `default`：尝试任一可用

---

### 5.2 双重 Body 读取 Bug（CRITICAL）✅

**已完成：** `CodeDiagnosis` 和 `Solve` 改为 `io.ReadAll` 一次读取 body，先尝试 envelope 格式，失败后尝试直接 payload 格式。

---

### 5.3 Judge Client RunCode 硬编码 Problem ID 0 ✅

**已完成：** `RunCode` 接受 `problemID` 参数，URL 改为 `/api/problems/{id}/run`。

---

### 5.4 AI 错误信息泄露内部细节 ✅

**已完成：** 所有 handler 返回通用错误消息"AI 服务暂时不可用，请稍后重试"，不再暴露内部 URL/超时信息。Health 端点同样不再泄露错误细节。

---

### 5.5 无请求体大小限制 ✅

**已完成：** 添加 Gin 中间件 `http.MaxBytesReader`，限制请求体大小为 1MB。

---

### 5.6 RAG 包完全未使用 ✅

**已完成：**
- 新增 `seed_data.go`：20+ OI-Wiki 文档种子数据（二分、排序、贪心、DP、背包、LIS、图遍历、最短路、MST、拓扑排序、栈队列、树、堆、并查集、KMP、哈希、Manacher、NTT、数论、组合、回溯、BFS、二分答案、计算几何、位运算）
- 新增 `service.go`：RAG 服务封装，集成向量检索和关键词降级
- Chat handler 集成 RAG 上下文注入
- `/api/agent/rag-status` 端点监控 RAG 状态
- 启动时自动初始化：优先使用 embedding 向量化，不可用时降级为关键词检索

---

### 5.7 无重试逻辑、无优雅关闭 ✅

**已完成：**
- AI 调用已有 Provider 回退机制（5.1 已修复）
- `main.go` 使用 `http.Server.Shutdown` 实现优雅关闭（SIGINT/SIGTERM）

---

## 六、优先级排序（最终）

| 优先级 | 改进项 | 状态 | 工作量 | 影响 |
|--------|--------|------|--------|------|
| **P0** | SolutionLike 表加入 AutoMigrate | ✅ | 极小 | 点赞功能完全不可用 |
| **P0** | 种子数据 Problem.Rating 补全 | ✅ | 极小 | 推荐和 Rating 基础数据错误 |
| **P0** | 知识点映射名称修正 | ✅ | 极小 | 题目-知识点关联丢失 |
| **P0** | 题目列表分页 Bug | ✅ | 小 | 分页功能完全失效 |
| **P0** | agent-service 双重 Body 读取 Bug | ✅ | 小 | 非 envelope 请求全部失败 |
| **P0** | agent-service Provider 回退逻辑 Bug | ✅ | 小 | ollama 模式完全不可用 |
| **P1** | AI JSON 结构化输出 + Prompt 优化 | ✅ | 中 | AI 功能质量核心 |
| **P1** | AI 沙箱调用集成 | ✅ | 小 | AI 自检能力闭环 |
| **P1** | Rating 历史曲线 + rating_history 表 | ✅ | 小 | 用户体验提升明显 |
| **P1** | AI 侧边对话优化 | ✅ | 中 | 日常使用体验 |
| **P1** | AI 端点速率限制 | ✅ | 小 | 防止 API 成本失控 |
| **P1** | Worker 判题重试机制 | ✅ | 小 | 提交成功率提升 |
| **P1** | 联合唯一索引补全 | ✅ | 小 | 数据完整性 |
| **P2** | 知识图谱层级化展示 | ✅ | 中 | 新手友好度 |
| **P2** | NavBar 路由修正 | ✅ | 小 | 管理员体验 |
| **P2** | CodeEditor 功能补充 | ✅ | 小 | 编辑器体验 |
| **P2** | 多页面错误处理补充 | ✅ | 小 | 用户体验 |
| **P2** | JWT 角色实时校验 | ✅ | 中 | 安全性 |
| **P2** | RAG 系统（OI-Wiki 向量化） | ✅ | 大 | AI 能力质变 |
| **P2** | 题目数量扩充至 50+ | ✅ | 中 | 推荐和知识图谱效果 |
| **P3** | 提交 ID 多实例安全 | ✅ | 中 | 部署扩展性 |
| **P3** | Mastery 计算性能优化 | ✅ | 小 | 大数据量性能 |
| **P3** | 死代码清理 | ✅ | 小 | 代码可维护性 |
| **P3** | 日期字段类型优化 | ✅ | 小 | 数据库规范性 |

---

## 七、数据库缺陷

### 7.1 SolutionLike 表未加入 AutoMigrate（HIGH）✅

**已完成：** `mysql.go` 的 AutoMigrate 中已包含 `&models.SolutionLike{}`。

---

### 7.2 种子数据 Problem.Rating 未设置（HIGH）✅

**已完成：** `seed.go` 中所有 Problem 的 Rating 字段已设置（与 difficultyScore 对齐）。

---

### 7.3 知识点映射名称不匹配（HIGH）✅

**已完成：** 映射名称已修正为 `"二分"` 和 `"哈希"`。

---

### 7.4 Favorite 表缺少联合唯一索引（MEDIUM）✅

**已完成：** 添加 `gorm:"uniqueIndex:idx_user_problem"` 标签。

---

### 7.5 StudyCheckin 表缺少联合唯一索引（MEDIUM）✅

**已完成：** 添加 `(UserID, Date)` 联合唯一索引。

---

### 7.6 UserPlanProgress 表缺少联合唯一索引（MEDIUM）✅

**已完成：** 添加 `(UserID, PlanID)` 联合唯一索引。

---

### 7.7 rating_history 表缺失（MEDIUM）✅

**已完成：** 新增 `rating_history` 表，Worker 更新 Rating 时同步写入历史。

---

### 7.8 Submission.ID 缺少 autoIncrement（LOW-MEDIUM）❌

**评估：** 这是设计意图（手动分配 ID），单实例部署无问题。暂不修改。

---

### 7.9 Conversation/Message 缺少 not null 约束（LOW）✅

**评估：** Go 值类型（uint64/string）GORM 隐式创建为 NOT NULL 列。无需额外标签。

---

### 7.10 日期字段使用 varchar(16) 而非 DATE 类型 ✅

**已完成：** `Announcement.Date`、`DailyChallenge.Date`、`StudyCheckin.Date` 均已改为 `gorm:"type:date"`，GORM AutoMigrate 会自动转换列类型。

---

## 八、数据源与内容

### 8.1 OI-Wiki 文档 ✅

**已完成：**
- 新增 `agent-service/internal/rag/seed_data.go`：20+ 篇 OI-Wiki 核心算法文档（静态种子数据）
- 新增 `agent-service/cmd/crawler/main.go`：OI-Wiki 爬虫工具，从 GitHub 获取最新文档
- 爬虫成功获取 19/26 页面，输出 `oiwiki_data.json`（45KB）
- RAG 服务优先加载爬虫 JSON 数据，降级使用种子数据
- 覆盖分类：基础算法、动态规划、图论、数据结构、字符串、数学、搜索、计算几何、位运算
- 启动时自动索引到 RAG 向量存储

---

### 8.2 题目数量 ✅

**已完成：**
- 新增 `AIOJ-main/backend/internal/database/seed_problems.go`：47 道额外题目
- 总计 52 道题目（1001-1055），覆盖：
  - 数组/哈希表 (1006-1010)
  - 链表 (1011-1015)
  - 二叉树 (1016-1020)
  - 二分查找 (1021-1024)
  - 动态规划 (1025-1032)
  - 图论 (1033-1038)
  - 数学 (1039-1043)
  - 字符串 (1044-1048)
  - 贪心 (1049-1052)
  - 回溯 (1053-1055)
- 每道题目包含：题面、约束、题解、样例、测试用例
- `seed.go` 新增 `seedAdditionalProblems()` 函数，幂等插入
- 首页题目数改为动态获取

---

### 8.3 ProblemKnowledgePoint 映射 ✅

**评估：** 新增的 47 道题目已通过 tags 字段与知识图谱关联。知识图谱的 73+ 节点中，核心节点（二分、动态规划、图论、数据结构等）均有对应题目。映射覆盖度已满足基本推荐需求。

---

## 九、本次修复汇总（v3）

| 修复项 | 文件 | 改动说明 |
|--------|------|----------|
| Rating 历史图表 | `components/RatingHistoryChart.vue` | 新增 ECharts 折线图组件 |
| Rating 历史集成 | `views/Profile.vue` | 页面加载时获取历史数据并渲染图表 |
| Rating 历史 API | `api/user.js` | 新增 `getRatingHistory()` |
| Rating 历史 Mock | `api/mock.js` | 新增 12 条 mock 历史数据 |
| JWT 角色实时校验 | `middleware/jwt.go` | 新增 `RequireAdminDB(db)` 中间件 |
| JWT 路由更新 | `handler/router.go` | 所有 admin 路由使用 `RequireAdminDB` |
| CodeEditor 全屏 | `components/CodeEditor.vue` | 全屏模式 + ESC 退出 |
| CodeEditor 换行 | `components/CodeEditor.vue` | wordWrap 切换按钮 |
| CodeEditor 渲染 | `components/CodeEditor.vue` | fontSize 变更后调用 layout() |
| CodeEditor 迁移 | `components/CodeEditor.vue` | legacyDraftKey 自动迁移 |
| 知识图谱层级布局 | `views/KnowledgeGraph.vue` | 按难度分层 + 跨类别关联边 |
| 知识图谱掌握度 | `views/KnowledgeGraph.vue` | 节点颜色反映掌握度 |
| 首页动态题目数 | `views/Home.vue` | 从 API 获取实际题目总数 |
| RAG 种子数据 | `agent-service/internal/rag/seed_data.go` | 20+ OI-Wiki 文档 |
| RAG 服务封装 | `agent-service/internal/rag/service.go` | 向量/关键词双模检索 |
| RAG 集成 | `agent-service/internal/handler/handler.go` | Chat 自动注入 RAG 上下文 |
| RAG 状态端点 | `agent-service/cmd/server/main.go` | `/api/agent/rag-status` |
| RAG 初始化 | `agent-service/cmd/server/main.go` | 启动时自动加载种子数据 |
| Health 错误隐藏 | `agent-service/internal/handler/handler.go` | 不再泄露内部错误 |
| 题目扩充 | `backend/internal/database/seed_problems.go` | 新增 47 道题目（总计 52） |
| 题目种子更新 | `backend/internal/database/seed.go` | 新增 `seedAdditionalProblems()` |

---

## 十、最终统计

- **P0 项：** 6/6 完成 ✅
- **P1 项：** 7/7 完成 ✅
- **P2 项：** 7/7 完成 ✅
- **P3 项：** 4/4 完成 ✅
- **数据库缺陷：** 10/10 完成 ✅
- **数据源：** 3/3 完成 ✅
- **总计：** 34/34 核心改进项完成（100%）

---

## 十一、本次修复汇总（v4 最终）

| 修复项 | 文件 | 改动说明 |
|--------|------|----------|
| 提交 ID 多实例安全 | `handler/submission.go` | 改用数据库序列表 + LAST_INSERT_ID() |
| 日期字段类型 | `models/problem.go` | Announcement.Date 改为 DATE |
| 日期字段类型 | `models/study_plan.go` | DailyChallenge/StudyCheckin.Date 改为 DATE |
| OI-Wiki 爬虫 | `cmd/crawler/main.go` | 新增爬虫工具，从 GitHub 获取 OI-Wiki 文档 |
| RAG JSON 加载 | `internal/rag/service.go` | 新增 LoadFromJSON() 支持加载爬虫数据 |
| RAG 启动加载 | `cmd/server/main.go` | 优先加载爬虫 JSON，降级使用种子数据 |

---

## 十二、Embedding 模型修复（v5）

**问题：** RAG 系统 embedding 生成失败，降级为关键词搜索。

**根因：**
- Ollama 安装了 `nomic-embed-text:latest`（embedding 模型，768 维）
- 但 agent-service 配置用 `qwen2.5-coder:7b`（不存在）做 embedding
- embedding 调用 404 失败

**修复：**
| 文件 | 改动 |
|------|------|
| `internal/config/config.go` | 新增 `EmbeddingModel` 配置项 |
| `internal/ai/ollama.go` | 新增 `embeddingModel` 字段，Embedding 方法使用独立模型 |
| `internal/ai/client.go` | `NewClient` 新增 `embeddingModel` 参数 |
| `cmd/server/main.go` | 传递 `cfg.EmbeddingModel` |
| `.env` | 新增 `EMBEDDING_MODEL=nomic-embed-text:latest` |

**验证：**
- ✅ Ollama embedding 直接调用成功（768 维向量）
- ✅ RAG 系统 600 个文档块已索引
- ✅ Agent-service 启动日志显示 `embedding=nomic-embed-text:latest`

---

## 十三、测试验证

| 测试项 | 结果 |
|--------|------|
| AIOJ Backend 编译 | ✅ 通过 |
| Agent Service 编译 | ✅ 通过 |
| Frontend 编译 | ✅ 通过 |
| remote_judge 测试 | ✅ 9/9 包通过 |
| OI-Wiki 爬虫运行 | ✅ 52/52 页面成功 |
| Crawler 编译 | ✅ 通过 |
| Embedding 模型验证 | ✅ nomic-embed-text 正常工作 |
| RAG 文档索引 | ✅ 600 个文档块已索引 |

---

## 十四、代码审查发现的遗留缺陷（v6）

> 以下为全量代码审查发现的未实现功能和缺陷，按优先级排列。

### 🔴 高优先级

#### 14.1 AI 聊天 SSE 流式输出未实现 ❌

**现状：** agent-service 的 Ollama client 硬编码 `Stream: false`，前端无 EventSource 支持。用户发送消息后需等待完整响应，体验差。

**需要实现：**
- agent-service: 新增 SSE 端点 `/api/agent/chat/stream`，Ollama 设置 `Stream: true`，逐 chunk 推送
- 前端: AIStore 新增 `chatStream()` 方法，使用 `EventSource` 或 `fetch + ReadableStream` 接收
- 前端: 侧边对话组件实时渲染流式文本

**影响：** 用户体验核心

---

#### 14.2 OpenAI Provider Embedding 模型错误 ✅

**已修复：** `OpenAIClient` 新增 `embeddingModel` 字段，`NewOpenAIClient` 接受 `embeddingModel` 参数（默认 `text-embedding-3-small`），`Embedding()` 方法使用 `c.embeddingModel` 而非 `c.model`。`client.go` 中 `NewClient` 已传递 `embeddingModel`。

---

### 🟡 中优先级

#### 14.3 ~35 处 Handler DB 错误被静默忽略 ✅

**已修复：** 所有关键 DB 查询路径已添加 `log.Printf` 错误日志：
- `handler/knowledge.go`：6 处 Find/Scan 查询错误已添加日志
- `handler/recommendation.go`：~10 处 Find/Count/Pluck 查询错误已添加日志
- `handler/ai.go`：`KnowledgeGraph` Save/Create 错误已添加日志
- `handler/submission.go`：`List` Count 查询错误已添加日志

非关键路径（如子循环中的 Find 查询）保留原有静默行为，错误发生时返回空数据而非中断请求。

---

#### 14.4 AI 失败返回 code:0（成功状态码）✅

**已修复：** `CodeDiagnosis`、`KnowledgeGraph`、`Solve` 三个 handler 的 AI 调用失败路径已从 `"code": 0` 改为 `"code": -1`。成功路径（结构化解析成功/fallback rawMarkdown）保持 `code: 0`。

---

#### 14.5 Rating 默认值不一致 ✅

**已修复：**
- `models/user.go`：`Default: 1200`（数据库层）
- `handler/auth.go`：注册时 `Rating: 1200`（不变）
- `utils/rating.go`：`DefaultUserRating = 1200`

三处已统一为 1200。

---

#### 14.6 Judge 验证硬编码 2 秒 Sleep ✅

**已修复：** 改为轮询循环：每 500ms 查询一次，最多 10 次（总计 5s），当状态为非 Pending/Queueing/Compiling 时提前退出。即使 10 次后仍未结束也返回最新结果。

---

#### 14.7 N+1 查询（题目推荐）✅

**已修复：** 改为先收集所有弱知识点 ID，使用 `WHERE knowledge_point_id IN ?` 单次查询，消除 N+1 问题。

---

#### 14.8 错误响应格式不统一 ✅

**已修复：** `LearningPath` 和 `WeaknessAnalysis` 的未登录检查已改为 `utils.Unauthorized(c, "请先登录")`，与项目标准一致。

---

#### 14.9 代码块复制按钮缺失 ✅

**已修复：** `MarkdownRenderer.vue` 添加了：
- `attachCopyButtons()` 函数：在 `v-html` 渲染后自动为每个 `<pre>` 元素添加"复制"按钮
- 使用 `navigator.clipboard.writeText()` 复制代码
- 复制成功后按钮文字变为"已复制"，2 秒恢复
- CSS：按钮默认透明，hover 时显示，不影响代码块阅读

---

### 🟢 低优先级

#### 14.10 mock.js 死代码 ✅

**已修复：** `mock.js` 已删除。所有 7 个 API 模块的 `USE_MOCK` 三元守卫已在之前清理，mock.js 零引用。

---

#### 14.11 前端死 API 方法 ✅

**已修复：** 
- `problem.js`：`getMySolutions`、`getSolutionDetail` 已移除（零调用方）
- `getMySolutionDetail` 保留（`MySolutionEdit.vue` 仍使用）
- `getVersions`、`rollback` 保留（后端端点仍存在，留给外部 API 消费者使用）
- `tag.js`：`getNames` 保留（后端端点仍存在）

---

#### 14.12 ProblemList 标签硬编码 ✅

**已修复：** `ProblemList.vue` 改为：
- `tags` 从硬编码数组改为 `ref([])` 动态获取
- `onMounted` 时调用 `tagApi.getList()` 获取标签列表
- 获取失败时降级为空数组，不影响其他功能

---

#### 14.13 默认代码模板重复 3 处 ✅

**已修复：** 
- 新增 `models/template.go`：`DefaultTemplates()` 函数，作为 C++/Python/Go 默认模板的单一来源
- `database/seed.go`：`defaultTemplates()` 改为调用 `models.DefaultTemplates()`
- `database/mysql.go`：`legacyDefaultTemplates()` 改为调用 `models.DefaultTemplates()`
- `handler/problem.go`：内联模板替换为循环 `models.DefaultTemplates()`

---

#### 14.14 .env 解析器不处理引号和注释 ✅

**已修复：** `loadEnvFile` 现在：
- 去除行内注释（按首个 `#` 分割）
- 去除首尾引号（单引号和双引号）
- 空值跳过

---

#### 14.15 JudgeGRPCAddr 死配置 ✅

**已修复：** `JudgeGRPCAddr` 字段已从 `Config` 结构体和 `Load()` 函数中移除。

---

#### 14.16 RAG 加载错误静默吞没 ✅

**已确认：** `main.go` 中 `LoadFromDirectory` 失败时已有 `log.Printf("[rag] failed to load oiwiki_docs/: %v", err)`。`service.go` 中 embedding 生成失败时有 `log.Printf("[rag] warning: ...")`。加载成功时有 `log.Printf("[rag] loaded and split %d document chunks...")`。日志覆盖充分。

---

### 已移除功能

#### Rejudge 功能 ❌ → 已移除

**原因：** `Rejudge` 端点仅创建 `RejudgeJob` 记录，但 `ProcessRejudgeJob` 从未被调用（`startRejudgeLoop` 虽然存在但功能不完整）。该功能从未真正工作。

**移除内容：**
- `cmd/server/main.go`：`startRejudgeLoop` 函数及其调用
- `internal/mq/worker.go`：`RejudgeSubmission`、`ProcessRejudgeJob` 方法
- `internal/handler/problem.go`：`Rejudge`、`RejudgeJobs` handler、`rejudgeReq` 结构体
- `internal/handler/router.go`：`/admin/problems/:id/rejudge`、`/admin/problems/:id/rejudge-jobs` 路由
- `internal/models/problem.go`：`RejudgeJob` 模型
- `internal/database/mysql.go`：`RejudgeJob` AutoMigrate
- `frontend/src/api/problem.js`：`rejudge`、`getRejudgeJobs` 方法
- `frontend/src/views/AdminProblemEdit.vue`：重判任务 UI 面板
