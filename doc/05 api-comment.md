## Comment 接口

### 发布评论

在视频下发布评论

| 项目 | 内容 |
| --- | --- |
| URL | `/comment/publish` |
| Method | `POST` |
| Auth | 需要 |
| Content-Type | `application/json` |

#### 请求参数

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `video_id` | number | 是 | 视频id |
| `content` | string | 是 | 评论内容 |

#### 请求示例

```json
{
  "video_id": 1,
  "content": "这是一条测试评论，111，111，111"
}
```

#### 成功响应

Status: `201 Created`

```json
{
  "comment": {
    "ID": 5,
    "video_id": 1,
    "account_id": 2,
    "content": "这是一条测试评论，111，111，111",
    "CreatedAt": "2026-06-24T17:09:00.942509029+08:00",
    "UpdatedAt": "2026-06-24T17:09:00.942509029+08:00",
    "DeletedAt": null
  },
  "comments_count": 4
}
```

#### 可能错误

| 状态码 | 场景 |
| --- | --- |
| `400` | 请求体格式错误 |
| `401` | 未登录、Token 缺失或 Token 无效 |
| `404` | 视频不存在 |
| `500` | 服务端内部错误 |

### 删除评论

删除评论

| 项目 | 内容 |
| --- | --- |
| URL | `/comment/delete` |
| Method | `POST` |
| Auth | 需要 |
| Content-Type | `application/json` |

#### 请求参数

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `comment_id` | number | 是 | 评论id |

#### 请求示例

```json
{
  "content_id": 1,
}
```

#### 成功响应

Status: `200 OK`

```json
{
  "comments_count": 2,
  "status": "ok"
}
```

#### 可能错误

| 状态码 | 场景 |
| --- | --- |
| `400` | 请求体格式错误 |
| `401` | 未登录、Token 缺失或 Token 无效 |
| `404` | 评论不存在 |
| `500` | 服务端内部错误 |

### 获取评论列表

列出视频评论区

| 项目 | 内容 |
| --- | --- |
| URL | `/comment/listComment` |
| Method | `POST` |
| Auth | 不需要 |
| Content-Type | `application/json` |

#### 请求参数

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `video_id` | number | 是 | 视频id |
| `limit` | number | 否 | 返回数量，默认 `20`，最大 `50` |
| `latest_time` | number | 否 | 分页游标，毫秒时间戳，第一页传 `0` |

#### 请求示例

```json
{
  "video_id": 1,
  "limit": 5,
  "latest_time": 0
}
```

#### 成功响应

Status: `200 OK`

```json
{
  "comments": [
    {
      "ID": 3,
      "video_id": 1,
      "account_id": 1,
      "content": "这是一条测试评论，111，111，111",
      "CreatedAt": "2026-06-24T16:28:02.639998916+08:00",
      "UpdatedAt": "2026-06-24T16:28:02.639998916+08:00",
      "DeletedAt": null
    }
  ],
  "has_more": false,
  "next_time": 1782289680750
}
```

#### 可能错误

| 状态码 | 场景 |
| --- | --- |
| `400` | 请求体格式错误 |
| `401` | 未登录、Token 缺失或 Token 无效 |
| `404` | 视频不存在 |
| `500` | 服务端内部错误 |