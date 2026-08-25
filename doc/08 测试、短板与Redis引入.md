# 当前项目测试、短板识别与 Redis 引入分析

本文面向当前 `my-feed` 仓库，不是照搬原项目的测试清单。当前项目已经完成 V0.3 MVP，并正在推进 V1.0 Redis 高并发读优化。Redis 基础设施、JWT token 缓存基础读路径和视频详情 Cache Aside 已落地，当前要判断的是：

1. 当前同步 MySQL 版本是否真的正确、稳定。
2. 哪些问题是业务逻辑短板，哪些问题是性能短板。
3. 已有 Redis 实现是否满足一致性、降级和可测试性要求。
4. 下一项主页缓存和后续 Feed ZSET 应该如何建立性能证据。

## 1. 当前基线结论

截至 2026-08-25，仓库包含 28 个 `_test.go` 文件和 79 个顶层 `TestXxx` 函数。已做过的本地验证：

```bash
env GOCACHE=/tmp/go-build-cache go test ./...
env GOCACHE=/tmp/go-build-cache go vet ./...
```

结果：默认 `go test ./...` 通过，单元测试、handler 测试和 miniredis 测试会执行。`internal/integration` 中依赖 MySQL 的测试只有设置 `RUN_MYSQL_TESTS=1` 才会执行，否则主动 `Skip`；因此“默认测试通过”不等于真实 MySQL 行为已持续验证。

当前已有覆盖包括：

- JWT 生成解析、缓存命中、mismatch 回源和旧值短 TTL 失效。
- Redis key、token cache 和视频详情 cache 生命周期。
- service 参数校验、非法游标和部分 HTTP 状态码。
- 显式启用 MySQL 后的点赞幂等、评论计数与权限、关注、私信、分页和 Cache Aside 回源。

仍不能仅凭默认测试证明：

- MySQL 集成测试是否在当前开发数据库版本上持续通过。
- 真实 Redis 进程在网络中断、超时时的完整 HTTP 故障链路是否与 stub/miniredis 测试一致。
- Redis 前后的 QPS、p95、p99 和 MySQL 查询次数有何变化。
- API 文档、前端契约、真实代码是否持续一致。

当前最重要的短板已经从“没有自动化测试”变成：MySQL 集成测试默认不执行、缺少真实 Redis 故障链路验收，以及没有可重复的性能基线。

## 2. 测试体系应该怎么搭

先把概念说清楚：测试不是一种东西。你现在缺的不是某一个神奇工具，而是知道“我想验证什么，就该用哪种工具”。

### 2.0 先建立测试工具地图

| 你想知道什么 | 用什么测 | 是否要启动后端服务 | 是否要真实数据库 | 适合阶段 |
| --- | --- | --- | --- | --- |
| 某个函数的业务规则是否正确 | Go 内置 `testing` | 不需要 | 不一定 | 开发时随手跑 |
| 某个 service 的参数校验是否正确 | Go 内置 `testing` | 不需要 | 不一定 | 最先补 |
| GORM 查询、事务、唯一索引是否正确 | Go 集成测试 + 测试 MySQL | 不需要 | 需要 | 补业务正确性 |
| 路由、状态码、JSON 响应是否正确 | `net/http/httptest` | 不需要监听端口 | 通常需要 | 自动化 API 测试 |
| 手动跑完整业务链路 | Bruno 或 curl | 需要 | 需要 | 联调、演示、验收 |
| 并发下计数和幂等是否正确 | Go 并发测试 | 不需要 | 需要 | 高风险业务 |
| 高并发下接口能扛多少 | `hey` / `wrk` / JMeter | 需要 | 需要 | Redis 前后对比 |
| SQL 为什么慢 | MySQL `EXPLAIN` / 慢查询日志 | 不一定 | 需要 | 定位瓶颈 |

对当前项目，推荐你按这个顺序学：

1. 先用 Bruno 手动跑通一条完整链路，建立“接口是怎么被调用的”这个直觉。
2. 再写最简单的 Go 单元测试，理解 `_test.go`、`TestXxx`、`go test`。
3. 再写数据库集成测试，验证事务、唯一索引、游标分页。
4. 再写 `httptest`，把 HTTP 状态码和 JSON 响应纳入自动化。
5. 最后再做压测，用数据证明 Redis 是否真的有必要。

不要一开始就追求“全自动测试平台”。你现在第一目标是：每写完一个功能，都知道用什么方式证明它没坏。

### 2.0.1 Go 测试到底是什么

Go 自带测试工具，不需要额外安装框架。只要文件名以 `_test.go` 结尾，并且函数长这样：

```go
func TestSomething(t *testing.T) {
    // 准备数据
    // 调用被测试函数
    // 判断结果是否符合预期
}
```

然后执行：

```bash
env GOCACHE=/tmp/go-build-cache go test ./...
```

Go 会自动扫描所有 `_test.go` 文件并执行里面的 `TestXxx` 函数。

