# My Feed System

[Golang找到实习需要什么？](https://www.bilibili.com/video/BV1njAwzcEeJ)

根据以上视频，在AI指导下，自行实现视频feed流系统

## V0.1X 账号注册 / 登录 / JWT / 基础分层

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


## V0.2X 视频上传

### latest_time游标分页

`listByAuthorID` 查询作者作品列表时，如果作者有很多视频，一次不能全返回，需要进行分页

`latest_time` 的意思是：给我创建时间早于 latest_time 的视频

第一页时 `latest_time = 0`

```json
{
  "author_id": 1,
  "limit": 20,
  "latest_time": 0
}
```
服务端返回最新20条视频。取最后一条视频的`created_at`，作为`next_time`

第二页时 `latest_time = next_time`

```json
{
  "author_id": 1,
  "limit": 20,
  "latest_time": next_time
}
```

这样即使中间作者发布了新的视频，第二页也不会被新视频打乱。

实现了从上一页的最后一个视频处，加载20条视频

但这样还有一个不足：多个视频可能有完全相同的 `created_at`，同一时间创建但还没展示完的视频可能被跳过

后续升级为 `created_at` + `id` 复合游标

```sql
order by created_at desc, id desc
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

## V0.3X 点赞、关注、私信功能

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

TODO: 单独抽一个 internal/cursor 包，把所有 list 接口统一升级成 cursor/next_cursor