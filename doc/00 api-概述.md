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