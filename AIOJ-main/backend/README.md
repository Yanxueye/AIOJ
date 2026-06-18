# AIOJ Backend

Go + Gin + GORM + RabbitMQ + gRPC 在线判题平台后端。负责用户系统、题目管理、异步判题、学习统计、知识图谱、AI 能力代理转发、Agent Tool API 等全部平台业务。

## 架构

```
                          ┌─────────────────────────────────────────────────┐
                          │              AIOJ Backend (Gin)                 │
                          │                                                 │
  前端 ─── HTTP ──────→  │  middleware ──→ handler ──→ models (GORM)       │
  (Vue 3)                │       │           │              │               │
                          │       │           │              ↓               │
                          │       │           │         MySQL (terminaloj)   │
                          │       │           │                              │
                          │       │           ├─→ mq.Broker (RabbitMQ)      │
                          │       │           │       ↓                     │
                          │       │           │   mq.Worker ──→ judger      │
                          │       │           │       │         (gRPC)      │
                          │       │           │       ↓                     │
                          │       │           │   remote_judge              │
                          │       │           │                              │
                          │       │           └─→ ai.Client (HTTP)          │
                          │       │                   ↓                     │
                          │       │               agent-service             │
                          └─────────────────────────────────────────────────┘
```

### 分层职责

| 层 | 目录 | 职责 |
|----|------|------|
| 入口 | `cmd/server/` | HTTP API 服务启动、依赖注入、优雅关闭 |
| 入口 | `cmd/judger/` | 独立 gRPC 判题服务入口 |
| 路由 | `handler/router.go` | Gin 路由注册（BuildRouter 统一组装） |
| 中间件 | `middleware/` | JWT 认证、CORS、速率限制、Recovery、RBAC |
| 业务模型 | `models/` | GORM 实体定义（20+ 模型）、DTO |
| 数据库 | `database/` | MySQL 初始化、AutoMigrate、种子数据（52 道题 + 默认用户） |
| 消息队列 | `mq/` | RabbitMQ 生产者 + Worker 消费者（异步判题驱动） |
| 判题客户端 | `judger/` | gRPC 客户端/服务端、JSON Codec（无需 protoc） |
| AI 客户端 | `ai/` | 代理转发到 agent-service |
| 工具 | `utils/` | JWT 管理、密码哈希、掌握度计算、Rating 计算 |

### 异步判题流程

```
前端提交代码
    ↓
POST /api/submissions → submission.Submit()
    ↓
nextSubmissionID() — MySQL LAST_INSERT_ID() 原子递增
    ↓
mq.Broker.Publish() — 发送到 RabbitMQ
    ↓
mq.Worker.Consume() — Worker 消费任务（令牌池并发，默认 4）
    ↓
worker.judgeSubmission()
    ├─ 加载题目 + 测试点（Preload PublishedVersion.TestCases）
    ├─ 指数退避重试 gRPC 调用（最多 3 次）
    └─ 写回结果：status, runtimeMs, memoryKb, caseResults
    ↓
如果 AC：更新 accept_count、study_plan 进度、user.rating、knowledge_mastery
```

### AI 请求转发 (Tool Calling Agent)

AIOJ Backend 作为代理层，转发 AI 请求到 agent-service 的 Tool Calling Agent：

```
前端 POST /api/ai/{action}
    ↓
AIOJ backend 组装上下文（题目信息、用户数据）
    ↓
HTTP POST → agent-service /api/agent/chat {mode, ...}
    ↓
agent-service Agent Loop (max 3 轮):
   LLM 推理 → 调用工具 (query_user_problems / submit_code) → 回调 OJ
    → 继续推理 → 最终回复
    ↓
AIOJ backend 透传给前端
```

> Solve 的 full 级别重试状态机已由 agent-service 的 Agent Loop 接管，LLM 自行调用 submit_code 验证代码正确性。

## 功能

### 用户系统
- 注册/登录（JWT + bcrypt）
- 用户资料（昵称、Bio、Rating、解题数、通过率）
- Rating 历史曲线（每次 AC 后根据题目难度更新）
- 做题热力图（按日统计提交数）
- 角色管理（普通用户 / 管理员，数据库实时校验）

### 题目管理
- 题目 CRUD（管理员权限，含版本控制 / 发布 / 回滚）
- 统一算法标签字典（200+ 标签，11 分类，硬编码在 `internal/data/knowledge.go`）
- 题目难度（简单/中等/困难 + 数值 Rating 800-3000）
- 收藏 / 点赞 / 题解（含管理员删除题解）
- 每日挑战（DailyChallenge）

