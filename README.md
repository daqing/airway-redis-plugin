# airway-redis-plugin

The shared Redis integration for [Airway](https://github.com/daqing/airway)
applications: environment-based configuration, a process-wide client handle,
and a health check. It is the foundation for features like Redis-backed
download aggregation and rate limiting.

[中文](README.zh-CN.md)

It complements the framework's `lib/redis_client`: the same `REDIS` URL
convention, but `Setup` verifies connectivity with `PING`, and absent
configuration is a first-class state (`Enabled` reports false) instead of a
missing client.

## Usage

Configure via environment (`AIRWAY_REDIS` wins over `REDIS`):

```
REDIS="redis://localhost:6379/0"
```

At boot, when Redis is optional:

```go
import "github.com/daqing/airway-redis-plugin"

cfg := redis.FromEnv()
if cfg.Enabled() {
    if err := redis.Setup(cfg); err != nil {
        log.Fatalf("redis setup failed: %v", err)
    }
}
```

Elsewhere in the app:

```go
rdb := redis.Must()                        // panics without Setup
rdb := redis.Current()                     // nil without Setup
err := redis.Ping(ctx)                     // health check
```

## API

- `Config{URL string}` — `redis://[username:password@]host:port/db`
- `FromEnv() Config`, `Config.Enabled() bool`
- `Setup(Config) error` — parses the URL, builds the client, verifies with PING
- `Current() *redis.Client` / `Must() *redis.Client`
- `Ping(ctx) error`

Depends on `github.com/redis/go-redis/v9` (same version the Airway framework
uses). Tests run against [miniredis](https://github.com/alicebob/miniredis),
no real Redis needed:

```bash
go test ./...
```
