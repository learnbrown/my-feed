# CLAUDE.MD

## 项目定位

本项目是一个从零手写的 Go 后端练习项目，目标是实现一个“高并发视频 Feed 流系统”，作为求职简历核心项目。

当前策略不是一口气复刻完整开源项目，而是按里程碑逐步演进：

```text
V0.1 单体骨架 + 账号系统
V0.2 视频发布 MVP + 基础 Feed
V0.3 互动社交闭环
V1.0 Redis 高并发读优化
V1.5 MQ + Worker 事件驱动
V2.0 可选微服务拆分
```

参考文档：

- `feedsystem_video_go项目设计.md`：原参考项目的业务和架构设计。
- `项目导览.md`：原参考项目代码结构和调用链分析。
- `实现指导.md`：本项目阶段性实现路线，当前已推进到 V1.0 前。
- `doc/*.md`：当前项目 API 文档。

## 协作角色

助手在本项目中的角色是：资深 Go 后端架构师与编程导师。

协作方式：

- 重点帮助用户理解架构、分层、错误处理、事务、一致性和高并发演进。
- 用户要求“检查”“给出修改意见”时，默认不直接修改代码。
- 用户明确要求生成或更新文档时，可以直接编辑文档。
- 代码审查时优先指出逻辑错误、边界问题、数据一致性风险和接口契约不清晰之处。
- 不糊大段代码，先讲清楚为什么这样设计，再给可落地的实现路径。

## 当前仓库

实际仓库路径：

```text
/home/reinerbrown/golang/my-feed
```

注意：

- 早期上下文中出现过旧路径 `/home/reinerbrown/golang/my_feed`，现在应忽略。
- Go module/import 路径里仍可能是 `my_feed`，这是代码层面的 module 名，不等同于磁盘目录名。

## 当前进度

### V0.1：账号与认证，已完成 MVP

已实现内容：

- `Account` 模型。
- 注册、登录、登出、当前用户 `/account/me`。
- bcrypt 密码哈希与校验。
- JWT 生成、解析、鉴权中间件。
- 服务端 token 存储与登出撤销。
- SQLite/MySQL 通过环境变量切换。
- 唯一索引错误识别。
- `handler -> service -> repository` 基础分层。

当前 JWT 方案不是纯无状态：

```text
登录成功 -> 生成 JWT -> 写入 accounts.token -> 返回前端
鉴权 -> 解析 JWT -> 查账号 -> 比较数据库 token -> 通过
登出 -> 清空 accounts.token
```

这个设计可以支持服务端主动撤销 token。后续 V1.0 可以把 token 校验缓存到 Redis，但 MySQL 仍是最终可信来源。

### V0.2：视频发布 MVP + 基础 Feed，已完成 MVP

已实现内容：

- 视频模型 `Video`。
- 标签模型 `Tag`、视频标签关联 `VideoTag`。
- 视频上传接口。
- 封面上传接口。
- 视频发布接口。
- 视频详情接口。
- 作者作品列表。
- 首页最新 Feed。
- 按标签查询 Feed。
- 发布视频时提取标签、创建标签、创建视频标签关联。
- 视频发布和标签关联写入使用事务，避免部分成功造成脏数据。

重要约定：

- 视频文件本体不进数据库，数据库只保存浏览器可访问的 URL。
- 本地磁盘路径和访问 URL 必须区分。
- 上传接口只负责保存文件并返回 URL，发布接口才写 `videos` 记录。
- `cover_url` 可以使用默认封面。
- `likes_count`、`comments_count` 是视频表上的冗余计数字段，V0.3 开始维护。
- `popularity` 是热度分字段，V1.0 做热榜时再使用。

### V0.3：互动社交闭环，已完成 MVP

已实现内容：

- 点赞：`/like/like`、`/like/unlike`、`/like/isLiked`、`/like/listLikedVideos`。
- 评论：`/comment/publish`、`/comment/delete`、`/comment/listComment`。
- 关注：`/follow/follow`、`/follow/unfollow`、`/follow/isFollowing`、`/follow/listFollower`、`/follow/listFollowing`。
- 私信：`/message/sendMsg`、`/message/listConversation`。
- 用户主页：`/account/getProfile`。
- DTO 响应结构已开始使用，避免直接把 GORM 模型完整返回给前端。
- 需要分页的列表接口已升级为双游标：`latest_time + latest_id`，响应返回 `next_time + next_id`。

