# remote_judge Goal

## 1. 最终系统目标

`remote_judge` 最终要完成的是一个可独立部署、可联调、可演示、可扩展的远程判题子系统。

最终目标包括：

- 完整的提交、查询、系统状态接口
- 完整的异步判题链路
- Docker 沙箱编译与运行
- 结构化判题结果输出
- gRPC 契约与远程调用能力
- 队列与仓储抽象
- MySQL 持久化能力
- RabbitMQ 异步队列能力
- 真实代码测试链路
- 可维护的测试、脚本和文档

## 2. 当前已经完成

### 2.1 基础工程

- [x] 独立 Go 工程初始化
- [x] 工程目录结构拆分
- [x] README
- [x] `goal.md`
- [x] 系统报告

### 2.2 HTTP API

- [x] `POST /api/submissions`
- [x] `GET /api/submissions`
- [x] `GET /api/submissions/:id`
- [x] `GET /api/submissions/:id/cases`
- [x] `GET /api/judge/languages`
- [x] `GET /api/system/health`
- [x] `GET /api/system/stats`

### 2.3 判题主链路

- [x] 提交服务
- [x] 查询服务
- [x] Worker 异步消费
- [x] Judger 主流程
- [x] 状态流转模型
- [x] 单测试点结果聚合
- [x] `Runtime Error` 判定
- [x] `Time Limit Exceeded` 判定
- [x] `Memory Limit Exceeded` 判定
- [x] `Output Limit Exceeded` 判定

### 2.4 沙箱与执行

- [x] MockSandbox
- [x] DockerCLISandbox
- [x] 执行前镜像检查逻辑
- [x] 真实 Docker 代码执行链路
- [x] C++17 真实编译运行
- [x] Python 3.11 真实运行
- [x] Go 1.22 真实编译运行
- [x] Go 在只读沙箱内使用 `/tmp` 构建缓存
- [x] Go 编译使用 `-p=1` 和 `GOMAXPROCS=1`

### 2.5 内部通信与配置

- [x] gRPC Server 骨架
- [x] gRPC JSON Codec
- [x] `proto/judger.proto`
- [x] 配置模块
- [x] 远程 gRPC client
- [x] embedded / remote Judger 模式切换
- [x] 独立 Judger 启动入口 `cmd/judger`
- [x] remote gRPC client/server 端到端测试

### 2.6 基础设施抽象

- [x] 内存队列
- [x] RabbitMQ 可选队列实现
- [x] 内存题目仓储
- [x] 内存提交仓储
- [x] MySQL 题目仓储
- [x] MySQL 提交仓储
- [x] `schema.sql`
- [x] `docker-compose.yml`
- [x] 统计采集模块

### 2.7 测试与验证

- [x] 单元测试：service / judger / worker / api / sandbox
- [x] HTTP 端到端测试
- [x] remote gRPC client/server 测试
- [x] benchmark
- [x] 压测工具 `cmd/stress`
- [x] 模拟外部调用工具 `cmd/smoke`
- [x] gRPC 压测工具 `cmd/grpcstress`
- [x] 真实 Docker 集成测试通过
- [x] Docker 场景覆盖：Accepted / Wrong Answer / Compile Error / Runtime Error / Time Limit Exceeded / Output Limit Exceeded
- [x] 多语言覆盖：C++17 / Go 1.22 / Python 3.11

### 2.8 镜像与部署

- [x] `cpp17` Alpine 判题镜像
- [x] `go1.22` Alpine 判题镜像
- [x] `python3.11` Alpine 判题镜像
- [x] 本地镜像构建脚本 `docker/build-images.ps1`
- [x] `server + judger + mysql + rabbitmq` Compose 部署
- [x] Compose 健康检查证据
- [x] 本地 remote 双进程模式 smoke 成功证据
- [x] 本地 remote 双进程模式 grpcstress 成功证据

### 2.9 工程增强

- [x] worker 并发配置与受控并发执行
- [x] 系统健康与统计接口增强
- [x] 请求级 `traceId`
- [x] 结构化日志
- [x] 更细的测试点资源结果字段
- [x] 基于 Docker stats / inspect 的资源采样
- [x] 熔断降级
- [x] 镜像预热
- [x] seccomp 安全策略
- [x] workspace pool

## 3. 当前未完成项

- [ ] Compose remote 模式的稳定 `Accepted` 成功终态证据
- [ ] 更完整的部署说明（生产资源参数、日志采集、监控项）
- [ ] 更细的安全加固（磁盘写入配额、输出流硬限制等）

## 4. 当前阶段结论

当前主线已经稳定：

- [x] `go test ./...` 全绿
- [x] `go test ./internal/judger -run Docker -v` 全绿
- [x] Docker CLI 单驱动（已移除 SDK 驱动）
- [x] 真实 Docker 判题可运行
- [x] 本地 remote 双进程模式已拿到 `Accepted` 与 gRPC 压测成功证据

## 5. 下一步计划

1. 继续收敛 Compose remote 模式的成功终态证据
2. 补更完整的部署与运维说明
3. 继续加强沙箱安全策略
