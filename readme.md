# Injective Chronos Go

一个基于 Go 的轻量服务，用于周期性拉取 Injective 的行情与配置数据，落库到 MongoDB，并提供简单的 HTTP 查询接口。

## 快速开始

- 依赖
  - Go 1.20+
  - MongoDB、Redis
- 配置
  - 编辑 `etc/config.dev.yaml`（或通过 `ENV=prd` 使用 `etc/config.prd.yaml`）
- 运行
  - 开发：`go run ./cmd/main.go -f etc/config.dev.yaml`
  - 生产：`ENV=prd go run ./cmd/main.go`

## 主要能力

- 定时任务（Cron）
  - 周期性拉取并写入 Mongo：
    - Spot：`config`、`summary_all`、`summary`、`history`
    - Derivative：`config`、`summary_all`、`summary`
    - Market（聚合现货/合约的 marketIds）：`history`
  - 周期由 `Cron.IntervalSec` 控制，启停由 `Cron.Enabled` 控制

- HTTP 接口（默认前缀无鉴权，便于内网调用）
  - 健康检查
    - GET `/healthz`
  - Spot
    - GET `/api/chart/v1/spot/config`
    - GET `/api/chart/v1/spot/market_summary_all?resolution=24h`
    - GET `/api/chart/v1/spot/market_summary?marketId=...&resolution=24h`
    - GET `/api/chart/v1/spot/market/history?marketIDs=...&marketIDs=...&resolution=5&countback=100`
  - Derivative
    - GET `/api/chart/v1/derivative/config`
    - GET `/api/chart/v1/derivative/market_summary_all?resolution=24h`
    - GET `/api/chart/v1/derivative/market_summary?marketId=...&resolution=24h`
    - GET `/api/chart/v1/derivative/market/history?marketIDs=...&resolution=5&countback=100`
  - Market（现货+合约聚合）
    - GET `/api/chart/v1/market/history?marketIDs=...&resolution=5&countback=100`

说明：`countback` 为可选整数，表示回溯的 K 线数量；`resolution` 支持 `1/5/15/30/60/120/240/720/1440`、`24h/7days/30days` 等（以配置/服务端为准）。

## 数据存储

- Mongo 集合（示例，名称由配置文件决定）：
  - `SpotColl`：`kind=config|summary_all|summary`
  - `DerivativeColl`：`kind=summary_all|summary`
  - `MarketColl`：`kind=history`（逐条 K 线，包含 `market/resolution/t/data/updated_at`）
- 建议索引
  - `MarketColl(kind, market, resolution, t)` 复合索引
  - `SpotColl(kind, resolution, updated_at)`、`DerivativeColl(kind, resolution, updated_at)`

## 目录结构

- `cmd/`：入口
- `internal/handler/`：HTTP 路由与处理
- `internal/logic/`：查询/聚合逻辑
- `internal/task/`：定时任务实现
- `internal/injective/`：Injective 客户端
- `internal/model/`：数据模型
- `internal/svc/`：依赖注入（Mongo/Redis/HTTP 客户端）
- `etc/`：配置文件

## 开发与测试

- 运行测试：`go test ./...`
- 代码风格：遵循 Go 官方规范，注意日志与错误处理；修改后请本地 `go vet`/`go test`。

## 注意

- 生产环境请正确设置 Mongo/Redis 连接与权限
- 若需变更拉取频率或开关，请调整配置 `Cron`
