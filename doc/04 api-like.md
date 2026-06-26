## Like 接口

### 视频点赞

根据当前用户，对视频点赞

| 项目 | 内容 |
| --- | --- |
| URL | `/like/like` |
| Method | `POST` |
| Auth | 需要 |
| Content-Type | `application/json` |

#### 请求参数

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `video_id` | number | 是 | 视频id |

#### 请求示例

```json
{
  "video_id": 1
}
```

#### 成功响应

Status: `200 OK`

```json
{
  "likes_count": 1,
  "status": "ok"
}
```

#### 可能错误

| 状态码 | 场景 |
| --- | --- |
| `400` | 请求体格式错误 |
| `404` | 视频不存在 |
| `500` | 服务端内部错误 |

### 视频取消点赞

根据当前用户，取消对视频的点赞

| 项目 | 内容 |
| --- | --- |
| URL | `/like/unlike` |
| Method | `POST` |
| Auth | 需要 |
| Content-Type | `application/json` |

#### 请求参数

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `video_id` | number | 是 | 视频id |

#### 请求示例

```json
{
  "video_id": 1
}
```

#### 成功响应

Status: `200 OK`

```json
{
  "likes_count": 0,
  "status": "ok"
}
```

#### 可能错误

| 状态码 | 场景 |
| --- | --- |
| `400` | 请求体格式错误 |
| `401` | 未登录、Token 缺失或 Token 无效 |
| `404` | 视频不存在 |
| `500` | 服务端内部错误 |

### 查询视频是否点赞

根据当前用户，查询是否对视频点赞

| 项目 | 内容 |
| --- | --- |
| URL | `/like/isLiked` |
| Method | `POST` |
| Auth | 需要 |
| Content-Type | `application/json` |

#### 请求参数

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `video_id` | number | 是 | 视频id |

#### 请求示例

```json
{
  "video_id": 1
}
```

#### 成功响应

Status: `200 OK`

```json
{
  "is_liked": false
}
```

#### 可能错误

| 状态码 | 场景 |
| --- | --- |
| `400` | 请求体格式错误 |
| `401` | 未登录、Token 缺失或 Token 无效 |
| `500` | 服务端内部错误 |

### 列出已点赞视频

根据当前用户，列出已点赞的视频

| 项目 | 内容 |
| --- | --- |
| URL | `/like/listLikedVideos` |
| Method | `POST` |
| Auth | 需要 |
| Content-Type | `application/json` |

#### 请求参数

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `limit` | number | 否 | 返回数量，默认 `20`，最大 `50` |
| `latest_time` | number | 否 | 分页游标，毫秒时间戳，第一页传 `0` |

#### 请求示例

```json
{
  "limit": 5,
  "latest_time": 0
}
```

#### 成功响应

Status: `200 OK`

```json
{
  "has_more": true,
  "likes": [
    {
      "id": 28,
      "author_id": 1,
      "title": "Urban Cityscape In Overcast Weather",
      "description": "这是一段由 Pexels 创作者 Yunus KARA 拍摄的低分辨率测试视频。极小体积，适合本地压测。ID: 36194544。 #city #test_data #low_res",
      "play_url": "/static/uploads/videos/1/20260625/1782371434653425496_00768.mp4",
      "cover_url": "/static/uploads/covers/1/20260625/1782371434662272057_58628.jpg",
      "likes_count": 1,
      "comments_count": 0,
      "created_at": 1782371434668,
      "liked_at": 1782481722006
    }
  ],
  "next_time": 1782371254336
}
```

#### 可能错误

| 状态码 | 场景 |
| --- | --- |
| `400` | 请求体格式错误 |
| `401` | 未登录、Token 缺失或 Token 无效 |
| `500` | 服务端内部错误 |