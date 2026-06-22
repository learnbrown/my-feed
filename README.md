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

## API 文档

### 通用约定

- Base URL: `http://localhost:8080`
- JSON 接口请求头：`Content-Type: application/json`
- 需要登录的接口请求头：`Authorization: Bearer <token>`
- Feed 分页使用 `latest_time` 游标，第一页传 `0` 或不传；下一页传上一次响应中的 `next_time`。
- `limit` 可选，默认 `20`，最大 `50`。

### 通用错误响应

```json
{
  "error": "error message"
}
```

常见状态码：

| 状态码 | 含义 |
| --- | --- |
| `200 OK` | 请求成功 |
| `201 Created` | 创建成功 |
| `400 Bad Request` | 请求参数错误 |
| `401 Unauthorized` | 未登录、Token 缺失或 Token 无效 |
| `404 Not Found` | 资源不存在 |
| `409 Conflict` | 资源冲突，例如用户名重复 |
| `500 Internal Server Error` | 服务端内部错误 |

### 数据结构

#### Account

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `id` | number | 用户 ID |
| `username` | string | 用户名 |
| `avatar_url` | string | 用户头像地址 |
| `bio` | string | 用户简介 |

#### Video

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| `ID` | number | 视频 ID，来自 GORM 默认模型 |
| `CreatedAt` | string | 创建时间 |
| `UpdatedAt` | string | 更新时间 |
| `DeletedAt` | object/null | 软删除字段 |
| `author_id` | number | 作者用户 ID |
| `title` | string | 视频标题 |
| `description` | string | 视频描述 |
| `play_url` | string | 视频播放地址 |
| `cover_url` | string | 视频封面地址 |
| `likes_count` | number | 点赞数 |
| `comments_count` | number | 评论数 |
| `popularity` | number | 热度分 |
| `status` | number | 视频状态，`1` 表示正常 |

---

## 账户接口

### 用户注册

注册新用户。用户名唯一，密码会加密存储。

| 项目 | 内容 |
| --- | --- |
| URL | `/account/register` |
| Method | `POST` |
| Auth | 不需要 |
| Content-Type | `application/json` |

#### 请求参数

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `username` | string | 是 | 用户名 |
| `password` | string | 是 | 密码 |

#### 请求示例

```json
{
  "username": "user",
  "password": "user"
}
```

#### 成功响应

Status: `201 Created`

```json
{
  "status": "ok",
  "id": 1,
  "username": "user"
}
```

#### 可能错误

| 状态码 | 场景 |
| --- | --- |
| `400` | 请求体格式错误、用户名为空、密码为空 |
| `409` | 用户名已存在 |
| `500` | 服务端内部错误 |

### 用户登录

校验用户名和密码，登录成功后返回 JWT Token。

| 项目 | 内容 |
| --- | --- |
| URL | `/account/login` |
| Method | `POST` |
| Auth | 不需要 |
| Content-Type | `application/json` |

#### 请求参数

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `username` | string | 是 | 用户名 |
| `password` | string | 是 | 密码 |

#### 请求示例

```json
{
  "username": "user",
  "password": "user"
}
```

#### 成功响应

Status: `200 OK`

```json
{
  "account": {
    "avatar_url": "",
    "bio": "",
    "id": 1,
    "username": "user"
  },
  "token": "jwt-token"
}
```

#### 可能错误

| 状态码 | 场景 |
| --- | --- |
| `400` | 请求体格式错误 |
| `401` | 用户名或密码错误 |
| `500` | 服务端内部错误 |

### 获取当前用户信息

根据 JWT Token 获取当前登录用户信息。

| 项目 | 内容 |
| --- | --- |
| URL | `/account/me` |
| Method | `GET` |
| Auth | 需要 |
| Content-Type | 无 |

#### Header

| 字段 | 必填 | 示例 |
| --- | --- | --- |
| `Authorization` | 是 | `Bearer <token>` |

#### 成功响应

Status: `200 OK`

```json
{
  "avatar_url": "",
  "bio": "",
  "id": 1,
  "username": "user"
}
```

#### 可能错误

| 状态码 | 场景 |
| --- | --- |
| `401` | Token 缺失、格式错误、过期或已登出 |
| `404` | 用户不存在 |
| `500` | 服务端内部错误 |

### 用户登出

清空数据库中的 Token，使当前 Token 失效。

| 项目 | 内容 |
| --- | --- |
| URL | `/account/logout` |
| Method | `POST` |
| Auth | 需要 |
| Content-Type | 无 |

#### Header

| 字段 | 必填 | 示例 |
| --- | --- | --- |
| `Authorization` | 是 | `Bearer <token>` |

#### 成功响应

Status: `200 OK`

```json
{
  "status": "ok"
}
```

#### 可能错误

| 状态码 | 场景 |
| --- | --- |
| `401` | Token 缺失、格式错误、过期或已登出 |
| `500` | 服务端内部错误 |

---

## 视频接口

### 上传视频文件

