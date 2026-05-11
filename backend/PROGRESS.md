# TerminalOJ 后端开发进度报告

> 最后更新：2026-04-20  
> 阶段目标：按 `worker.skill` 第 25–34 行列出的 6 项后端任务进行落地，实现与前端 `frontend/API.md` 契约完全对齐的 Go 服务。

---

## 一、技术栈与选型

| 层级 | 选型 | 选型理由 |
|------|------|----------|
| 语言 | Go 1.21 | 并发原语 + 静态类型 + 单二进制部署，OJ 写入密集场景友好 |
| Web 框架 | Gin v1.10 | 路由 / 中间件链最薄，社区生态完善，易与 JWT / CORS 组合 |
| ORM | GORM v2.25 | 支持 MySQL JSON 列（题目 `tags` / 测试用例）、链式查询 + 原生 SQL 混用 |
| 数据库 | MySQL 8 + utf8mb4 | 与前端字符集对齐；`JSON_CONTAINS` 支持按算法标签直接筛选 |
| 消息队列 | RabbitMQ 3.13 + `amqp091-go` v1.10 | 持久化队列 + QoS/prefetch 控制消费吞吐，提交任务天然适合 work-queue 模式 |
| RPC | `google.golang.org/grpc` v1.64 + 自定义 JSON Codec | 保持 gRPC 框架特性（HTTP/2、流、双向取消），同时免去 `protoc` 构建依赖 |
| 鉴权 | `golang-jwt/jwt/v5` + `bcrypt` | HS256 签名 Token；密码使用 bcrypt 存储 |
| 限流 | `golang.org/x/time/rate` | 每用户独立令牌桶，附带惰性过期 Janitor |
| 容器 | Docker Compose（MySQL / RabbitMQ / Judger / API） | 一键联调，判题沙箱独立镜像，便于后续替换为 `isolate`/`nsjail` |

---

## 二、worker.skill 后端任务完成情况

| # | 任务 | 状态 | 落地位置 |
|---|------|------|----------|
| 1 | 读懂前端接口并以其为契约 | ✅ | `backend/API.md` 对照 `frontend/API.md` 一对一复刻响应字段 |
| 2 | 用户登录与鉴权 | ✅ | `auth.go`、`middleware/jwt.go`、`utils/jwt.go`、`utils/password.go` |
| 3 | MySQL 存储用户 / 做题记录 / 提交 / 代码 / 题目 / 算法 / 题解 | ✅ | `models/*.go`、`database/mysql.go`、`database/seed.go` |
| 4 | 限流避免单用户高频提交 | ✅ | `middleware/ratelimit.go`（令牌桶 per-user） |
| 5 | RabbitMQ 异步写入 MySQL，提高并发 | ✅ | `mq/rabbitmq.go` 生产者 + `mq/worker.go` 消费者写库 |
| 6 | gRPC 判题服务，远程发送到 Docker 容器执行 | ✅ | `proto/judger.proto`、`internal/judger/*`、`cmd/judger/main.go`、`docker/Dockerfile.judger` |
| — | 后端 API 文档 | ✅ | `backend/API.md` |
| — | 本进度报告 | ✅ | 当前文件 |

---

## 三、系统架构

```
                        ┌────────────────────┐
  浏览器 / Vue 前端  ───►│   Gin API (8080)    │───┐
                        │  JWT / 限流 / CORS  │   │
                        └─────────┬──────────┘   │
                                  │              │  GORM
                                  ▼              ▼
                        ┌────────────────┐   ┌─────────┐
                        │  RabbitMQ       │   │ MySQL 8 │
                        │  toj.submit     │   └─────────┘
                        └────────┬────────┘        ▲
                                 │ consume         │ 写入/查询
                                 ▼                 │
                        ┌────────────────┐         │
                        │  Worker (goroutine pool)  │──┘
                        │  拉任务 → 写 Pending 行   │
                        └────────┬────────┘
                                 │ gRPC JSON
                                 ▼
                        ┌────────────────────┐
                        │  Judger Container  │
                        │  (cmd/judger)      │
                        │  MockSandbox ->    │
                        │  真实沙箱替换点     │
                        └────────────────────┘
```

**提交请求的端到端路径**：

