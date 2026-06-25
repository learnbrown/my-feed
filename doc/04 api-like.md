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
  "has_more": false,
  "likes": [
    {
      "ID": 1,
      "UpdatedAt": "2026-06-23T22:42:54.988391537+08:00",
      "DeletedAt": null,
      "author_id": 1,
      "title": "First video #GO #gin #gorm",
      "description": "feed system #development #GO",
      "play_url": "/static/uploads/videos/1/20260623/1782225767603097957_55811.mp4",
      "cover_url": "/static/uploads/covers/1/20260623/1782225772735292875_50825.png",
      "likes_count": 1,
      "comments_count": 0,
      "popularity": 0,
      "status": 1,
      "CreatedAt": "2026-06-23T22:42:54.988391537+08:00",
      "LikedAt": "2026-06-23T22:49:12.669080241+08:00"
    }
  ],
  "next_time": 1782226152669
}
```

#### 可能错误

| 状态码 | 场景 |
| --- | --- |
| `400` | 请求体格式错误 |
| `401` | 未登录、Token 缺失或 Token 无效 |
| `500` | 服务端内部错误 |