例如你可以先写一个完全不需要数据库的测试，验证 `feed.ListLatest` 的非法游标判断。文件位置：

```text
internal/feed/service_test.go
```

示例：

```go
package feed

import (
    "errors"
    "testing"
    "time"
)

func TestListLatestRejectsHalfCursor(t *testing.T) {
    service := &FeedService{}

    _, err := service.ListLatest(20, time.UnixMilli(1000), 0)

    if !errors.Is(err, ErrInvalidCursor) {
        t.Fatalf("expected ErrInvalidCursor, got %v", err)
    }
}
```

这个测试的价值在于：它不用启动服务，不用连 MySQL，也不依赖浏览器。它只验证一个业务规则：

```text
传了 latest_time，却没传 latest_id，必须报非法游标。
```

运行单个包：

```bash
env GOCACHE=/tmp/go-build-cache go test ./internal/feed -v
```

只运行一个测试：

```bash
env GOCACHE=/tmp/go-build-cache go test ./internal/feed -run TestListLatestRejectsHalfCursor -v
```

这就是最小的自动化测试。

### 2.0.2 Bruno 到底测什么

Bruno 是 API 客户端，类似 Postman。它不是 Go 测试框架，它做的是：

```text
把真实 HTTP 请求发给正在运行的后端服务
```

所以 Bruno 测试必须先启动后端：

```bash
go run ./cmd/main.go
```

当前项目默认读取：

```text
configs/config.yaml
```

默认数据库配置是：

```text
host: localhost
port: 3306
user: dev_user
password: qwerdf
dbname: myfeed
```

因此 Bruno 适合回答这些问题：

- 路由 URL 写对了吗？
- 请求 JSON 应该怎么填？
- 登录 token 应该放在哪个 Header？
- 上传接口 `multipart/form-data` 怎么发？
- 一个完整业务链路能不能跑通？
- 登出后旧 token 是否返回 `401`？

Bruno 不适合回答这些问题：

- 20 个并发点赞会不会导致计数错。
- 某个 service 的边界条件有没有遗漏。
- 每次改代码后是否自动防止回归。

这些应该交给 Go 自动化测试。

### 2.0.3 当前项目 Bruno 最小链路

Bruno 里先建一个环境，例如 `local`，放这些变量：

```text
baseUrl=http://localhost:8080
tokenA=
tokenB=
accountIdA=
accountIdB=
videoId=
playUrl=
coverUrl=
commentId=
```

然后按顺序建请求。

#### 1. 注册用户 A

```text
POST {{baseUrl}}/account/register
Content-Type: application/json
```

Body:

```json
{
  "username": "user_a_001",
  "password": "123456"
}
```

期望：

```text
201 Created
```

#### 2. 登录用户 A

```text
POST {{baseUrl}}/account/login
Content-Type: application/json
```

Body:

```json
{
  "username": "user_a_001",
  "password": "123456"
}
```

期望响应里有：

```json
{
  "token": "xxx",
  "account": {
    "id": 1
  }
}
```

在 Bruno 的 post-response script 里保存变量：

```js
bru.setEnvVar("tokenA", res.body.token);
bru.setEnvVar("accountIdA", res.body.account.id);
```

之后需要用户 A 登录的请求，Header 写：

```text
Authorization: Bearer {{tokenA}}
```

#### 3. 验证当前用户

```text
GET {{baseUrl}}/account/me
Authorization: Bearer {{tokenA}}
```

期望：

```text
200 OK
```

响应里的 `id` 应该等于 `accountIdA`。

#### 4. 发布视频

如果你只是测试发布逻辑，不想先处理上传文件，可以直接使用合法前缀的测试 URL。当前 service 主要校验 URL 前缀：

```text
play_url 必须以 /static/uploads/videos/ 开头
cover_url 必须以 /static/uploads/covers/ 开头
```

请求：

```text
POST {{baseUrl}}/video/publish
Authorization: Bearer {{tokenA}}
Content-Type: application/json
```

Body:

```json
{
  "title": "First test video #go #gin",
  "description": "测试发布链路 #feed",
  "play_url": "/static/uploads/videos/test/demo.mp4",
  "cover_url": "/static/uploads/covers/default.png"
}
```

在 post-response script 里保存视频 ID：

```js
bru.setEnvVar("videoId", res.body.video.id);
```

如果你要测试上传链路本身，就必须改用：

- `/video/uploadVideo`，Body 选择 `multipart/form-data`，字段名是 `video`。
- `/video/uploadCover`，Body 选择 `multipart/form-data`，字段名是 `cover`。

上传成功后，再把返回的 `play_url` 和 `cover_url` 传给 `/video/publish`。

#### 5. 查询视频详情

```text
POST {{baseUrl}}/video/getDetail
Content-Type: application/json
```

Body:

```json
{
  "id": {{videoId}}
}
```

期望：

```text
200 OK
```

响应里的 `video.id` 应该等于 `videoId`。

#### 6. 查询最新 Feed

```text
POST {{baseUrl}}/feed/listLatest
Content-Type: application/json
```

