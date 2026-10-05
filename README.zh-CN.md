# airway-redis-plugin

[Airway](https://github.com/daqing/airway) 应用的统一 Redis 集成:环境变量配置、
进程级全局句柄、健康检查。它是"下载统计 Redis 聚合""限流 Redis 后端"这类功能的
地基。

[English](README.md)

它对框架的 `lib/redis_client` 做薄封装并对齐命名:同样的 `REDIS` URL 约定,但
`Setup` 会用 `PING` 验证连通性,"未配置"是一等状态(`Enabled` 返回 false)而不是
一个缺失的 client。

## 用法

通过环境变量配置(`AIRWAY_REDIS` 优先于 `REDIS`):

```
REDIS="redis://localhost:6379/0"
```

启动时(Redis 可选的场景):

```go
import "github.com/daqing/airway-redis-plugin"

cfg := redis.FromEnv()
if cfg.Enabled() {
    if err := redis.Setup(cfg); err != nil {
        log.Fatalf("redis setup failed: %v", err)
    }
}
```

应用内其他位置:

```go
rdb := redis.Must()                        // 未 Setup 时 panic
rdb := redis.Current()                     // 未 Setup 时返回 nil
err := redis.Ping(ctx)                     // 健康检查
```

## API

- `Config{URL string}`——`redis://[username:password@]host:port/db`
- `FromEnv() Config`、`Config.Enabled() bool`
- `Setup(Config) error`——解析 URL、建立 client、PING 验证
- `Current() *redis.Client` / `Must() *redis.Client`
- `Ping(ctx) error`

依赖 `github.com/redis/go-redis/v9`(与 Airway 框架同版本)。测试基于
[miniredis](https://github.com/alicebob/miniredis),不需要真实 Redis:

```bash
go test ./...
```
