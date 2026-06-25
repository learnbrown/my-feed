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