上传视频文件，返回可访问的 `play_url`。该接口只保存文件，不创建视频记录。

| 项目 | 内容 |
| --- | --- |
| URL | `/video/uploadVideo` |
| Method | `POST` |
| Auth | 需要 |
| Content-Type | `multipart/form-data` |

#### Header

| 字段 | 必填 | 示例 |
| --- | --- | --- |
| `Authorization` | 是 | `Bearer <token>` |

#### 表单参数

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `video` | file | 是 | 视频文件，仅支持 `.mp4`，最大 `200MB` |

#### 成功响应

Status: `200 OK`

```json
{
  "play_url": "/static/uploads/videos/1/20260621/1782029118133453392_83162.mp4"
}
```

#### 可能错误

| 状态码 | 场景 |
| --- | --- |
| `400` | 缺少文件、文件类型不支持、文件过大 |
| `401` | 未登录或 Token 无效 |
| `500` | 文件保存失败或服务端内部错误 |

### 上传封面文件

上传视频封面，返回可访问的 `cover_url`。该接口只保存文件，不创建视频记录。

| 项目 | 内容 |
| --- | --- |
| URL | `/video/uploadCover` |
| Method | `POST` |
| Auth | 需要 |
| Content-Type | `multipart/form-data` |

#### Header

| 字段 | 必填 | 示例 |
| --- | --- | --- |
| `Authorization` | 是 | `Bearer <token>` |

#### 表单参数

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `cover` | file | 是 | 封面图片，支持 `.jpg/.jpeg/.png/.webp`，最大 `5MB` |

#### 成功响应

Status: `200 OK`

```json
{
  "cover_url": "/static/uploads/covers/1/20260621/1782029119631490429_89459.png"
}
```

#### 可能错误

| 状态码 | 场景 |
| --- | --- |
| `400` | 缺少文件、文件类型不支持、文件过大 |
| `401` | 未登录或 Token 无效 |
| `500` | 文件保存失败或服务端内部错误 |

### 发布视频

创建视频记录，并从 `title + description` 中提取标签，写入 `tags` 和 `video_tags`。视频、标签、关联关系在同一个数据库事务中创建。

| 项目 | 内容 |
| --- | --- |
| URL | `/video/publish` |
| Method | `POST` |
| Auth | 需要 |
| Content-Type | `application/json` |

#### Header

| 字段 | 必填 | 示例 |
| --- | --- | --- |
| `Authorization` | 是 | `Bearer <token>` |

#### 请求参数

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `title` | string | 是 | 视频标题，可包含 `#tag` |
| `description` | string | 否 | 视频描述，可包含 `#tag` |
| `play_url` | string | 是 | 视频地址，通常来自 `/video/uploadVideo` |
| `cover_url` | string | 否 | 封面地址，缺省时使用默认封面 |

#### 请求示例

```json
{
  "title": "First video #go #gin",
  "description": "feed system #gorm",
  "play_url": "/static/uploads/videos/1/20260621/1782029781170370643_58358.mp4",
  "cover_url": "/static/uploads/covers/1/20260621/1782029119631490429_89459.png"
}
```

#### 成功响应

Status: `201 Created`

```json
{
  "video": {
    "ID": 26,
    "CreatedAt": "2026-06-21T16:16:22.707788517+08:00",
    "UpdatedAt": "2026-06-21T16:16:22.707788517+08:00",
    "DeletedAt": null,
    "author_id": 1,
    "title": "First video #go #gin",
    "description": "feed system #gorm",
    "play_url": "/static/uploads/videos/1/20260621/1782029781170370643_58358.mp4",
    "cover_url": "/static/uploads/covers/1/20260621/1782029119631490429_89459.png",
    "likes_count": 0,
    "comments_count": 0,
    "popularity": 0,
    "status": 1
  }
}
```

#### 可能错误

| 状态码 | 场景 |
| --- | --- |
| `400` | 请求体格式错误、标题为空、播放地址为空 |
| `401` | 未登录或 Token 无效 |
| `500` | 视频、标签或关联关系创建失败 |

### 查询视频详情

根据视频 ID 查询单个正常状态的视频。

| 项目 | 内容 |
| --- | --- |
| URL | `/video/getDetail` |
| Method | `POST` |
| Auth | 不需要 |
| Content-Type | `application/json` |

#### 请求参数

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `id` | number | 是 | 视频 ID |

#### 请求示例

```json
{
  "id": 26
}
```

#### 成功响应

Status: `200 OK`

```json
{
  "video": {
    "ID": 26,
    "CreatedAt": "2026-06-21T16:16:22.707788517+08:00",
    "UpdatedAt": "2026-06-21T16:16:22.707788517+08:00",
    "DeletedAt": null,
    "author_id": 1,
    "title": "First video #go #gin",
    "description": "feed system #gorm",
    "play_url": "/static/uploads/videos/1/20260621/1782029781170370643_58358.mp4",
    "cover_url": "/static/uploads/covers/default.png",
    "likes_count": 0,
    "comments_count": 0,
    "popularity": 0,
    "status": 1
  }
}
```