Body:

```json
{
  "limit": 10,
  "latest_time": 0,
  "latest_id": 0
}
```

期望：

- `200 OK`
- `videos` 数组里能看到刚发布的视频
- 响应里有 `next_time`
- 响应里应该有 `next_id`

文档示例必须同时包含 `next_time` 和 `next_id`，并通过 HTTP 响应断言避免再次漂移。

#### 7. 注册并登录用户 B

重复用户 A 的注册、登录流程，但变量保存成：

```js
bru.setEnvVar("tokenB", res.body.token);
bru.setEnvVar("accountIdB", res.body.account.id);
```

#### 8. 用户 B 点赞

```text
POST {{baseUrl}}/like/like
Authorization: Bearer {{tokenB}}
Content-Type: application/json
```

Body:

```json
{
  "video_id": {{videoId}}
}
```

连续点两次。期望：

- 两次都应该成功。
- 数据库里只应该有一条点赞记录。
- 视频 `likes_count` 只应该加 1。

Bruno 只能帮你发两次请求；“数据库里是不是只有一条”需要用数据库查询或 Go 测试验证。

#### 9. 用户 B 评论

```text
POST {{baseUrl}}/comment/publish
Authorization: Bearer {{tokenB}}
Content-Type: application/json
```

Body:

```json
{
  "video_id": {{videoId}},
  "content": "这是一条测试评论"
}
```

保存评论 ID：

```js
bru.setEnvVar("commentId", res.body.comment.id);
```

然后查询评论列表：

```text
POST {{baseUrl}}/comment/listComment
Content-Type: application/json
```

Body:

```json
{
  "video_id": {{videoId}},
  "limit": 20,
  "latest_time": 0,
  "latest_id": 0
}
```

#### 10. 用户 B 关注用户 A

```text
POST {{baseUrl}}/follow/follow
Authorization: Bearer {{tokenB}}
Content-Type: application/json
```

Body:

```json
{
  "vlogger_id": {{accountIdA}}
}
```

连续关注两次。期望：

- 两次都应该成功。
- 数据库里只应该有一条关注关系。

#### 11. 用户主页聚合

```text
POST {{baseUrl}}/account/getProfile
Content-Type: application/json
```

Body:

```json
{
  "account_id": {{accountIdA}}
}
```

重点看：

- `videos_count` 是否至少为 1。
- `likes_count` 是否包含用户 B 的点赞。
- `followers_count` 是否包含用户 B 的关注。

#### 12. 登出和旧 token 失效

```text
POST {{baseUrl}}/account/logout
Authorization: Bearer {{tokenA}}
```

然后再次请求：

```text
GET {{baseUrl}}/account/me
Authorization: Bearer {{tokenA}}
```

期望：

```text
401 Unauthorized
```

这一步很重要。当前项目的 JWT 不是纯无状态，数据库里的 `accounts.token` 是服务端撤销 token 的依据。现在已经加入 Redis，测试必须继续证明缓存没有破坏这个语义。

### 2.0.4 curl 是什么位置

curl 是命令行版 API 客户端，适合快速确认一个接口，不适合管理复杂链路。

例如注册：

```bash
curl -i -X POST http://localhost:8080/account/register \
  -H "Content-Type: application/json" \
  -d '{"username":"curl_user_001","password":"123456"}'
```

登录：

```bash
curl -i -X POST http://localhost:8080/account/login \
  -H "Content-Type: application/json" \
  -d '{"username":"curl_user_001","password":"123456"}'
```

curl 的问题是 token 保存、变量传递、批量执行都麻烦。所以当前项目日常手测建议用 Bruno。

### 2.0.5 `httptest` 是什么位置

`httptest` 是 Go 标准库里的 HTTP 测试工具。它不需要你真的执行：

```bash
go run ./cmd/main.go
```

它是在测试代码里构造一个 HTTP 请求，直接打到 Gin router 上：

```text
构造请求 -> router.ServeHTTP -> 读取响应状态码和响应体 -> 断言
```

概念示例：

```go
func TestRegisterRequiresPassword(t *testing.T) {
    router := setupTestRouter(t)

    req := httptest.NewRequest(
        http.MethodPost,
        "/account/register",
        strings.NewReader(`{"username":"u1"}`),
    )
    req.Header.Set("Content-Type", "application/json")

    recorder := httptest.NewRecorder()
    router.ServeHTTP(recorder, req)

    if recorder.Code != http.StatusBadRequest {
        t.Fatalf("expected 400, got %d, body=%s", recorder.Code, recorder.Body.String())
    }
}
```

这个测试适合验证：

- URL 是否存在。
- Method 是否正确。
- Header 是否正确。
- JSON 绑定是否正确。
- 状态码是否符合文档。
- 响应字段是否符合前端契约。

当前项目的 `router.SetRouter(sqlDB, redisCache)` 需要传入真实 `*gorm.DB`；测试可以把第二个参数传 `nil` 验证 Redis 禁用时的 HTTP 链路。MySQL 集成测试已经通过 `internal/db/testutil` 强制要求数据库名以 `_test` 结尾，例如：

