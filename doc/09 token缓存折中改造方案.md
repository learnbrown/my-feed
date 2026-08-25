# JWT Token 缓存折中改造方案

> 状态（2026-08-25）：已编码实施并通过默认测试。当前使用短 TTL、MySQL 优先写入、mismatch 回源和 Redis 故障降级方案。

## 1. 目标与取舍

当前项目使用 MySQL `accounts.token` 保存账号当前有效 token，并用 Redis 减少鉴权时的 MySQL 查询。

本次改造面向只在本地运行的个人单体项目，选择“主动失效 + 短 TTL”的有界最终一致性方案。

目标：

- MySQL 始终是 token 最终数据源。
- Redis 命中时保留鉴权快速路径。
- Redis miss、超时或连接错误时回源 MySQL。
- 登录和登出不因 Redis 写失败而失败。
- 正常登录、登出后主动更新或删除缓存。
- 异常情况下，用 5 分钟 TTL 限制旧 token 缓存窗口。
- 方案保持单体项目可理解、可测试，不引入账号锁、分布式锁或 MQ。

边界：Redis 写失败或极端并发时，旧 token 不保证严格实时失效。

## 2. 一致性边界

正常路径：

```text
登录成功 -> MySQL 写新 token -> Redis SET 新 token
登出成功 -> MySQL 清空 token -> Redis DEL token
```

Redis 操作成功后，旧 token 正常情况下立即失效。

异常路径：

```text
Redis SET/DEL 失败
  -> MySQL 状态仍然成功提交
  -> 接口仍然成功
  -> Redis 旧值等待短 TTL 过期
```

token cache TTL：

```text
min(JWT 剩余有效期, 5 分钟)
```

因此单个旧缓存值的存活时间最多为 5 分钟。极端并发下，如果一个较早开始的鉴权请求在登录或登出之后才回填旧值，窗口从最后一次旧值回填开始计算；当前项目接受这个边界。

缓存 TTL 不是滑动过期时间。Redis hit 时不能续期，否则旧 token 窗口可能被不断延长。

## 3. Redis Key

继续使用当前 key：

```text
myfeed:account:token:{account_id}
```

不修改 key 结构。实施前手动清理本地 token 测试数据，之后新登录会用 `SET` 覆盖旧值并写入新 TTL。无需为此增加业务代码。

## 4. 详细流程

### 4.1 登录

```text
1. 查询账号并校验密码。
2. 生成 JWT。
3. 更新 MySQL：accounts.token = newToken。
4. 尝试 Redis SET，TTL = min(JWT 剩余时间, 5 分钟)。
5. Redis 写成功或失败都返回登录成功；失败时记录日志。
```

不再在登录前删除 Redis key。`SET` 本身会原子覆盖旧值；如果 `SET` 失败，旧值最多保留到 5 分钟 TTL 结束。

MySQL 更新失败时不写 Redis，也不返回 token。

### 4.2 鉴权

```text
1. 解析并验证 JWT 签名和过期时间。
2. 读取 Redis token key。
3. Redis hit 且 cachedToken == requestToken：直接通过。
4. Redis miss 或读取异常：查询 MySQL。
5. Redis hit 但 cachedToken != requestToken：也查询 MySQL，不直接返回 401。
6. MySQL token == requestToken：鉴权通过，尝试用短 TTL 回填 Redis。
7. MySQL token != requestToken、为空或账号不存在：返回 401。
```

缓存不一致时回查 MySQL 非常重要。例如连续登录后 Redis `SET` 失败，缓存里可能还是旧 token；新 token 请求必须通过 MySQL 被确认并覆盖缓存，不能被旧缓存直接拒绝。

回填失败不影响本次鉴权，因为本次请求已经由 MySQL 完成可信校验。

### 4.3 登出

```text
1. 清空 MySQL：accounts.token = ""。
2. MySQL 成功后尝试 Redis DEL。
3. Redis 删除成功或失败都返回登出成功；失败时记录日志。
```

必须先写 MySQL。MySQL 更新失败时不能删除 Redis，也不能返回登出成功。

