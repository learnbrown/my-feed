# CLAUDE.MD

## 项目背景

本项目是一个从零手写的 Go 后端练习项目，目标是实现一个“高并发视频 Feed 流系统”，作为求职简历核心项目。当前开发策略不是一次性复刻完整开源项目，而是按里程碑逐步演进：先跑通单体 MVP，再补齐业务闭环，最后引入 Redis、MQ、Worker、SSE 等高并发与异步能力。

参考文档位于仓库根目录：

- `feedsystem_video_go项目设计.md`：原优秀开源项目的业务和架构设计。
- `项目导览.md`：原项目代码结构和调用链分析。
- `实现指导.md`：本项目从 V0.1 到 V2.0 的阶段性实现路线，已重点细化 V0.2。

## 协作角色

助手在本项目中的角色：资深 Go 后端架构师与编程导师。

协作方式：

- 不直接堆大段代码，优先帮助理解架构、分层、错误处理和实现顺序。
- 对已写代码做审查，指出逻辑问题、边界问题和改进方向。
- 代码修改只有在用户明确要求或文档生成类任务中执行。
- 对后端工程质量要求较高，强调 `handler -> service -> repository` 分层、错误语义、HTTP 状态码和可演进设计。

## 当前仓库路径

实际仓库路径是：

```text
/home/reinerbrown/golang/my-feed
```

注意：早期上下文中曾出现旧路径 `/home/reinerbrown/golang/my_feed`，但仓库已移动到 `my-feed`。

## 总体里程碑路线

项目采用以下阶段推进：

```text
V0.1 单体骨架 + 账号系统
Gin + GORM + SQLite/MySQL + JWT + bcrypt

V0.2 视频发布 MVP + 基础 Feed
本地上传 + 视频发布 + 最新 Feed + 视频详情 + 作者作品 + 可选话题

V0.3 互动社交闭环
点赞 / 评论 / 关注 / 用户主页 / 基础私信，全部同步写 MySQL

V1.0 高并发读优化
Redis 缓存 / Feed 时间线 / 热榜 / 限流 / 分片上传

V1.5 事件驱动增强
RabbitMQ + Worker + Outbox + 异步互动 + SSE 通知

V2.0 可选微服务体验
gRPC 拆分账号、视频、互动、Feed、通知等服务
```

## 当前进度

### V0.1 状态：基本告一段落

已经完成并反复审查的内容：

- 账号模型 `Account`。
- 注册接口。
- 登录接口。
- bcrypt 密码哈希和密码校验。
- JWT 生成和解析。
- JWT 鉴权中间件。
- `/account/me` 当前用户信息接口。
- `/account/logout` 清空服务端 token。
- SQLite / MySQL 可通过环境变量切换。
- 唯一索引错误识别。
- HTTP 状态码常用场景文档：`HTTP状态码常用场景.md`。

当前代码结构已拆成较清晰的三层：

```text
internal/account/
  model.go        // Account 模型
  repository.go   // AccountRepo，数据库操作
  service.go      // AccountService，注册/登录/登出/当前用户业务逻辑
  handler.go      // HTTP 参数绑定、响应和状态码映射

internal/auth/
  jwtToken.go     // JWT token 生成和解析

internal/middleware/
  jwtAuth.go      // JWT 鉴权中间件

internal/db/
  db.go           // 数据库初始化、DB_DRIVER 切换、数据库错误识别

internal/router/
  router.go       // 路由装配
```

### V0.2 状态：设计已细化，尚未开始实现

`实现指导.md` 中的 V0.2 部分已经扩展，包含：

- `videos`、`tags`、`video_tags` 表设计和字段解释。
- `internal/video` 和 `internal/feed` 模块建议。
- V0.2 要设计的路由和功能说明。
- `latest_time` 游标分页解释。
- 上传文件、本地路径、URL、Feed 分页、话题提取、事务等难点说明。

## 已达成的关键共识

### 1. 分层边界

当前项目坚持以下依赖方向：

```text
router -> handler -> service -> repository -> db
```

职责边界：

- `router`：只做路由装配和中间件挂载。
- `handler`：只处理 HTTP 层，包括 JSON 绑定、表单文件读取、从 `gin.Context` 取用户 ID、返回状态码和响应体。
- `service`：处理业务逻辑，例如注册校验、密码哈希、登录校验、token 生成、发布视频、话题提取。
- `repository`：只处理数据库读写，不理解 HTTP，也不处理业务响应。
- `model`：只放数据结构和 GORM tag。