```text
myfeed_test
```

不要直接在开发库 `myfeed` 上跑自动化测试。设置 `RUN_MYSQL_TESTS=1` 后，测试会清空测试表；数据库名不以 `_test` 结尾时会拒绝执行。

### 2.0.6 数据库集成测试怎么理解

数据库集成测试不是为了证明 GORM 没 bug，而是证明你写的查询条件、事务和索引符合业务。

例子：点赞必须满足：

```text
likes 表只插入一条记录
videos.likes_count 只增加 1
重复点赞返回成功，但不重复计数
```

这类问题不能只靠 service 单元测试，因为真正兜底的是 MySQL 唯一索引和事务。

推荐做法：

1. 建一个测试库，例如 `myfeed_test`。
2. 测试启动时连接测试库。
3. 执行 `AutoMigrate`。
4. 每个测试前清空相关表。
5. 插入准备数据。
6. 调用 service 或 repository。
7. 查询数据库，确认最终状态。

测试库清理要注意外键和依赖顺序。当前项目没有显式外键约束，但仍建议按业务依赖从子表到主表清理：

```text
messages
comments
likes
follows
video_tags
tags
videos
accounts
```

### 2.0.7 并发测试怎么理解

并发测试不是压测。并发测试关注的是“同时发生时数据会不会错”。

点赞例子：

```go
func TestLikeConcurrentIdempotent(t *testing.T) {
    // 1. 准备测试 DB、用户、视频
    // 2. 启动 20 个 goroutine 同时调用 Like(accountID, videoID)
    // 3. 等全部结束
    // 4. 查询 likes 表，应该只有 1 条
    // 5. 查询 videos.likes_count，应该只加 1
}
```

这种测试能暴露：

- 是否缺唯一索引。
- 是否事务边界不对。
- 是否重复请求导致计数重复增加。
- 是否取消点赞导致计数变负。

Redis 不能替你修这些问题。必须先让 MySQL 同步版本在并发下正确。

### 2.0.8 压测怎么理解

压测关注的是：

```text
在一定并发下，接口有多快、会不会错、MySQL 压力多大
```

这和并发正确性测试不是一回事。

最简单可以用 `hey`：

```bash
hey -z 60s -c 50 \
  -m POST \
  -H "Content-Type: application/json" \
  -d '{"limit":20,"latest_time":0,"latest_id":0}' \
  http://localhost:8080/feed/listLatest
```

含义：

- `-z 60s`：持续压 60 秒。
- `-c 50`：50 个并发 worker。
- `-m POST`：请求方法是 POST。
- `-H`：请求头。
- `-d`：请求体。

你要记录的不是“看起来能跑”，而是：

```text
Requests/sec
Average latency
p95 / p99 latency
非 2xx 响应数量
MySQL CPU
MySQL 连接数
慢 SQL
```

Redis 前做一次，Redis 后做一次。这样你才能说：

```text
引入 Redis 后，/feed/listLatest 的 p95 从多少降到多少，MySQL 查询次数从多少降到多少。
```

没有这个对比，就不要说“Redis 提升了性能”。

### 2.0.9 当前测试覆盖与下一批重点

下面三批测试已经开始落地，不再是完全空白的规划。保留清单用于继续补覆盖率。

#### 第一批：不需要数据库的 service 参数校验测试，已覆盖主要规则

这些最容易写：

```text
internal/feed/service_test.go
- ListLatest 只传 latest_time 不传 latest_id -> ErrInvalidCursor
- ListLatest 只传 latest_id 不传 latest_time -> ErrInvalidCursor
- ListByTag tag 为空 -> ErrTagNameRequired

internal/video/service_test.go
- Publish title 为空 -> ErrTitleRequired
- Publish play_url 为空 -> ErrPlayURLRequired
- Publish play_url 前缀非法 -> ErrInvalidPlayURL
- Publish cover_url 前缀非法 -> ErrInvalidCoverURL

internal/comment/service_test.go
- CreateComment content 为空 -> ErrContentRequired
- CreateComment content 超过 500 字 -> ErrContentTooLarge

internal/follow/service_test.go
- Follow 自己 -> ErrSelfFollowing
- IsFollowing 自己 -> false
```

这些测试大多在进入数据库前就返回错误，所以不需要 MySQL。当前 account、video、feed、comment、follow、message、profile 等模块均已有 `_test.go`，后续重点是避免新增规则没有对应测试。

#### 第二批：需要测试 MySQL 的业务正确性测试，已落地但默认跳过

```text
like
- 重复点赞只产生一条 like，likes_count 只加 1
- 重复取消点赞不报错，likes_count 不为负

follow
- 重复关注只产生一条 follow
- 重复取关不报错

comment
- 创建评论后 comments_count +1
- 删除评论后 comments_count -1
- 用户不能删除别人的评论

video/feed
- 同一 created_at 下分页不漏、不重
- status != 1 的视频不出现在列表
```