#### 可能错误

| 状态码 | 场景 |
| --- | --- |
| `400` | 请求体格式错误或缺少 `id` |
| `404` | 视频不存在 |
| `500` | 服务端内部错误 |

### 查询作者作品列表

查询某个作者发布的视频列表，按创建时间倒序返回。

| 项目 | 内容 |
| --- | --- |
| URL | `/video/listByAuthorID` |
| Method | `POST` |
| Auth | 不需要 |
| Content-Type | `application/json` |

#### 请求参数

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `author_id` | number | 是 | 作者用户 ID |
| `limit` | number | 否 | 返回数量，默认 `20`，最大 `50` |
| `latest_time` | number | 否 | 分页游标，毫秒时间戳，第一页传 `0` |

#### 请求示例

```json
{
  "author_id": 1,
  "limit": 10,
  "latest_time": 0
}
```

#### 成功响应

Status: `200 OK`

```json
{
  "has_more": false,
  "next_time": 1781950727842,
  "videos": [
    {
      "ID": 26,
      "CreatedAt": "2026-06-21T16:16:22.707788517+08:00",
      "UpdatedAt": "2026-06-21T16:16:22.707788517+08:00",
      "DeletedAt": null,
      "author_id": 1,
      "title": "First video #go #gin",
      "description": "feed system #gorm",
      "play_url": "/static/uploads/videos/1/20260621/1782029781170370643_58358.mp4",
      "cover_url": "/static/uploads/covers/default.png",
      "likes_count": 0,
      "comments_count": 0,
      "popularity": 0,
      "status": 1
    }
  ]
}
```

#### 可能错误

| 状态码 | 场景 |
| --- | --- |
| `400` | 请求体格式错误或缺少 `author_id` |
| `500` | 服务端内部错误 |

---

## Feed 接口

### 查询最新视频流

查询全站最新视频，按创建时间倒序返回。

| 项目 | 内容 |
| --- | --- |
| URL | `/feed/listLatest` |
| Method | `POST` |
| Auth | 不需要 |
| Content-Type | `application/json` |

#### 请求参数

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `limit` | number | 否 | 返回数量，默认 `20`，最大 `50` |
| `latest_time` | number | 否 | 分页游标，毫秒时间戳，第一页传 `0` |

#### 请求示例

```json
{
  "limit": 20,
  "latest_time": 0
}
```

#### 成功响应

Status: `200 OK`

```json
{
  "has_more": false,
  "next_time": 1781950727842,
  "videos": [
    {
      "ID": 33,
      "CreatedAt": "2026-06-22T11:09:38.194827956+08:00",
      "UpdatedAt": "2026-06-22T11:09:38.194827956+08:00",
      "DeletedAt": null,
      "author_id": 1,
      "title": "First video #go #gin",
      "description": "feed system #gorm",
      "play_url": "/static/uploads/videos/1/20260622/1782097776037592606_92258.mp4",
      "cover_url": "/static/uploads/covers/default.png",
      "likes_count": 0,
      "comments_count": 0,
      "popularity": 0,
      "status": 1
    }
  ]
}
```

#### 可能错误

| 状态码 | 场景 |
| --- | --- |
| `400` | 请求体格式错误 |
| `500` | 服务端内部错误 |

### 按标签查询视频流

根据标签名查询视频列表，按创建时间倒序返回。服务端会兼容 `go` 和 `#go` 两种传法，并统一转为小写查询。

| 项目 | 内容 |
| --- | --- |
| URL | `/feed/listByTag` |
| Method | `POST` |
| Auth | 不需要 |
| Content-Type | `application/json` |

#### 请求参数

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `tag_name` | string | 是 | 标签名，例如 `go` 或 `#go` |
| `limit` | number | 否 | 返回数量，默认 `20`，最大 `50` |
| `latest_time` | number | 否 | 分页游标，毫秒时间戳，第一页传 `0` |

#### 请求示例

```json
{
  "tag_name": "go",
  "limit": 10,
  "latest_time": 0
}
```

#### 成功响应

Status: `200 OK`

```json
{
  "has_more": false,
  "next_time": 1782116861497,
  "videos": [
    {
      "ID": 34,
      "CreatedAt": "2026-06-22T16:27:41.497388754+08:00",
      "UpdatedAt": "2026-06-22T16:27:41.497388754+08:00",
      "DeletedAt": null,
      "author_id": 1,
      "title": "First video #go #gin",
      "description": "feed system #gorm",
      "play_url": "/static/uploads/videos/1/20260622/1782116858305691268_62898.mp4",
      "cover_url": "/static/uploads/covers/default.png",
      "likes_count": 0,
      "comments_count": 0,
      "popularity": 0,
      "status": 1
    }
  ]
}
```

#### 可能错误

| 状态码 | 场景 |
| --- | --- |
| `400` | 请求体格式错误或缺少 `tag_name` |
| `500` | 服务端内部错误 |