### 提交与判题
- 代码提交（C++17 / Go / Python3）
- 异步判题（RabbitMQ → Worker → gRPC → remote_judge）
- 代码运行（不保存记录，仅返回结果）
- 提交列表/详情（含源码、测试点结果、编译输出）

### 学习系统
- 学习计划（按知识点分组，含进度追踪）
- 签到打卡（每日统计）
- 知识图谱（200+ 节点，11 分类，硬编码在 `internal/data/knowledge.go`）
- 每日推荐（基于薄弱知识点）
- 学习路径 / 薄弱分析

### AI 能力（经 agent-service Tool Calling Agent）
- AI 对话（支持题目上下文注入、多题关联、RAG 增强、AI 主动工具调用）
- 代码诊断（滑动窗口 3 次提交、判题结果注入）
- 解题辅助（hint / explain / full 三级，Agent Loop 自动验证重试）
- 知识图谱生成（AI 分析用户做题记录，计算掌握度）
- AI 创建题单（AI 分析薄弱点，自动选择未做题目编排题单）
- AI 生成题解（基于用户通过的代码，按模板生成题解）
- AI 端点限流（10/min/user）
- Agent 内部 API（`/api/agent/problems`、`/api/agent/judge`，供 agent-service 工具调用）

### 管理后台
- 用户列表 / 角色变更
- 题目管理（创建、编辑、发布、删除）
- 审计日志（操作记录追踪）

## API 清单

<details>
<summary>展开查看完整 API 列表（40+ 端点）</summary>

### 公开端点

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/auth/register` | 注册 |
| POST | `/api/auth/login` | 登录 |
| GET | `/api/problems` | 题目列表（分页、搜索、筛选） |
| GET | `/api/announcements` | 公告列表 |
| GET | `/api/daily-challenge` | 每日挑战 |
| GET | `/api/study-plans` | 学习计划列表 |
| GET | `/api/study-plans/:id` | 学习计划详情 |
| GET | `/api/knowledge` | 知识点列表 |
| GET | `/api/knowledge/graph` | 知识图谱数据 |
| GET | `/api/knowledge/:id/problems` | 知识点关联题目 |
| GET | `/api/recommendations/daily` | 每日推荐 |
| GET | `/api/tags` | 算法标签字典（按分类分组） |
| GET | `/api/tags/names` | 标签名列表 |

### 认证端点

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/user/profile` | 获取个人资料 |
| PUT | `/api/user/profile` | 更新个人资料 |
| GET | `/api/user/rating-history` | Rating 历史 |
| GET | `/api/user/heatmap` | 做题热力图 |
| GET | `/api/learning-path` | 学习路径 |
| GET | `/api/weakness-analysis` | 薄弱分析 |
| GET | `/api/problems/:id` | 题目详情 |
| POST | `/api/problems/:id/favorite` | 收藏 |
| DELETE | `/api/problems/:id/favorite` | 取消收藏 |
| POST | `/api/problems/:id/solution` | 提交/更新题解 |
| GET | `/api/problems/:id/my-solution` | 我的题解 |
| POST | `/api/solutions/:sid/like` | 点赞题解 |
| POST | `/api/problems/:id/run` | 运行代码 |
| POST | `/api/submissions` | 提交代码 |
| GET | `/api/submissions` | 提交列表 |
| GET | `/api/submissions/:id` | 提交详情（含源码） |
| GET | `/api/submissions/:id/cases` | 测试点结果 |
| POST | `/api/ai/chat` | AI 对话（统一入口，支持 mode） |
| GET | `/api/ai/history` | 对话历史 |
| GET | `/api/ai/conversations/:id/messages` | 对话消息 |
| DELETE | `/api/ai/conversations/:id` | 删除对话 |
| POST | `/api/ai/code-diagnosis` | 代码诊断 |
| POST | `/api/ai/generate-solution` | AI 生成题解 |
| POST | `/api/ai/solve` | 解题辅助 |
| POST | `/api/ai/knowledge-graph` | 知识图谱生成 |
| POST | `/api/ai/create-study-plan` | AI 创建题单 |