#### 第三批：HTTP 状态码和响应结构测试，已覆盖部分主链路

```text
account
- register 缺 password -> 400
- login 密码错误 -> 401
- logout 后 me -> 401

feed
- listLatest 正常请求 -> 200
- listByTag 缺 tag_name -> 400
- 非法游标 -> 400

video
- publish 未登录 -> 401
- getDetail 不存在 -> 404
```

这批测试能防止“service 是对的，但 handler 状态码写错了”。当前已有账号、视频、Feed 及部分完整 API 链路测试，仍需继续补齐缓存故障降级时成功响应不变的契约。

当前 Feed 和视频 handler 已覆盖负数游标；其余列表接口也要保持同样断言。handler 如果返回错误后没有 `return`，就可能继续执行后面的正常逻辑，测试应该能抓住这种问题。

### 2.1 编译、格式、静态检查

这是最低门槛，只能证明代码能过工具检查。

建议固定为每次改动后的基础命令：

```bash
gofmt -w internal cmd
env GOCACHE=/tmp/go-build-cache go test ./...
env GOCACHE=/tmp/go-build-cache go vet ./...
```

注意：`go test ./...` 现在会执行真实业务单元测试和 miniredis 测试，但 MySQL 集成测试默认跳过。需要分别记录“默认测试通过”和“显式 MySQL 集成测试通过”，不能混为一个结论。

显式执行 MySQL 集成测试：

```bash
RUN_MYSQL_TESTS=1 \
MYSQL_DBNAME=myfeed_test \
env GOCACHE=/tmp/go-build-cache go test ./internal/integration -v
```

### 2.2 Service 层单元测试

目标：不经过 HTTP，直接验证业务规则。

优先覆盖：

| 模块 | 重点用例 |
| --- | --- |
| `account` | 注册空用户名、空密码、重复用户名；登录错误密码；登出清空 token |
| `video` | 标题必填、播放地址前缀校验、默认封面、标签提取和去重、非法游标 |
| `feed` | `limit` 默认值和最大值、非法游标组合、空 tag |
| `like` | 重复点赞幂等、重复取消点赞幂等、视频不存在、计数不为负 |
| `comment` | 空评论、超长评论、删除非本人评论、计数回滚 |
| `follow` | 不能关注自己、重复关注幂等、取关不存在关系幂等、用户不存在 |
| `message` | 不能给自己发消息、空消息、超长消息、会话分页 |
| `profile` | 用户不存在、只统计正常视频、粉丝/关注/获赞聚合 |

当前 service 多数直接依赖具体 repository 类型，短期可以先用测试数据库做 service 集成测试。中长期如果要提高单测隔离度，可以把 repository 抽成小接口，但不要为了测试框架过早大改结构。

### 2.3 Repository 与数据库集成测试

这类测试要连接测试 MySQL，重点不是测 GORM，而是测 SQL 条件、事务、唯一索引和游标。

必须覆盖：

- `videos.status = 1` 过滤是否生效。
- `created_at DESC, id DESC` 顺序是否稳定。
- 多条记录 `created_at` 相同时，`latest_time + latest_id` 是否不漏数据。
- 点赞唯一索引是否阻止重复记录。
- 关注唯一索引是否阻止重复关系。
- 评论软删除后列表是否不再返回。
- `likes_count`、`comments_count` 的增减和明细表是否一致。

建议每个数据库集成测试都用独立测试库或事务回滚，避免污染本地开发数据。

### 2.4 HTTP/API 集成测试

目标：验证路由、中间件、参数绑定、状态码、响应结构和 service 是否串起来。

建议用 `httptest` 写 Go 集成测试，Bruno 作为人工验收和演示工具。最小闭环如下：

1. `POST /account/register` 注册用户 A、用户 B。
2. `POST /account/login` 登录用户 A，保存 `tokenA`、`accountIdA`。
3. `GET /account/me` 使用 `Authorization: Bearer <tokenA>`，应返回当前用户。
4. `POST /video/uploadVideo` 上传测试视频，拿到 `play_url`。
5. `POST /video/uploadCover` 上传封面，拿到 `cover_url`。
6. `POST /video/publish` 发布视频。
7. `POST /video/getDetail` 查询详情。
8. `POST /feed/listLatest` 查询首页流。
9. `POST /feed/listByTag` 查询标签流。
10. 用户 B 登录后执行点赞、评论、关注、私信。
11. `POST /account/getProfile` 检查作者主页聚合值。
12. `POST /account/logout` 登出用户 A。
13. 再用旧 `tokenA` 请求 `GET /account/me`，必须返回 `401`。

Bruno 环境变量建议至少包含：

```text
baseUrl=http://localhost:8080
tokenA=
tokenB=
accountIdA=
accountIdB=
videoId=
playUrl=
coverUrl=
commentId=
```

这里要强调：Bruno 可以证明一次手动流程能跑通，但不能替代自动化测试。真正防回归的是 Go 测试和 CI 命令。