V0.3 的核心训练点已经覆盖：

- 幂等：点赞、取消点赞、关注、取关。
- 事务：点赞记录与 `likes_count`，评论记录与 `comments_count`。
- 权限：只能删除自己的评论，私信不能发给自己。
- 聚合：`getProfile` 统计视频数、获赞数、粉丝数、关注数。
- 分页：避免简单 offset，使用稳定游标分页。

## 当前模块结构

```text
internal/account/     账号、登录、登出、当前用户
internal/auth/        JWT token 生成和解析
internal/comment/     评论业务
internal/db/          DB 初始化、SQLite/MySQL 切换、数据库错误识别
internal/feed/        Feed 查询
internal/follow/      关注关系
internal/like/        点赞业务
internal/message/     私信业务
internal/middleware/  JWT 鉴权中间件
internal/profile/     用户主页聚合
internal/router/      路由装配
internal/video/       视频、上传、标签、视频查询
```

稳定依赖方向：

```text
router -> handler -> service -> repository -> db
```

职责边界：

- `router`：只做路由装配、中间件挂载、依赖组装。
- `handler`：只处理 HTTP 层，包括参数绑定、从 `gin.Context` 取登录用户、状态码映射和响应。
- `service`：处理业务规则、事务编排、DTO 组装、哨兵错误返回。
- `repository`：处理数据库查询和写入，不理解 HTTP。
- `model`：定义 GORM 数据模型。

强约束：

- `service` 不应该依赖 `gin`。
- `repository` 不应该依赖 `gin`、`auth`、`bcrypt`。
- `handler` 不应该直接写 GORM 查询，也不应该直接处理 bcrypt/JWT 细节。
- DTO 应该靠近 service/response shaping，不要塞进 repository 作为数据库模型。

## 当前接口概览

账号：

- `POST /account/register`
- `POST /account/login`
- `POST /account/logout`
- `GET /account/me`
- `POST /account/getProfile`

视频与 Feed：

- `POST /video/uploadVideo`
- `POST /video/uploadCover`
- `POST /video/publish`
- `POST /video/getDetail`
- `POST /video/listByAuthorID`
- `POST /feed/listLatest`
- `POST /feed/listByTag`

互动：

- `POST /like/like`
- `POST /like/unlike`
- `POST /like/isLiked`
- `POST /like/listLikedVideos`
- `POST /comment/publish`
- `POST /comment/delete`
- `POST /comment/listComment`
- `POST /follow/follow`
- `POST /follow/unfollow`
- `POST /follow/isFollowing`
- `POST /follow/listFollower`
- `POST /follow/listFollowing`
- `POST /message/sendMsg`
- `POST /message/listConversation`

具体请求和响应格式以 `doc/*.md` 为准。

## 游标分页共识

当前项目已经从单字段游标升级为复合游标：

```text
latest_time + latest_id
next_time + next_id
```

原因：

- 只用 `created_at` 时，如果多条记录时间相同，翻页可能漏数据或重复数据。
- 复合游标用 `created_at + id` 可以形成稳定顺序。

典型排序：

```sql
ORDER BY created_at DESC, id DESC
```

典型下一页条件：

```sql
WHERE created_at < ?
   OR (created_at = ? AND id < ?)
```

对某些列表，游标字段不是视频发布时间，而是关系记录时间：

- `listByAuthorID`：通常按 `videos.created_at + videos.id`。
- `listLatest`：按 `videos.created_at + videos.id`。
- `listByTag`：按 `videos.created_at + videos.id`。
- `listLikedVideos`：按 `likes.created_at + likes.id`。
- `listComment`：按 `comments.created_at + comments.id`。
- `listFollower`：按 `follows.created_at + follows.id`。
- `listFollowing`：按 `follows.created_at + follows.id`。
- `listConversation`：按 `messages.created_at + messages.id`。

