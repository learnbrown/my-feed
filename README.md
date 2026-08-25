# My Feed System

[Golang找到实习需要什么？](https://www.bilibili.com/video/BV1njAwzcEeJ)

根据以上视频，在AI指导下，自行实现视频feed流系统

当前里程碑：

```text
V0.1 单体骨架 + 账号认证                    已完成 MVP
V0.2 视频发布 + 基础 Feed                   已完成 MVP
V0.3 点赞、评论、关注、主页、私信           已完成 MVP
V1.0 Redis 高并发读优化                     进行中
V1.5 MQ + Worker 事件驱动                   未开始
V2.0 可选微服务拆分                         未开始
```

## V0.1 账号注册 / 登录 / JWT / 基础分层

将数据库、路由、中间件、JWT令牌方法模型定义处理放入不同模块中实现

### 路由Router
仅进行路由注册，分组的操作，不进行任何业务逻辑

### JWT生成与解析方法
将用户名与id加密生成token
解密token获得用户名与id

### Login路由
登录成功后生成token，更新数据库字段

### JWT中间件
验证请求头中的token
与数据库token字段比对
解析获取用户名与id，放入gin上下文
字段名`userID`与`username`

### 需鉴权路由
应用JWT中间件，在访问时验证token有效性

### handler -> service -> repository 三层分明

#### Model 

定义表结构模型

#### Handler 层：只做输入输出

接收请求参数，调用service，处理错误，返回信息

不进行业务逻辑和数据库操作

#### Service 层：只做业务逻辑

接收handler的参数，实现注册、登录、登出业务逻辑，

#### Repository 层：只做数据库读写 (CRUD)

实现对数据库的增删改查


## V0.2 视频上传

### `latest_time + latest_id` 复合游标分页

所有列表接口均使用 `created_at + id` 形成稳定顺序，避免多条记录创建时间相同时漏数据或重复数据。

第一页：

```json
{
  "author_id": 1,
  "limit": 20,
  "latest_time": 0,
  "latest_id": 0
}
```

下一页使用上次响应的 `next_time + next_id`：

```json
{
  "author_id": 1,
  "limit": 20,
  "latest_time": next_time,
  "latest_id": next_id
}
```

```sql
ORDER BY created_at DESC, id DESC

WHERE created_at < ?
   OR (created_at = ? AND id < ?)
```

### publish 提取tag并创建关联

在`repository`层使用gorm事务，`service`层中将创建视频、创建标签、创建video-tag关联的业务传入，某一步出错时会整体回滚，不会出现只创建视频而没创建tag或video-tag的情况

```go
func (repo *VideoRepo) Transaction(fn func(txRepo *VideoRepo) error) error {
	return repo.db.Transaction(func(tx *gorm.DB) error {
		txRepo := NewVideoRepo(tx)
		return fn(txRepo)
	})
}
```

## V0.3 点赞、关注、私信功能

### like

冗余字段
幂等
唯一索引
gorm事务 创建like与更新计数
取消点赞时不能为负

#### 重复点赞只增加一条记录

唯一索引负责兜底防重复数据
service 负责把重复请求解释成幂等成功
事务负责保证 like 记录和 likes_count 一致

#### like模型不能包含deleted_at字段

如果有DeletedAt字段，gorm会采用软删除，导致点赞记录被删但唯一索引依然存在，无法再次点赞

like model只包含id和created_at字段

#### 升级复合游标

是不是所有的分页都要升级成复合游标 -> 是

与List相关接口
- [x] listByAuthorID
- [x] listLatest
- [x] listByTag
- [x] listLikedVideos
- [x] listComment
- [x] listFollower
- [x] listFollowing
- [x] listConversation

#### DTO

DTO 是 Data Transfer Object，数据传输对象。
在你的项目里可以简单理解成：
数据库 Model：服务于数据库读写
DTO：服务于接口入参/出参

已将所有直接向前端返回model的，改为返回DTO，去除不需要字段，并将时间转化为毫秒时间戳

#### 时间精度损失

接收前端请求和返回响应时，都是使用毫秒时间戳，并将其转为UnixMilli

而数据库保存的`created_at`精度含有微秒/纳秒，会导致有些数据被漏掉

后续可选方案是抽取 `internal/cursor`，升级为不透明的 `cursor/next_cursor`，把完整时间精度和 ID 一起编码。

## V1.0 Redis 高并发读优化（进行中）

已完成：

- Redis 配置、client、启动 `Ping`、统一 key 前缀，以及启动时不可用则禁用缓存。
- JWT token 缓存：Redis hit 校验，miss 或读取异常回源 MySQL，成功后回填。
- 视频详情 Cache Aside，TTL 为 5 分钟。
- 点赞、取消点赞、发布评论、删除评论后删除视频详情缓存。
- 用户主页 60 秒 Cache Aside；关注、取关、点赞、取消点赞和发布视频后主动删除受影响的主页缓存。
- service/handler 单元测试、miniredis 测试和需要显式启用的 MySQL 集成测试。

本轮 token 缓存收口已完成：

- token cache TTL 改为 `min(JWT 剩余有效期, 5 分钟)`，缓存命中不续期。
- 登录先更新 MySQL 再尝试 Redis `SET`；登出先清空 MySQL 再尝试 Redis `DEL`。
- Redis `GET/SET/DEL` 失败不阻断业务；缓存 mismatch 回查 MySQL。
- 删除高频缓存 hit/miss/success 日志，缓存异常日志统一带 component、operation 和业务 ID。

当前待补：

- MySQL 集成测试默认跳过，尚未建立持续执行环境。
- 尚未建立 Bruno 主链路、批量数据生成和 Redis 前后性能基线。

下一步：

```text
Feed 最新流 ZSET
  -> 热榜 ZSET
  -> 基础限流
```

用户主页缓存采用适合当前个人项目的轻量方案：缓存现有 Profile DTO，固定 TTL 60 秒，Redis 异常回源 MySQL；关注、点赞和发布成功后主动删除受影响的主页缓存。实现和验收结果见 `doc/10 用户主页缓存实现方案.md`。

MySQL 仍是 token 最终数据源。当前实现采用有界最终一致性：正常登录/登出主动更新或删除 Redis，Redis 失败时业务仍由 MySQL 完成，旧缓存异常窗口最多 5 分钟。当前只在本地开发，继续使用现有 token key，具体见 `doc/09 token缓存折中改造方案.md`。

## 本地开发

前后端分别启动。Gin 只提供 API 和 `/static/uploads` 上传文件，Vue 页面由 Vite 开发服务器提供。

首次运行时，根据示例创建各自的本地环境变量文件：

```bash
cp .env.example .env
cp frontend/.env.example frontend/.env.local
```

`.env` 保存后端数据库、Redis 和 JWT 配置；`frontend/.env.local` 保存前端本地配置。这两个文件都不会提交到 Git。前端本地开发时保持 `VITE_API_BASE_URL` 为空，请求会经过 Vite 代理访问 Gin；只有前后端部署在不同域名时才需要填写公开的后端地址。`VITE_*` 变量会进入浏览器构建产物，不能存放密码或密钥。

启动后端：

```bash
go run ./cmd
```

另开一个终端启动前端：

```bash
npm --prefix frontend install
npm --prefix frontend run dev
```

浏览器访问 `http://localhost:5173`。前端开发服务器会把账号、视频、Feed、互动接口和 `/static/uploads` 请求代理到 `http://localhost:8080`。

前端生产构建：

```bash
npm --prefix frontend run build
```

构建产物位于 `frontend/dist`，当前开发服务器不会由 Gin 托管这些文件。

## 验证

默认测试：

```bash
env GOCACHE=/tmp/go-build-cache go test ./...
env GOCACHE=/tmp/go-build-cache go vet ./...
```

MySQL 集成测试必须使用以 `_test` 结尾的独立数据库，并显式启用：

```bash
RUN_MYSQL_TESTS=1 \
MYSQL_DBNAME=myfeed_test \
env GOCACHE=/tmp/go-build-cache go test ./internal/integration -v
```