### 2.5 并发与一致性测试

当前项目里最值得测的不是“接口返回 200”，而是并发下数据是否还正确。

必须设计这些用例：

- 20 个 goroutine 同时点赞同一个视频：`likes` 只新增一条，`likes_count` 最终只加 1。
- 20 个 goroutine 同时取消同一个点赞：`likes_count` 不为负。
- 20 个 goroutine 同时关注同一个用户：`follows` 只新增一条。
- 评论创建成功但计数更新失败时，事务要整体回滚。
- 发布视频时创建 tag 或 video_tag 失败时，视频记录不能半成功。
- 登出和鉴权并发时，旧 token 不能继续通过。

这些测试比 Redis 更重要。因为 Redis 解决不了错误的事务边界和错误的幂等语义。

### 2.6 分页专项测试

所有列表接口都应有同一类测试：

- 第一页：`latest_time = 0, latest_id = 0`。
- 下一页：使用上一页最后一条的 `next_time + next_id`。
- 同一毫秒或同一数据库时间戳下插入多条数据，不允许漏、不允许重复。
- 只传 `latest_time` 不传 `latest_id`，应返回非法游标。
- 只传 `latest_id` 不传 `latest_time`，应返回非法游标。
- `latest_time < 0`，应返回参数错误。
- `limit <= 0` 使用默认值。
- `limit > 50` 截断为最大值。

当前需要覆盖的接口：

- `/video/listByAuthorID`
- `/feed/listLatest`
- `/feed/listByTag`
- `/like/listLikedVideos`
- `/comment/listComment`
- `/follow/listFollower`
- `/follow/listFollowing`
- `/message/listConversation`

### 2.7 性能与压测

性能测试不要一开始追求“很大的 QPS”。先建立基线，然后比较 Redis 前后变化。

建议先准备三档测试数据：

| 档位 | 数据量 | 目的 |
| --- | --- | --- |
| 小 | 10 用户、100 视频、几百互动 | 本地调试和接口验收 |
| 中 | 1,000 用户、10,000 视频、10 万互动 | 暴露普通索引和分页问题 |
| 大 | 10,000 用户、100,000 视频、100 万互动 | 暴露 Feed、详情、主页聚合的读压力 |

优先压测这些接口：

| 接口 | 为什么测 |
| --- | --- |
| `POST /feed/listLatest` | 首页流是最高频读接口，所有用户都可能刷 |
| `POST /video/getDetail` | 热门视频会被重复读取 |
| `POST /account/getProfile` | 聚合多张表，容易造成多次 DB 查询 |
| `POST /like/isLiked` | 前端渲染列表时可能高频查询 |
| `POST /follow/isFollowing` | 用户主页和关注按钮高频查询 |
| 需要鉴权的任意接口 | 当前 JWT 中间件每次都查账号 token |

可以用 `hey` 或 `wrk` 做 HTTP 压测，例如：

```bash
hey -z 60s -c 50 \
  -m POST \
  -H "Content-Type: application/json" \
  -d '{"limit":20,"latest_time":0,"latest_id":0}' \
  http://localhost:8080/feed/listLatest
```

压测时至少记录：

- QPS。
- 平均耗时。
- p95 / p99 延迟。
- 错误率。
- MySQL CPU、连接数、慢查询。
- 关键 SQL 的 `EXPLAIN`。
- 单次请求触发的 DB 查询次数。

如果没有这些指标，就不能严肃地说“需要 Redis”。最多只能说“原项目用了 Redis”。

## 3. 如何识别当前项目短板

### 3.1 正确性短板

表现：

- 已有 28 个 `_test.go` 文件，但依赖 MySQL 的集成测试默认 `Skip`。
- token 缓存已覆盖 5 分钟 TTL、mismatch 回源、Redis 写失败降级和旧值有界失效测试。
- Redis 故障注入已覆盖 token 读写降级，但尚未做真实 Redis 进程的故障链路验收。
- Bruno 主链路和持续集成尚未建立。

解决方向：

- 在固定的 `myfeed_test` 环境持续执行 MySQL 集成测试。
- 手动停止本地 Redis 验收登录、鉴权和登出的 HTTP 降级链路。
- 为主页缓存补 hit、miss、回源、坏 JSON 和主动失效测试。
- 用 Bruno 或等价脚本保留可演示的完整业务链路。

### 3.2 API 契约短板

表现：

- `doc/00 api-概述.md` 已同步为 `latest_time + latest_id` 和 `VideoDTO` 字段，但目前仍依靠人工维护。
- token 缓存折中方案已落地：Redis 写失败不改变 MySQL 成功结果和 HTTP 成功响应。
- 缓存属于内部实现，正常命中、miss 和回源不应改变成功响应 JSON；这一点需要响应断言保护。

解决方向：

- 把 API 文档也当作测试对象。
- 每次接口字段变化后，同步更新 `doc/00 api-概述.md` 和具体接口文档。
- HTTP 集成测试断言响应字段，避免前端和文档继续漂移。