强约束：

- `service` 不应该 import `gin` 或 `net/http`。
- `repository` 不应该 import `gin`、`auth`、`bcrypt`。
- `handler` 不应该直接写 GORM 查询，也不应该直接做 bcrypt/JWT 业务细节。

### 2. 错误处理方式

共识：service 返回业务错误，handler 将业务错误映射成 HTTP 状态码。

示例：

```text
ErrUsernameRequired          -> 400 Bad Request
ErrPasswordRequired          -> 400 Bad Request
ErrUsernameExists            -> 409 Conflict
ErrInvalidUsernameOrPassword -> 401 Unauthorized
ErrAccountNotFound           -> 404 Not Found
未知数据库/系统错误            -> 500 Internal Server Error
```

`db.ErrRecordNotFound = gorm.ErrRecordNotFound` 已在 `internal/db` 中定义，用于避免业务层直接依赖 GORM。

### 3. 账号登录态设计

当前 JWT 不是纯无状态方案。登录成功后：

```text
生成 JWT -> 保存到 accounts.token -> 返回给前端
```

鉴权时：

```text
解析 Authorization: Bearer <token>
校验 JWT 签名和过期时间
根据 claims.AccountID 查账号
比较数据库中的 acc.Token 是否等于请求 token
通过后把 userID 和 username 放入 gin.Context
```

登出时：

```text
清空 accounts.token
```

这样可以实现服务端主动撤销 token。

### 4. 数据库切换设计

`internal/db/db.go` 已支持环境变量切换 SQLite / MySQL。

默认：

```bash
go run ./cmd
```

默认使用 SQLite：

```text
DB_DRIVER=sqlite
SQLITE_DSN=./data.db
```

使用 MySQL：

```bash
DB_DRIVER=mysql MYSQL_USER=dev_user MYSQL_PASSWORD=qwerdf MYSQL_HOST=127.0.0.1 MYSQL_PORT=3306 MYSQL_DATABASE=db001 go run ./cmd
```

也可以直接提供完整 DSN：

```bash
DB_DRIVER=mysql DB_DSN='dev_user:qwerdf@tcp(127.0.0.1:3306)/db001?charset=utf8mb4&parseTime=True&loc=Local' go run ./cmd
```

### 5. 本地测试方式

当前环境普通 `go test ./...` 可能因为默认 Go build cache 目录只读而失败。推荐使用：

```bash
env GOCACHE=/tmp/go-build-cache go test ./...
```

最近一次检查结果：所有包通过编译测试，但没有正式单元测试文件。

## V0.1 当前仍需留意的小事项

如果继续打磨账号模块，优先处理：

- 确认 `handler.go`、`repository.go` 已经 `gofmt`。
- `Logout` 和 `Me` 中如果 `userID == 0`，返回错误后必须 `return`，避免继续执行业务逻辑。
- `/me` 中 `ErrAccountNotFound` 返回 404，其他未知错误应该返回 500。
- `repository.UpdateToken` 建议返回 `db.ErrRecordNotFound`，再由 service 转成 `ErrAccountNotFound`。
- 密码强度校验暂时不做，这是用户明确决定的。

## V0.2 设计重点

V0.2 的目标链路：

```text
登录用户上传视频文件
  -> 上传封面文件
  -> 发布视频元数据
  -> 首页最新 Feed 能看到视频
  -> 点进详情能播放
  -> 按作者查看作品
  -> 可选：按 #话题 查询视频
```

### 建议新增模块

```text
internal/video/
  model.go
  repository.go
  service.go
  handler.go

internal/feed/
  service.go
  handler.go
```

V0.2 可以让 `feed` 直接依赖 `video.Repository` 查询视频。后面引入 Redis 后再拆得更细。

### V0.2 表模型

核心表：

```text
videos      视频主表，保存标题、描述、作者、播放 URL、封面 URL、计数字段等

tags        话题表，保存唯一话题名，例如 Go、后端、旅行

video_tags  视频和话题关系表，多对多关系
```

重要共识：

- 视频文件本体不进数据库。
- 数据库存浏览器可访问的 URL，例如 `/static/uploads/videos/...`。
- 本地磁盘路径和访问 URL 必须区分。
- `likes_count`、`comments_count`、`popularity` 在 V0.2 先保留，后续阶段维护。

### V0.2 路由规划

