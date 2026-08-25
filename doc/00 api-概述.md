## API 文档

### 通用约定

- Base URL: `http://localhost:8080`
- JSON 接口请求头：`Content-Type: application/json`
- 需要登录的接口请求头：`Authorization: Bearer <token>`
- 所有列表分页使用 `latest_time + latest_id` 复合游标。第一页两个字段都传 `0` 或都不传；下一页传上一次响应中的 `next_time + next_id`。
- `latest_time` 使用毫秒时间戳。只传一个游标字段或传入负数 `latest_time` 均返回 `400 Bad Request`。
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
| `id` | number | 视频 ID |
| `author_id` | number | 作者用户 ID |
| `title` | string | 视频标题 |
| `description` | string | 视频描述 |
| `play_url` | string | 视频播放地址 |
| `cover_url` | string | 视频封面地址 |
| `likes_count` | number | 点赞数 |
| `comments_count` | number | 评论数 |
| `created_at` | number | 创建时间，毫秒时间戳 |

接口返回 `VideoDTO`，不会暴露 GORM 的 `UpdatedAt`、`DeletedAt`、内部状态或热度字段。
