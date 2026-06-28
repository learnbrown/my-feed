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
    "id": 32,
    "author_id": 2,
    "title": "First video #GO #gin #gorm",
    "description": "feed system #development #GO",
    "play_url": "/static/uploads/videos/2/20260626/1782484933623067701_59746.mp4",
    "cover_url": "/static/uploads/covers/default.png",
    "likes_count": 0,
    "comments_count": 0,
    "created_at": 1782484935444
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
    "id": 26,
    "author_id": 1,
    "title": "Drone Aerial View Of Seattle Skyline With Space Needle",
    "description": "这是一段由 Pexels 创作者 Josh Hild 拍摄的低分辨率测试视频。极小体积，适合本地压测。ID: 35033742。 #city #test_data #low_res",
    "play_url": "/static/uploads/videos/1/20260625/1782371434559115240_86096.mp4",
    "cover_url": "/static/uploads/covers/1/20260625/1782371434567647409_19104.jpg",
    "likes_count": 0,
    "comments_count": 0,
    "created_at": 1782371434573
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
| `latest_id` | number | 否 | 分页游标，第一页传`0` |

#### 请求示例

```json
{
  "author_id": 1,
  "limit": 10,
  "latest_time": 0,
  "latest_id": 0
}
```

#### 成功响应

Status: `200 OK`

```json
{
  "has_more": true,
  "next_time": 1782371434346,
  "next_id": 8,
  "videos": [
    {
      "id": 31,
      "author_id": 1,
      "title": "First video #GO #gin #gorm",
      "description": "feed system #development #GO",
      "play_url": "/static/uploads/videos/1/20260626/1782482677973872449_49663.mp4",
      "cover_url": "/static/uploads/covers/default.png",
      "likes_count": 0,
      "comments_count": 0,
      "created_at": 1782482684898
    }
  ]
}
```

#### 可能错误

| 状态码 | 场景 |
| --- | --- |
| `400` | 请求体格式错误或缺少 `author_id` |
| `500` | 服务端内部错误 |