### 3.3 数据库与索引短板

表现：

- Feed、作者列表、标签列表都依赖 `status + created_at + id` 的过滤和排序。
- 评论、点赞、关注、私信列表都依赖关系字段加 `created_at + id` 的排序。
- 当前模型已有部分索引，但应通过 `EXPLAIN` 确认是否完全匹配实际查询。

重点检查：

```text
videos:   (status, created_at, id)
videos:   (author_id, status, created_at, id)
comments: (video_id, created_at, id)
likes:    (account_id, created_at, id)
follows:  (vlogger_id, created_at, id)
follows:  (follower_id, created_at, id)
messages: (from_id, to_id, created_at, id)
messages: (to_id, from_id, created_at, id)
```

判断方式：

- 用中、大数据量压测列表接口。
- 对慢查询执行 `EXPLAIN`。
- 如果出现全表扫描、文件排序、扫描行数过大，先补索引和查询条件，再谈 Redis。

### 3.4 可观测性短板

表现：

- 缺少请求耗时、错误率、慢 SQL、DB 查询次数等指标。
- 压测时只能看到客户端耗时，不知道瓶颈在 Gin、业务逻辑还是 MySQL。

解决方向：

- 至少先加请求日志：method、path、status、latency、account_id。
- MySQL 开启慢查询日志。
- 压测记录机器资源和数据库资源。
- 后续可接 Prometheus/Grafana，但 V1.0 当前阶段不必急。

### 3.5 可测试性短板

表现：

- Router 内部直接组装所有 repo/service/handler，HTTP 测试不容易替换依赖。
- Service 依赖具体 repo 类型，纯 mock 单测不方便。
- 上传接口会写本地 `.run/uploads`，测试时需要隔离目录。

解决方向：

- 短期：用测试 MySQL + `httptest` 做集成测试。
- 中期：为复杂 service 抽最小 repository interface。
- 上传测试使用临时目录或可配置上传根目录。

## 4. 当前 Redis 落点和后续价值

Redis 的价值不是“替代 MySQL”，而是把高频读、排序、短期状态和原子计数从 MySQL 压力路径上拆出来。当前已经落地 token 和视频详情缓存，后续还有几个自然落点。

### 4.1 JWT token 校验

当前鉴权链路已经改为：

```text
解析 JWT -> 查 Redis token
Redis miss/读取异常 -> 查询 accounts -> 比对 accounts.token -> 成功后回填
```

它减少了缓存命中时的 MySQL 查询，同时保留 MySQL 回源。

当前采用适合个人项目的折中方案：

```text
登录：更新 MySQL token -> 尝试 Redis SET
登出：清空 MySQL token -> 尝试 Redis DEL
鉴权：Redis miss/异常/mismatch -> 回查 MySQL
TTL：min(JWT 剩余有效期, 5 分钟)
```

MySQL 是最终数据源，Redis `GET/SET/DEL` 失败都可以降级。正常情况下主动 `SET/DEL` 让登录和登出立即反映到缓存；Redis 写失败或极端并发时接受最长 5 分钟的旧值窗口，不宣称故障场景下严格实时撤销。

缓存 mismatch 必须回查 MySQL，不能直接返回 `401`。否则连续登录后 Redis `SET` 失败时，缓存里的旧 token 会误拒绝数据库中的新 token。

继续使用现有 `myfeed:account:token:{account_id}` key。当前只在本地开发，不增加 key 版本、启动扫描或数据迁移逻辑。完整实施方案见 `doc/09 token缓存折中改造方案.md`。

### 4.2 视频详情缓存

状态：已完成 MVP。

`/video/getDetail` 对热门视频会被重复读取，数据主体变化频率低，适合 Cache Aside：

```text
读 Redis -> miss -> 读 MySQL -> 写 Redis -> 返回
```

失效策略：

- 视频删除、状态变化、信息修改时删除缓存。
- 点赞数、评论数变化后，学习阶段建议先删除详情缓存，不要急着局部更新缓存字段。

### 4.3 用户主页聚合缓存

状态：下一项功能目标。

`/account/getProfile` 会读取账号信息、作品数、获赞数、粉丝数、关注数。这个接口天然比普通详情接口重。

适合策略：

- 第一版固定使用 60 秒 TTL。
- 关注/取关删除双方主页缓存；点赞/取消点赞删除视频作者缓存；发布视频删除作者缓存。
- Redis 读取或回填失败时回源 MySQL 并正常返回。
- 主动失效失败不回滚已提交的 MySQL 事务，记录日志并接受最多 60 秒旧数据。

### 4.4 Feed 最新流

`/feed/listLatest` 是最典型的高频读。所有用户刷首页都可能访问同一批最新视频。

适合使用 Redis ZSET：

```text
key:   myfeed:feed:latest
score: created_at 毫秒时间戳
value: video_id
```

读流程：

```text
ZSET 取 video_id 列表
批量取视频详情缓存
详情 miss 再回源 MySQL
```