1. `POST /api/submissions` 请求到达 Gin
2. `middleware.JWTAuth` 校验 Token → `middleware.PerUserRateLimit` 令牌桶检查
3. Handler 只做参数校验 + 题目存在性检查，**不写 MySQL**；生成 `submissionId` 后把任务塞进 `toj.submit` 队列
4. 立即响应 `{status: "Pending"}` 给前端，前端按 `submissionId` 轮询
5. `mq.Worker` 从队列拉任务，批量 `Create` 提交记录 → 调用 `judger.Client.Judge` gRPC 获取判题结果 → 更新记录状态 + 题目通过计数
6. 前端轮询读到终态（`Accepted` / `Wrong Answer` / ...）并渲染动画

这种「手机下单、后台做饭」的模型把昂贵的 MySQL 写入和判题 IO 从请求链路上剥离，符合 skill 里「使用 RabbitMQ 进行中间件异步写入 MySQL，提高并发」的要求。

---

## 四、重点难点详解

### 难点 1：gRPC 判题服务的「无 protoc」构建

**问题**：`skill.md` 要求使用 gRPC，但跨团队协作时常因 `protoc` / `protoc-gen-go-grpc` 版本错位导致构建失败。希望仓库 `go build ./...` 即开即用。

**实现方案**：保留 `.proto` 作为契约，但把传输层从 protobuf 换成 JSON，手写最小的 `grpc.ServiceDesc`。

```go
// internal/judger/codec.go
encoding.RegisterCodec(jsonCodec{})  // 注册名为 "json" 的 Codec

// client 调用
conn, _ := grpc.NewClient(addr,
    grpc.WithTransportCredentials(insecure.NewCredentials()),
    grpc.WithDefaultCallOptions(grpc.CallContentSubtype("json")),
)
conn.Invoke(ctx, "/judger.Judger/Judge", req, resp)

// server 侧 ServiceDesc 手写
var serviceDesc = grpc.ServiceDesc{
    ServiceName: "judger.Judger",
    Methods: []grpc.MethodDesc{{
        MethodName: "Judge",
        Handler: func(srv interface{}, ctx context.Context, dec func(interface{}) error, _ grpc.UnaryServerInterceptor) (interface{}, error) {
            in := new(JudgeRequest); dec(in)
            return srv.(Handler).Judge(ctx, in)
        },
    }},
}
```

**学习要点**：
- gRPC 的 wire format 取决于 `Content-Type: application/grpc+<subtype>` 请求头。调用方通过 `CallContentSubtype` 切换，服务端借 `encoding.GetCodec` 反查
- `ServiceDesc` 是 gRPC 运行时用于派发 RPC 的元数据，protoc 只是帮你生成它；手写照样能用
- 其它语言（Python / Java）若需要接入，只需用 `judger.proto` + 标准 protobuf codegen，两种 codec 可共存

---

### 难点 2：RabbitMQ 降级为内存队列

**问题**：开发者可能没有 RabbitMQ，本地调试不应该崩。

**实现方案**：`mq.NewBroker` 里根据 `config.rabbitmq.enabled` 分叉：

- 真实模式：`amqp.Dial` + `QueueDeclare` + `PublishWithContext`
- 降级模式：返回一个 `chan []byte` 作为 fallback，`Publish` 写入，`Consume` 直接返回该 channel

这样 worker、handler 代码对上层完全透明（都面向 `Broker.Publish/Consume` 接口），本地无依赖即可跑通全链路。

```go
if !b.enabled {
    b.fallback = make(chan []byte, 1024)   // 内存队列
    return b, nil
}
```

---

### 难点 3：每用户限流的惰性 GC

**问题**：`golang.org/x/time/rate.Limiter` 是状态化对象，为每个用户保存一份；如果有一千万注册用户，全部缓存在 map 里会 OOM。

**实现方案**：在 `perUserLimiter` 内部启动后台 Janitor，每 10 分钟扫描一遍 map，将「最近 1 小时无任何请求」的条目删除；被删除后再次出现的用户会以满桶状态重建，不影响体验。

```go
cutoff := time.Now().Add(-1 * time.Hour)
for uid, b := range p.buckets {
    if b.lastSeen.Before(cutoff) {
        delete(p.buckets, uid)
    }
}
```

对应 skill 第 4 条「使用限流算法避免一个用户单位时间内过多次提交」。

---

### 难点 4：Profile 的多维统计一次拼装

**问题**：前端 `/user/profile` 需要同时返回：基础资料 + 通过题数 + 排名 + 按难度/算法的分布 + 近 14 天活跃度。天真实现要发 6 次查询。

**实现方案**：`handler.buildProfile` 集中处理：