正常情况下 `DEL` 成功，旧 token 立即失效。`DEL` 失败时，旧缓存可能继续命中，但最多存活 5 分钟。

## 5. TTL 计算

定义最大 TTL：

```go
const MaxTokenCacheTTL = 5 * time.Minute
```

统一计算函数：

```text
remaining = expiresAt - now
remaining <= 0             -> 不写缓存
remaining < 5 分钟          -> TTL = remaining
remaining >= 5 分钟         -> TTL = 5 分钟
```

登录生成 JWT 时通过 `auth.GenerateToken` 同时得到 token 和 `ExpiresAt`。鉴权回填直接使用解析后的 `claims.ExpiresAt`。

TTL 计算已集中在 `CalculateTokenCacheTTL` 中，登录和 JWT middleware 共用同一套规则。

## 6. 已完成的代码改动

### `internal/auth/jwtToken.go`

- 将 `GenerateToken` 改为同时返回 token 和 `ExpiresAt`，并同步它的现有调用点和测试。
- JWT 有效期仍为 2 小时，缓存上限独立为 5 分钟。

### `internal/account/token_cache.go`

- 保留 `TokenCache` 的 `GetToken/SetToken/DelToken` 小接口。
- 增加统一的 `CalculateTokenCacheTTL` 函数。
- `SetToken` 收到非正 TTL 时不写 Redis。

### `internal/account/service.go`

- 删除登录前的缓存 `DEL`。
- 登录改为 MySQL 更新成功后尝试 `SET`。
- 登出改为先清空 MySQL，再尝试 `DEL`。
- Redis `SET/DEL` 失败只记录日志，不返回业务错误。
- 删除 `ErrDelCacheFailed` 及相关错误分支。
- 修正当前 `cacha/cacahe` 和误写成 detail cache 的日志文本。

### `internal/middleware/jwt/jwtAuth.go`

- Redis hit 且 token 相同时保留快速放行。
- Redis hit 但 token 不同时改为查询 MySQL 兜底，不直接返回 `401`。
- miss、读取异常和 mismatch 共用同一段 MySQL 校验代码。
- 回填使用 `min(JWT 剩余有效期, 5 分钟)`。
- Redis 回填失败只记录日志，不影响已通过的请求。

### 日志收口涉及文件

- `cmd/main.go`：保留启动阶段日志，统一 Redis 启用/降级用词。
- `internal/auth/jwtToken.go`：不再把“记录后继续运行”的分支标记为 `FATAL`。
- `internal/account/service.go` 和 `internal/middleware/jwt/jwtAuth.go`：删除 token cache hit/miss/success 日志，`ErrDisabled` 静默跳过，异常日志补 `account_id`。
- `internal/video/service.go`：删除 detail cache hit/miss/set success 日志，异常日志补 `video_id`。
- `internal/like/service.go` 和 `internal/comment/service.go`：删除 detail cache delete success 日志，删除失败时保留带 `video_id` 的降级日志。

不新增通用 logger 封装和依赖注入；对当前体量来说，统一文案和字段已足够。

### 测试

已更新或新增：

1. TTL 表驱动测试：剩余时间大于 5 分钟、小于 5 分钟、已过期。
2. 登录场景：正常 `SET`、Redis `SET` 失败仍成功、MySQL 失败时不写 Redis。
3. 登出场景：正常 `DEL`、Redis `DEL` 失败仍成功、MySQL 失败时不删 Redis。
4. mismatch 场景：MySQL 匹配时通过并覆盖缓存，MySQL 也不匹配时返回 `401`。
5. Redis miss 和读取错误都能回源 MySQL。
6. 连续登录时旧 token 失效、新 token 可用。
7. 登出场景：正常删除后旧 token 失效；`DEL` 失败时，短 TTL 过期后回源 MySQL 并返回 `401`。
8. miniredis 验证 token key 命中时不续期。

## 7. 日志与安全

当前启动日志基本合理：配置、MySQL、Redis 和 HTTP 服务器的启动结果都可见，`gin.Default()` 也已提供每次 HTTP 请求的访问日志和 recovery。