当前已记录的技术债：

- 请求和响应使用毫秒时间戳 `UnixMilli()`，如果数据库保存更高精度时间，存在精度损失风险。用户已决定后续找一个版本统一升级。
- 后续可选方案一：把数据库时间字段精度统一为毫秒。
- 后续可选方案二：改成不透明字符串游标 `next_cursor`，把完整时间精度和 ID 编码进去。

## 当前注意事项

这些不是立刻必须完成，但后续继续写 V1.0 前要心里有数：

- 检查所有列表接口是否都拒绝非法游标组合，例如只传 `latest_time` 不传 `latest_id`，或只传 `latest_id` 不传 `latest_time`。
- 对负数 `latest_time` 的处理要统一。不要让负时间戳被当成第一页。
- API 文档要持续跟进双游标字段，避免前后端契约漂移。
- DTO 不要半途而废，尤其是视频、评论、私信、关注列表，避免把模型内部字段暴露出去。
- `#C#` 这类标签提取规则如果后续继续打磨，需要专门设计正则和测试样例。
- 当前仍是同步 MySQL 版本。不要急着上 MQ；同步版本不稳，异步化只会把错误藏得更深。

建议重点检查和补齐的索引：

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

私信后续如果要继续演进，可以考虑引入 `conversation_id`，否则双方消息的 OR 查询会越来越别扭。

## V1.0：Redis 高并发读优化

下一阶段目标不是“把 Redis 塞进所有地方”，而是用 Redis 解决 V0.3 后已经出现的高频读问题：

- Feed 最新流读取频繁。
- 视频详情读取频繁。
- 用户主页聚合统计读取频繁。
- 点赞状态、关注状态读取频繁。
- 热榜需要按热度快速排序。
- 登录、上传、发布、评论、私信等接口需要基础限流。

### V1.0 技术原则

必须坚持：

- MySQL 仍是最终数据源。
- Redis 首先作为缓存和加速层，不要一开始就当主存储。
- Redis 不可用时，核心业务应能降级回 MySQL，只是变慢，不能直接全站挂掉。
- 每个 Redis key 都要能讲清楚解决什么问题。
- 缓存必须考虑失效策略，不能只写读缓存不写删除/更新逻辑。

暂时不要做：

- 不要一上来引入 MQ。
- 不要一上来做异步点赞。
- 不要一上来做复杂推荐算法。
- 不要把所有接口都缓存一遍。
- 不要为了简历堆中间件，必须能解释一致性边界。

### V1.0 推荐推进顺序

#### 1. Redis 基础设施

新增建议：

```text
internal/cache/
  redis.go
```

要做的事：

- 使用 `go-redis/v9`。
- 通过环境变量配置 Redis 地址、密码、DB。
- 启动时 `Ping`。
- Redis 连接失败时允许服务降级启动，或者至少给出清晰错误；具体策略要在文档里写明。
- 封装统一 key 前缀，避免裸字符串散落在业务代码里。

建议 key 前缀：

```text
myfeed:account:token:{account_id}
myfeed:video:detail:{video_id}
myfeed:profile:{account_id}
myfeed:feed:latest
myfeed:rank:hot
myfeed:rate:{scene}:{identifier}
```

#### 2. JWT token 缓存

改造目标：

```text
鉴权 -> 先查 Redis token
Redis miss -> 查 MySQL accounts.token
MySQL 命中 -> 回填 Redis
登出 -> 删除 Redis token + 清空 MySQL token
```

注意：

- token 缓存 TTL 不应超过 JWT 自身过期时间。
- Redis token 只是加速，不能让它破坏登出语义。
- 登出时 Redis 和 MySQL 都要处理。

#### 3. 视频详情缓存

适合先做，因为它边界清楚。

缓存模式：

```text
Cache Aside
读 Redis -> miss -> 读 MySQL -> 写 Redis -> 返回
```

写操作失效：

- 发布视频后可以不立刻缓存详情。
- 修改视频、删除视频、改变状态时要删除详情缓存。
- 点赞数、评论数变化后，要么删除详情缓存，要么更新缓存里的计数字段。学习阶段建议先删除缓存，逻辑更稳。