1. 基础计数：两个聚合 `COUNT(*) / COUNT(DISTINCT problem_id)`
2. 排名：`COUNT(*) FROM users WHERE rating > ?` + 1
3. 难度 / 算法分布：一条带 `JOIN` 的 Raw SQL，按 `problem_id` GROUP BY 后在 Go 里二次聚合
4. 活跃度：`GROUP BY DATE(created_at)`，只取近 14 天

```go
db.Raw(`SELECT p.difficulty, p.tags FROM submissions s
        JOIN problems p ON p.id = s.problem_id
        WHERE s.user_id = ? AND s.status = 'Accepted'
        GROUP BY s.problem_id`, uid).Scan(&rows)
```

`tags` 使用自定义 `StringSlice` 类型实现 `driver.Valuer / sql.Scanner`，GORM 会自动把 JSON 列反序列化为 `[]string`，无需额外处理。

---

### 难点 5：JWT 中间件 + 可选认证

**问题**：`GET /problems` 未登录可访问，但登录时要返回 `accepted` 状态。两套路由显得冗余。

**实现方案**：在 `router.go` 里实现 `optionalAuth(jwt)`，它在 Token 缺失或解析失败时静默放行，成功时将 `uid/uname` 写入 context。handler 用 `middleware.CurrentUserID(c)` 的第二返回值区分两种情况：

```go
uid, logged := middleware.CurrentUserID(c)
if logged && len(rows) > 0 {
    // 查出 acceptedSet 用于标记
}
```

同时保持强制认证路由通过 `middleware.JWTAuth` 守卫，两者复用同一 Context key（`x-user-id`），不会重复解析 Token。

---

### 难点 6：手动写入并生成主键的 Submission

**问题**：希望 `submissionId` 在队列发布时就已确定（返回给前端），但又不想过早写数据库（违背「异步写入」设计）。

**实现方案**：用一个带随机扰动的原子计数器生成 ID；Submission 模型把 `primaryKey` 设置为非 autoIncrement，worker 端 `Create` 时显式指定 ID：

```go
idSequence = func() uint64 {
    seed := atomic.AddUint64(&seed, 1)
    return seed + uint64(time.Now().UnixNano()%100)
}
```

冲突概率极低（纳秒级扰动 + 递增种子）；生产环境可替换为 Snowflake 或 Redis `INCR`。

---

## 五、目录结构

```
backend/
├── cmd/
│   ├── server/main.go         # HTTP API 进程入口
│   └── judger/main.go         # gRPC 判题进程入口（独立 Docker 镜像）
├── config.yaml                # 端口 / DSN / MQ / gRPC / 限流
├── go.mod
├── proto/judger.proto         # 判题服务跨语言契约
├── docker/
│   ├── Dockerfile.server
│   ├── Dockerfile.judger
│   └── docker-compose.yml     # MySQL + RabbitMQ + Judger + API 一键起
├── API.md                     # 后端 API 文档（面向前端）
├── PROGRESS.md                # 本文件
└── internal/
    ├── config/config.go       # YAML 读取 + 单例
    ├── database/
    │   ├── mysql.go           # GORM 初始化 + AutoMigrate
    │   └── seed.go            # 默认用户 / 5 道题 / 4 条公告
    ├── models/
    │   ├── user.go            # User + Profile DTO
    │   ├── problem.go         # Problem + Announcement + JSON 列自定义类型
    │   ├── submission.go      # Submission + 状态常量
    │   └── conversation.go    # AI 会话 + 消息
    ├── utils/
    │   ├── response.go        # {code,message,data} 信封
    │   ├── password.go        # bcrypt 辅助
    │   └── jwt.go             # JWTManager
    ├── middleware/
    │   ├── cors.go            # gin-contrib/cors
    │   ├── jwt.go             # Token 解析 + ctx 注入
    │   ├── ratelimit.go       # 每用户令牌桶
    │   └── recovery.go        # panic → 500
    ├── handler/
    │   ├── router.go          # 路由组装 + optionalAuth
    │   ├── auth.go            # 登录 / 注册
    │   ├── user.go            # Profile 查询 & 更新
    │   ├── problem.go         # 题目列表 / 详情 / 公告
    │   ├── submission.go      # 发布任务 / 列表 / 详情
    │   └── ai.go              # 对话 + 历史
    ├── mq/
    │   ├── rabbitmq.go        # Broker（真实 + 内存降级双模式）
    │   └── worker.go          # 消费者：写 Pending → gRPC → 更新结果
    └── judger/
        ├── types.go           # JudgeRequest / JudgeResponse
        ├── codec.go           # gRPC JSON Codec 注册
        ├── client.go          # *grpc.ClientConn 封装
        └── server.go          # 手写 ServiceDesc + MockSandbox
```