```text
POST /video/uploadVideo
上传 mp4 视频文件，返回 play_url。需要登录。

POST /video/uploadCover
上传封面图片，返回 cover_url。需要登录。

POST /video/publish
发布视频元数据，创建 videos 记录。需要登录。

POST /video/listByAuthorID
查询某作者作品列表。

POST /video/getDetail
查询视频详情。

POST /feed/listLatest
首页最新视频流，按 created_at desc 查询。

POST /feed/listByTag
按话题查询视频。V0.2 后半段可做。
```

### V0.2 开发顺序建议

1. 新建 `internal/video/model.go`，定义 `Video`、`Tag`、`VideoTag`。
2. 在 `cmd/main.go` 中把新模型加入 `AutoMigrate`。
3. 新建 `video.Repository`，先写 `CreateVideo`、`FindVideoByID`、`ListByAuthorID`、`ListLatest`。
4. 新建 `video.Service`，先写 `Publish` 和 `GetDetail`。
5. 新建 `video.Handler`，先实现 `publish/getDetail/listByAuthorID`。
6. 在 router 注册 `/video` 路由，`publish/uploadVideo/uploadCover` 需要 JWT 鉴权。
7. 注册静态文件服务：`router.Static("/static/uploads", ".run/uploads")`。
8. 实现 `uploadVideo` 和 `uploadCover`。
9. 新建 `internal/feed`，实现 `POST /feed/listLatest`。
10. 加入 `latest_time` 游标分页。
11. 后半段实现话题提取、`tags/video_tags` 入库和 `listByTag`。

## latest_time 游标分页共识

不要一开始用 offset 分页。

offset 写法：

```sql
select * from videos order by created_at desc limit 20 offset 20;
```

问题：Feed 数据会不断新增。用户看完第一页后，如果中间新增了视频，再请求 offset 第二页，可能重复或漏数据。

游标分页写法：

```text
第一页：latest_time = 0
服务端返回本页最后一条视频的 created_at 毫秒时间戳作为 next_time
第二页：latest_time = 上次返回的 next_time
查询 created_at < latest_time 的更旧视频
```

推荐在 Go 里把毫秒时间戳转换成 `time.Time`，避免绑定 MySQL 函数：

```go
before := time.UnixMilli(req.LatestTime)
db.Where("created_at < ?", before)
```

后续更稳的方案是复合游标：

```text
created_at + id
```

排序：

```sql
order by created_at desc, id desc
```

下一页条件：

```sql
where created_at < ?
   or (created_at = ? and id < ?)
```

V0.2 先实现 `latest_time` 即可，知道升级方向就行。

## V0.2 难点提醒

- 上传文件名不能直接使用用户原始文件名，必须服务端生成唯一文件名，保留扩展名。
- 要防止路径穿越，例如 `../../../xxx.mp4`。
- 上传接口和发布接口要分开：上传只返回 URL，发布才写 `videos` 表。
- `play_url` 是浏览器访问 URL，不是服务器磁盘路径。
- `router.Static("/static/uploads", ".run/uploads")` 只适合本地开发。
- `latest_time` 单位要统一，建议毫秒。
- 话题提取 V0.2 不要写复杂，先支持 `#Go #后端` 这种简单格式。
- 发布视频并写入 tags/video_tags 时，后半段最好用事务，避免部分写成功导致数据不完整。

## 重要文件索引

```text
实现指导.md
V0.1-V2.0 实现路线，V0.2 已细化。

HTTP状态码常用场景.md
后端接口状态码速查。

internal/account/model.go
账号模型。

internal/account/repository.go
账号数据库访问层。

internal/account/service.go
账号业务逻辑层。

internal/account/handler.go
账号 HTTP 处理层。

internal/auth/jwtToken.go
JWT 生成和解析。

internal/middleware/jwtAuth.go
JWT 鉴权中间件。

internal/db/db.go
数据库初始化、SQLite/MySQL 切换、数据库错误判断。

internal/router/router.go
路由装配。
```

## 后续接续建议

新对话如果继续 V0.2，实现顺序建议从模型开始，不要先写上传接口：

1. 先建 `internal/video/model.go`。
2. 补 `AutoMigrate`。
3. 写 repository/service/handler 的骨架。
4. 先跑通 `POST /video/publish` 和 `POST /video/getDetail`，可以手动传一个假的 `play_url`。
5. 再实现真实文件上传和静态文件服务。
6. 最后实现 `feed/listLatest` 和游标分页。

原因：先跑通数据库主链路，再处理文件上传，调试会更清晰。