注意：当前项目已经使用 `latest_time + latest_id` 双游标。Redis 实现也必须保持稳定分页语义，不能因为用了 ZSET 又退回只按时间戳分页。

### 4.5 热榜

MySQL 不适合每次请求都实时按热度排序大量视频。Redis ZSET 很适合维护热榜：

```text
key:   myfeed:rank:hot
score: popularity
value: video_id
```

V1.0 可以先同步更新 Redis 热榜。V1.5 再考虑 MQ + Worker 异步更新。

### 4.6 限流

登录、上传、发布、评论、私信、点赞都需要基础限流。Redis 的 `INCR + EXPIRE` 适合做固定窗口限流：

```text
myfeed:rate:{scene}:{identifier}
```

例子：

- `myfeed:rate:login:ip:127.0.0.1`
- `myfeed:rate:comment:account:1001`
- `myfeed:rate:upload:account:1001`

被限流时返回 `429 Too Many Requests`。

## 5. 哪些问题不该用 Redis 解决

| 问题 | 正确处理 |
| --- | --- |
| 重复点赞导致多条记录 | 唯一索引 + 事务 + 幂等逻辑 |
| 取消点赞后计数为负 | SQL 条件 + 事务 + 测试 |
| 评论删除权限错误 | service 权限规则 + repository 条件 |
| 分页漏数据/重复数据 | 复合游标和排序修正 |
| API 字段和文档不一致 | 更新文档 + 响应断言测试 |
| SQL 没有合适索引 | 先补索引，验证 `EXPLAIN` |
| 上传路径混乱 | 明确本地路径和访问 URL |

严厉一点说：Redis 不能修复错误的业务模型。它只能让正确的读路径更快，让短期状态更容易维护。

## 6. Redis 引入顺序与验收标准

### 6.1 推荐顺序

1. `[已完成]` Redis 基础设施：配置、client、Ping、统一 key 前缀。
2. `[已完成]` JWT token 缓存：5 分钟 TTL、mismatch 回源、写失败降级和日志收口。
3. `[已完成 MVP]` 视频详情 Cache Aside 和点赞/评论主动失效。
4. `[下一项]` 用户主页 60 秒缓存和主动失效。
5. `[未开始]` Feed 最新流 ZSET。
6. `[未开始]` 热榜 ZSET。
7. `[未开始]` 基础限流。

不要一上来做 MQ，不要一上来异步点赞。同步链路没测稳之前，异步化只会把错误藏到后台。

### 6.2 每个 Redis key 必须回答的问题

每加一个 key，都必须写清楚：

```text
它缓存什么？
什么时候写入？
什么时候失效？
TTL 是多少？
Redis miss 怎么回源？
Redis 挂了怎么办？
是否允许短暂不一致？
```

### 6.3 当前基线验收状态

当前状态：

- `[已完成]` 默认 `go test ./...` 执行单元测试和 miniredis 测试。
- `[已落地但需显式启用]` 点赞、评论、关注、私信、分页的 MySQL 集成测试。
- `[本次已同步]` `doc/00 api-概述.md` 的复合游标和 Video DTO。
- `[未完成]` Bruno 主链路一键或半自动执行。
- `[未完成]` 中等数据量的 MySQL-only 压测及 QPS、p95、p99、慢 SQL 记录。
- `[未完成]` Redis 前后相同数据、相同负载的对比报告。

### 6.4 Redis 后的验收标准

Redis 改造后，不要只看接口还能不能返回 200。要看：

- Redis 可用时，目标接口 p95/p99 是否下降。
- MySQL 查询次数是否下降。
- Redis miss 时是否能回源 MySQL。
- Redis 不可用时登录、鉴权、登出和普通读缓存是否降级可用。
- 正常登出是否主动删除缓存；删除失败时旧值是否在 5 分钟内过期。
- 视频详情、主页、Feed 是否存在不可接受的旧数据。
- 缓存失效逻辑是否有测试覆盖。

## 7. 最小落地路线

建议接下来按这个顺序推进：

1. 实现主页 60 秒 Cache Aside，并在关注、点赞和发布后主动失效。
2. 让 `RUN_MYSQL_TESTS=1` 的测试在固定测试库中可重复执行。
3. 写 Bruno 主链路集合，作为人工验收和演示脚本。
4. 准备测试数据生成脚本，至少支持 1 万视频、10 万互动。
5. 对 `/feed/listLatest`、`/video/getDetail`、`/account/getProfile` 做相同条件下的 MySQL-only/Redis 对比压测。
6. 根据 `EXPLAIN`、慢查询和 DB 查询次数补索引，再实现 Feed ZSET。

最终目标不是“实现原项目也有的 Redis 功能”，而是你能清楚讲出：

- 当前项目哪里慢。
- 你如何测出来的。
- Redis key 如何设计。
- 缓存什么时候失效。
- Redis 挂了系统如何降级。
- Redis 前后性能和 DB 压力差异是多少。