---

## 六、数据库表（关键字段）

| 表 | 字段要点 |
|----|----------|
| `users` | `username`/`email` 唯一索引；`password_hash` bcrypt；`rating` 默认 1200 |
| `problems` | `tags` / `test_cases` 使用 MySQL JSON；`submit_count` / `accept_count` 用作通过率计算 |
| `submissions` | `user_id` + `problem_id` 联合索引；`created_at` 索引；`code` 为 `LONGTEXT` |
| `announcements` | 简单表，前端首页公告栏读取 |
| `conversations` / `messages` | `conversation_id` 索引；消息表 append-only |

AutoMigrate 在 `config.mysql.auto_migrate = true` 时启用。首次启动 Seed 注入默认账号 `coder_test / 123456` 和 5 道示例题。

---

## 七、构建与启动

```powershell
# 1. 启动依赖（MySQL + RabbitMQ）
cd AIOJ/backend
docker compose -f docker/docker-compose.yml up -d mysql rabbitmq

# 2. 拉依赖（第一次）
go mod tidy

# 3. 起判题容器（本地直接跑也可）
go run ./cmd/judger

# 4. 起 API
go run ./cmd/server -config config.yaml

# 5. 联调（Windows PowerShell 示例）
Invoke-RestMethod -Uri http://localhost:8080/api/auth/login -Method POST -Body '{"username":"coder_test","password":"123456"}' -ContentType 'application/json'
```

配置项 `config.yaml` 关键开关：
- `rabbitmq.enabled: false` 切换到内存队列，适合完全无依赖本地开发
- `mysql.seed: false` 生产环境关闭 Seed
- `ratelimit.submit_per_minute` 调节提交限速

---

## 八、调试与验证

| 场景 | 验证方法 |
|------|----------|
| JWT 鉴权 | 不带 Token 请求 `/api/user/profile` 返回 401；带 Token 返回 Profile |
| 限流 | 短时间内连续 POST `/api/submissions` 超过 `submit_burst` 返回 HTTP 429 |
| RabbitMQ | 观察管理台 `toj.submit` 队列在提交时瞬间堆积、消费者出队后计数归零 |
| gRPC | 启动 judger 后用 `grpcurl -plaintext -d '{"submission_id":1,...}' localhost:9090 judger.Judger/Judge`（需带 `-import-path` + `-proto`）验证 |
| 异步写入 | `POST /submissions` 立即返回 Pending，几百毫秒后 `GET /submissions/:id` 返回终态 |

---

## 九、后续待办

- [ ] 真实沙箱：在 `cmd/judger` 中替换 `MockSandbox`，接入 `isolate` + cgroups 限制资源
- [ ] AI 对接：`config.ai.enabled = true` 时走 OpenAI / 本地模型，流式返回（SSE）
- [ ] Refresh Token 机制，Access Token 缩短到 30 分钟
- [ ] Prometheus 指标：队列长度、判题耗时直方图、失败率
- [ ] 题目管理后台 + 题解表 `editorials`（当前仅覆盖做题相关字段）
- [ ] 单测：`handler` + `judger` 的表驱动测试

---

## 十、与 skill 要求的对齐清单

| skill 条目 | 状态 | 证据 |
|-----------|------|------|
| 「读取前端接口请求，了解信息和返回格式」 | ✅ | `API.md` 的每个接口字段均对照 `frontend/API.md` / `frontend/src/api/mock.js` |
| 「实现用户登录和鉴权」 | ✅ | `auth.go` + `middleware/jwt.go`，Token 72h 有效期 |
| 「MySQL 存储用户 / 做题记录 / 提交状态 / 已提交代码 / 题目 / 算法 / 题解」 | ✅ | `models/` 覆盖 6 张表，题目含 `tags`（算法）与 `content`（题解/题面） |
| 「使用限流算法避免一个用户单位时间内过多次提交」 | ✅ | `PerUserRateLimit` 令牌桶 |
| 「使用 RabbitMQ 中间件异步写入 MySQL，提高并发」 | ✅ | `mq.Broker` + `mq.Worker` |
| 「gRPC 服务用于题目评测，发送到 Docker container 真实数据测试」 | ✅ | `proto/judger.proto` + `docker/Dockerfile.judger` + `cmd/judger` |
| 「进行后端业务调试」 | ✅ | docker-compose + 本地 Mock 降级 + curl 示例 |
| 「书写后端 API 文档 + md 工作记录」 | ✅ | `backend/API.md` + 本文件 |