### 管理员端点

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/admin/users` | 用户列表 |
| PUT | `/api/admin/users/:id/role` | 变更角色 |
| GET | `/api/admin/audit-logs` | 审计日志 |
| POST | `/api/problems` | 创建题目 |
| PUT | `/api/problems/:id` | 编辑题目 |
| DELETE | `/api/problems/:id` | 删除题目（级联清理） |
| GET | `/api/admin/problems/:id` | 题目管理详情 |
| POST | `/api/admin/problems/:id/publish` | 发布版本 |
| POST | `/api/admin/problems/:id/rollback` | 回滚版本 |
| DELETE | `/api/solutions/:sid` | 删除题解 |

</details>

## 目录结构

```
backend/
├─ cmd/
│  ├─ server/              HTTP API 入口
│  └─ judger/              独立 gRPC 判题服务入口
├─ docker/
│  └─ docker-compose.yml   MySQL + RabbitMQ
├─ internal/
│  ├─ ai/                  AI 代理客户端
│  ├─ config/              YAML 配置加载
│  ├─ database/            MySQL 初始化 + 种子数据
│  ├─ data/                知识图谱硬编码数据
│  │  └─ knowledge.go     200+ 节点树 + 标签字典
│  ├─ handler/             Gin Handler（按领域拆分文件）
│  │  ├─ router.go         路由注册
│  │  ├─ auth.go           认证
│  │  ├─ user.go           用户
│  │  ├─ problem.go        题目（含题解、收藏、点赞）
│  │  ├─ submission.go     提交
│  │  ├─ ai.go             AI Handler
│  │  ├─ agent_handlers.go Agent 内部 API (query / judge)
│  │  ├─ study_plan.go     学习计划
│  │  ├─ knowledge.go      知识图谱 + 标签
│  │  ├─ recommendation.go 推荐
│  │  └─ audit.go          审计日志
│  ├─ judger/              gRPC 客户端/服务端
│  ├─ middleware/           JWT、CORS、限流、RBAC
│  ├─ models/              GORM 实体（20+ 模型）
│  ├─ mq/                  RabbitMQ 生产者 + Worker
│  └─ utils/               JWT、密码、Rating、Mastery
├─ proto/                  gRPC proto 定义
├─ API.md                  API 契约文档
└─ config.yaml             运行配置
```

## 数据模型

| 模型 | 说明 |
|------|------|
| User | 用户（含 Rating、Role） |
| Problem / ProblemVersion / ProblemTestCase / ProblemSample / ProblemTemplate | 题目体系 |
| Submission / SubmissionCaseResult | 提交及测试点结果 |
| ProblemSolution / SolutionLike | 题解及点赞 |
| Favorite | 收藏 |
| UserKnowledgeMastery | 用户知识点掌握度（动态计算） |
| RatingHistory | Rating 变更历史 |
| StudyPlan / StudyPlanItem / UserPlanProgress / UserPlanProgressItem | 学习计划 |
| DailyChallenge / StudyCheckin / Announcement | 每日挑战、签到、公告 |
| AuditLog | 审计日志 |
| Conversation / Message | AI 对话 |
| IdSequence | 提交 ID 序列（原子递增） |

## 配置

`config.yaml` 关键配置项：

```yaml
server:
  host: "0.0.0.0"
  port: 8080
  mode: "debug"

mysql:
  seed: true          # 首次启动自动创建种子数据
  dsn: "toj:toj_password@tcp(127.0.0.1:3306)/terminaloj?charset=utf8mb4&parseTime=True&loc=Local"
  auto_migrate: true

rabbitmq:
  url: "amqp://guest:guest@127.0.0.1:5672/"
  queue: "submissions"
  enabled: true

judger:
  grpc_addr: "127.0.0.1:9090"
  remote: true
  timeout_seconds: 30

ai:
  enabled: true
  endpoint: "http://127.0.0.1:8090/api/agent"
  model: "deepseek-chat"
  timeout_seconds: 180

jwt:
  secret: "your-secret-key"
  expire_hours: 72

rate_limit:
  submit_per_minute: 10
  submit_burst: 3
```

## 快速启动

```cmd
REM 基础设施
docker compose -f docker\docker-compose.yml up -d mysql rabbitmq

REM 启动 API 服务
go run .\cmd\server -config config.yaml
```

默认账号：普通用户 `coder_test` / `123456`，管理员 `admin` / `123456`。

## 测试

```cmd
go test ./...
```

## 相关文档

- [API.md](API.md) — 完整 API 契约
- [../frontend/API.md](../frontend/API.md) — 前端接口约定
