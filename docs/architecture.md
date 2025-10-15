## 项目思维导图与架构总览

下图为基于当前代码实现梳理出的整体逻辑思维导图（使用 Mermaid 表达，支持直接在 GitHub 或本地插件中渲染）：

```mermaid
mindmap
  root((injective-chronos-go))
    启动流程(cmd/main.go)
      加载配置(etc/config.yaml)
      初始化日志(internal/log)
      构建ServiceContext(internal/svc)
      启动REST服务(go-zero/rest)
      启动定时任务(internal/task.StartCron)
    配置(internal/config)
      RestConf
      Redis{Address,Password,DB}
      Mongo{URI,Database,Collections{Spot,Derivative}}
      Injective{BaseURL,Paths...,TimeoutMs}
      Cron{Enabled,IntervalSec}
    运行时上下文(ServiceContext)
      Redis客户端
      Mongo客户端
        集合: spot_market_summaries
        集合: derivative_market_summaries
      HTTP客户端(超时=Injective.TimeoutMs)
    HTTP接口(internal/handler)
      路由(internal/handler/routes.go)
        GET /api/chart/v1/spot/market_summary_all
          -> SpotMarketSummaryAllHandler
          -> ChartLogic.GetMarketSummaryAll("spot")
        GET /api/chart/v1/spot/market_summary?market=&resolution=
          -> SpotMarketSummaryHandler
          -> ChartLogic.GetMarketSummary("spot", market, res)
        GET /api/chart/v1/spot/config
          -> SpotConfigHandler
          -> ChartLogic.GetSpotConfig()
        GET /api/chart/v1/derivative/market_summary_all
          -> DerivativeMarketSummaryAllHandler
          -> ChartLogic.GetMarketSummaryAll("derivative")
        GET /api/chart/v1/derivative/market_summary?market=&resolution=
          -> DerivativeMarketSummaryHandler
          -> ChartLogic.GetMarketSummary("derivative", market, res)
        GET /api/chart/v1/derivative/config
          -> DerivativeConfigHandler
          -> ChartLogic.GetDerivativeConfig()
        GET /healthz
    领域逻辑(internal/logic/ChartLogic)
      MarketSummaryAll
        Redis缓存(chart:summary_all:{type})
        Mongo读取(kind=summary_all，最新updated_at)
        命中则回填Redis(5min)
      MarketSummary(单市场)
        解析/校验resolution(NormalizeResolution)
          从Mongo(derivative kind=config)读取supported_resolutions
          默认"24h"
        优先Redis缓存(chart:summary:{type}:{res}:{market})
        Mongo优先(kind=summary, 可包含resolution/market)
        回退: 从summary_all列表挑选对应market
        命中则回填Redis(5min)
      DerivativeConfig
        先Redis(chart:derivative:config)
        再Mongo(kind=config，最新)
        再回退实时请求Injective并持久化+缓存(10min)
      SpotConfig
        先Redis(chart:spot:config)
        再Mongo(kind=config，最新)
        再回退实时请求Injective并持久化+缓存(10min)
    定时任务(internal/task/cron.go)
      条件: Cron.Enabled=true
      周期: IntervalSec
      Tick循环
        抓取Spot Summary All
          -> 插入Mongo(spot, kind=summary_all)
        抓取Spot Config
          -> 插入Mongo(spot, kind=config)
        抓取Derivative Config
          -> 插入Mongo(derivative, kind=config)
          -> 解析supported_resolutions
        抓取Derivative Summary All
          -> 插入Mongo(derivative, kind=summary_all)
          -> 解析marketIds
        对每个(marketId × resolution)
          -> 抓取Derivative Summary
          -> 插入Mongo(derivative, kind=summary, market, resolution)
    外部服务(Injective API)
      BaseURL: Injective.BaseURL
      端点:
        /spot/market_summary_all
        /spot/market_summary
        /derivative/market_summary_all
        /derivative/market_summary(含resolution)
        /derivative/config
      客户端: internal/injective/client.go(HTTP GET，JSON解码，非2xx报错)
    数据存储
      Redis: 短期缓存(chart:*)
      Mongo: 文档结构
        { kind, data, updated_at, [market], [resolution] }
    可观测性
      日志: logx -> 文件 log/app.log
```

### 模块概览

- 启动流程：`cmd/main.go` 加载 `etc/config.yaml`，初始化日志，创建 `ServiceContext`，注册路由并启动 REST 服务，同时启动 `task.StartCron` 定时任务。
- 配置聚合：`internal/config` 定义 Redis、Mongo、Injective、Cron 以及 go-zero 的 `RestConf`。
- 运行时上下文：`internal/svc.ServiceContext` 聚合 Redis、Mongo 各集合、带超时的 `http.Client`。
- HTTP 接口：`internal/handler` 注册各 GET 路由，调用 `internal/logic` 完成业务。
- 领域逻辑：`internal/logic/ChartLogic` 负责缓存优先、Mongo 回源、回退策略及分辨率校验。
- 定时任务：`internal/task/cron.go` 周期性抓取 Injective 数据并写入 Mongo，为在线查询提供最新素材。
- 外部依赖：`internal/injective/client.go` 封装 Injective HTTP 接口访问与错误处理。

### 典型数据流

- 请求路径：Client -> REST 路由 -> Handler -> ChartLogic -> Redis 缓存（命中返回） -> Mongo（未命中回源） -> 返回 -> 回填 Redis。
- 定时路径：Cron Tick -> Injective API 抓取 -> 写入 Mongo（summary_all / config / summary）-> 在线查询命中更快。

### 配置项（摘录）

- `Redis`: `Address`, `Password`, `DB`
- `Mongo`: `URI`, `Database`, `Collections.{Spot,Derivative}`
- `Injective`: `BaseURL`, 多个 *Path，`TimeoutMs`
- `Cron`: `Enabled`, `IntervalSec`

### 如何预览 Mermaid

- 本地：安装 VS Code 插件“Mermaid Markdown”或“Markdown Preview Mermaid Support”。
- 在线：使用 [Mermaid Live Editor](https://mermaid.live) 粘贴上述代码预览/导出 PNG、SVG。

---

文档来源于代码自动梳理，若路由、任务或配置有更新，建议同步修订本文件。