改造前的业务日志有以下问题：

- token 和视频详情缓存在每次 hit、miss、set 成功、delete 成功时都输出，会淹没真正的异常。
- `cache.ErrDisabled` 是预期的降级状态，启动时已记录一次，不应在每个请求里反复打印。
- 错误日志大多没有 `account_id` 或 `video_id`，出问题时难以定位对象。
- `cacha/cacahe/caches` 等拼写不一致，token cache 还有误写成 detail cache 的日志。
- `log.Printf("FATAL: ...")` 只打印而不退出，与 `FATAL` 语义不符。
- `gin.Default()` 的访问日志能看到 `500`，但看不到内部错误原因；当前多个 handler 反而将 `err.Error()` 直接返回响应。

本地个人项目没有引入 `slog`、zap、日志文件轮转、request ID 或链路追踪，继续使用标准库 `log`。已按以下规则收口：

1. 保留启动阶段结果：配置加载、MySQL 连接、Redis 启用/降级、HTTP 服务器失败。
2. 删除缓存 hit、miss 和写入/删除成功日志；这些是正常高频路径。
3. `cache.ErrDisabled` 在业务层静默跳过；只在启动时记录“缓存未启用”。
4. 保留 Redis 运行时错误，统一为简单 `key=value` 格式，便于本地搜索。
5. 业务可预期错误，如密码错误、token 无效、记录不存在，不额外打错误日志；Gin 访问日志已能看到 HTTP 状态码。
6. `FATAL` 只用于随后确实退出的启动错误。继续运行的分支使用 `WARN` 或 `ERROR`。

缓存运行时错误日志示例：

```text
level=WARN component=token_cache operation=set account_id=12 err="context deadline exceeded"
level=WARN component=video_detail_cache operation=delete video_id=35 err="connection refused"
```

日志可以包含 `account_id`、`video_id`、`operation` 和 `err`，但绝不能打印完整 JWT、Authorization header、密码、password hash、MySQL DSN 或 Redis 密码。当前代码没有记录这些敏感数据，这一点应保持。

Redis 错误属于降级事件，不应输出“登录失败”或“登出失败”。当前也不为了统计 hit rate 把高频事件打成日志；如果后续真要做性能观测，应使用计数指标而不是逐请求日志。

handler 的 `500` 错误输出放到后续独立收口：在 HTTP 边界只记录一次内部错误，并向客户端返回统一的 `internal server error`；预期的 `400/401/404` 不打 error 日志。这会涉及所有 handler，不和本次 token/cache 修改混在一起。

## 8. 验收标准

- Redis 可用时，重复鉴权可以命中缓存，不查询 MySQL。
- Redis 不可用时，登录、鉴权和登出仍可使用 MySQL 完成。
- 正常连续登录后，旧 token 被拒绝，新 token 可用。
- 正常登出后旧 token 被拒绝。
- Redis `SET/DEL` 故障不改变 MySQL 成功结果和 HTTP 成功响应。
- token key TTL 始终不超过 5 分钟，也不超过 JWT 剩余有效期。
- 缓存 mismatch 会回查 MySQL，不会误拒绝数据库中的当前 token。
- 文档明确说明异常情况下存在有界旧 token 窗口，不宣称严格实时撤销。
- 高频缓存 hit/miss/success 和 `cache.ErrDisabled` 不再逐请求打印。
- 缓存运行时错误包含 component、operation 和业务 ID，且不包含 token 等敏感数据。

## 9. 实施结果

```text
1. `[已完成]` 增加统一 TTL 计算和 TTL 测试。
2. `[已完成]` 调整登录、登出写入顺序和 Redis 错误语义。
3. `[已完成]` 调整 JWTAuth mismatch 回源和短 TTL 回填。
4. `[已完成]` 删除高频成功日志，统一 Redis 异常日志。
5. `[已完成]` 补 service、middleware 和 miniredis 测试，默认 `go test ./...` 通过。
6. `[下一项]` 用户主页 60 秒 Cache Aside 和主动失效。
```
