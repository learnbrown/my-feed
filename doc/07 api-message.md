## Message 接口

### 发送私信

向其他用户发送私信

| 项目 | 内容 |
| --- | --- |
| URL | `/message/sendMsg` |
| Method | `POST` |
| Auth | 需要 |
| Content-Type | `application/json` |

#### 请求参数

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `to_id` | number | 是 | 接收用户id |
| `content` | string | 是 | 私信内容 |

#### 请求示例

```json
{
  "to_id": 1,
  "content": "这是一条测试消息"
}
```

#### 成功响应

Status: `200 OK`

```json
{
  "message": {
    "id": 16,
    "from_id": 2,
    "to_id": 1,
    "content": "这是一条测试消息",
    "created_at": 1782485270704
  }
}
```

#### 可能错误

| 状态码 | 场景 |
| --- | --- |
| `400` | 请求体格式错误 |
| `401` | 未登录、Token 缺失或 Token 无效 |
| `404` | 用户不存在 |
| `422` | 不能关注自己 |
| `500` | 服务端内部错误 |

### 查看私信

查询和某个用户的最近私信

| 项目 | 内容 |
| --- | --- |
| URL | `/message/listConversation` |
| Method | `POST` |
| Auth | 需要 |
| Content-Type | `application/json` |

#### 请求参数

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `to_id` | number | 是 | 接收用户id |
| `limit` | number | 否 | 返回数量，默认 `20`，最大 `50` |
| `latest_time` | number | 否 | 分页游标，毫秒时间戳，第一页传 `0` |

#### 请求示例

```json
{
  "to_id": 1,
  "limit": 5,
  "latest_time": 0
}
```

#### 成功响应

Status: `200 OK`

```json
{
  "messages": [
    {
      "id": 16,
      "from_id": 2,
      "to_id": 1,
      "content": "这是一条测试消息",
      "created_at": 1782485270704
    }
  ],
  "has_more": true,
  "next_time": 1782399223154
}
```

#### 可能错误

| 状态码 | 场景 |
| --- | --- |
| `400` | 请求体格式错误 |
| `401` | 未登录、Token 缺失或 Token 无效 |
| `404` | 用户不存在 |
| `422` | 不能关注自己 |
| `500` | 服务端内部错误 |