#### 4. 用户主页聚合缓存

`getProfile` 会聚合多张表：

```text
account 基本信息
videos_count
likes_count
followers_count
followings_count
```

V1.0 可以把主页结果缓存短 TTL，例如 30 秒到 120 秒。

注意：

- 关注/取关会影响粉丝数和关注数。
- 点赞/取消点赞会影响作者获赞数。
- 发布/删除视频会影响作品数。
- 学习阶段可以先用短 TTL 降低失效复杂度，再逐步加主动删除。

#### 5. Feed 最新流缓存

建议使用 Redis ZSET：

```text
key:   myfeed:feed:latest
score: created_at 毫秒时间戳
value: video_id
```

读流程：

```text
先从 ZSET 取一页 video_id
再批量取视频详情缓存
详情 miss 再回源 MySQL
```

注意：

- 仅用毫秒 score 仍有同分问题，后续可以把 score 和 member 设计得更严谨。
- 如果继续使用双游标，Redis 查询也要保持和 MySQL 一致的排序语义。
- 冷数据可以回源 MySQL，不要强求 Redis 保存全部历史。

#### 6. 热榜

可以使用 Redis ZSET：

```text
key:   myfeed:rank:hot
score: popularity
value: video_id
```

热度可以先用简单公式：

```text
popularity = likes_count * 3 + comments_count * 5
```

V1.0 先同步更新 Redis 热榜即可。V1.5 再考虑 MQ + Worker 异步更新。

#### 7. 基础限流

可以先做固定窗口：

```text
INCR myfeed:rate:{scene}:{identifier}
EXPIRE key window_seconds
```

适合场景：

- 登录失败限流。
- 上传接口限流。
- 发布视频限流。
- 评论/私信限流。
- 点赞/取消点赞限流。

注意：

- 限流失败返回 `429 Too Many Requests`。
- key 要区分 IP、用户 ID、业务场景。
- 登录前没有用户 ID，只能用 IP 或用户名维度。

## V1.0 面试亮点

真正值得打磨、也最容易被问深的点：

- 为什么不用 offset，而用复合游标。
- Redis 缓存如何设计 key、TTL、失效策略。
- Cache Aside 的一致性边界。
- Redis 挂了系统怎么降级。
- 点赞/评论计数字段为什么是冗余字段，如何保证与明细表一致。
- Feed 最新流为什么适合 ZSET。
- 热榜如何计算、如何更新、如何避免每次全表排序。
- 限流为什么要按场景设计 key。
- 为什么 V1.0 先做 Redis，同步链路稳定后 V1.5 再做 MQ。

## 常用验证命令

当前环境中普通 `go test ./...` 可能因为默认 Go build cache 目录权限失败，建议使用：

```bash
env GOCACHE=/tmp/go-build-cache go test ./...
```

格式化：

```bash
gofmt -w internal
```

运行：

```bash
go run ./cmd
```

切换 MySQL 示例：

```bash
DB_DRIVER=mysql \
MYSQL_USER=dev_user \
MYSQL_PASSWORD=qwerdf \
MYSQL_HOST=127.0.0.1 \
MYSQL_PORT=3306 \
MYSQL_DATABASE=db001 \
go run ./cmd
```

## 新对话接手提示

如果后续新对话继续本项目，应先确认：

1. 当前工作目录是否为 `/home/reinerbrown/golang/my-feed`。
2. 用户是否要求“只给思路，不改代码”。
3. 当前任务属于代码审查、文档更新，还是直接实现。
4. API 文档 `doc/*.md` 是否已经和代码同步。
5. 列表分页是否仍保持 `latest_time + latest_id` 双游标契约。
6. V1.0 Redis 改造是否从基础设施和读缓存开始，而不是直接异步化。

当前下一步建议：

```text
先封装 Redis client 和配置
再改造 JWT token 缓存
再做 video detail / profile 的 Cache Aside
然后做 Feed ZSET
最后做热榜和限流
```

严厉一点说：V1.0 开始后，每加一个 Redis key，都必须能回答三个问题：

```text
它缓存什么？
什么时候失效？
Redis 挂了怎么办？
```
