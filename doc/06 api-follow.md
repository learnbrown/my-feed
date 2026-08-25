## Follow 接口

### 关注博主

关注其他用户

| 项目 | 内容 |
| --- | --- |
| URL | `/follow/follow` |
| Method | `POST` |
| Auth | 需要 |
| Content-Type | `application/json` |

#### 请求参数

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `vlogger_id` | number | 是 | 博主id |

#### 请求示例

```json
{
  "vlogger_id": 1 
}
```

#### 成功响应

Status: `201 Created`

```json
{
  "status": "ok"
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

### 取消关注

取消关注其他用户

| 项目 | 内容 |
| --- | --- |
| URL | `/follow/unfollow` |
| Method | `POST` |
| Auth | 需要 |
| Content-Type | `application/json` |

#### 请求参数

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `vlogger_id` | number | 是 | 博主id |

#### 请求示例

```json
{
  "vlogger_id": 1 
}
```

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
| `400` | 请求体格式错误 |
| `401` | 未登录、Token 缺失或 Token 无效 |
| `422` | 不能取关自己 |
| `500` | 服务端内部错误 |

### 查看是否关注

查看是否关注了对方

| 项目 | 内容 |
| --- | --- |
| URL | `/follow/isFollowing` |
| Method | `POST` |
| Auth | 需要 |
| Content-Type | `application/json` |

#### 请求参数

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `vlogger_id` | number | 是 | 用户id |

#### 请求示例

```json
{
  "vlogger_id": 1 
}
```

#### 成功响应

Status: `200 OK`

```json
{
  "is_following": true
}
```

#### 可能错误

| 状态码 | 场景 |
| --- | --- |
| `400` | 请求体格式错误 |
| `401` | 未登录、Token 缺失或 Token 无效 |
| `500` | 服务端内部错误 |

### 查看粉丝列表

查看某人的粉丝列表

| 项目 | 内容 |
| --- | --- |
| URL | `/follow/listFollower` |
| Method | `POST` |
| Auth | 不需要 |
| Content-Type | `application/json` |

#### 请求参数

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `account_id` | number | 是 | 博主id |
| `limit` | number | 否 | 返回数量，默认 `20`，最大 `50` |
| `latest_time` | number | 否 | 分页游标，毫秒时间戳，第一页传 `0` |
| `latest_id` | number | 否 | 分页游标，第一页传`0` |

#### 请求示例

```json
{
  "account_id": 1,
  "latest_time": 0,
  "latest_id": 0
}
```

#### 成功响应

Status: `200 OK`

```json
{
  "accounts": [
    {
      "id": 2,
      "username": "user2",
      "avatar_url": "",
      "bio": "",
      "followed_at": 1782481739250
    }
  ],
  "has_more": false,
  "next_time": 1782481739250,
  "next_id": 8
}
```

#### 可能错误

| 状态码 | 场景 |
| --- | --- |
| `400` | 请求体格式错误 |
| `500` | 服务端内部错误 |

### 查看关注列表

查看某人的关注列表

| 项目 | 内容 |
| --- | --- |
| URL | `/follow/listFollowing` |
| Method | `POST` |
| Auth | 不需要 |
| Content-Type | `application/json` |

#### 请求参数

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `account_id` | number | 是 | 博主id |
| `limit` | number | 否 | 返回数量，默认 `20`，最大 `50` |
| `latest_time` | number | 否 | 分页游标，毫秒时间戳，第一页传 `0` |
| `latest_id` | number | 否 | 分页游标，第一页传`0` |

#### 请求示例

```json
{
  "account_id": 2,
  "latest_time": 0,
  "latest_id": 0
}
```

#### 成功响应

Status: `200 OK`

```json
{
  "accounts": [
    {
      "id": 1,
      "username": "user",
      "avatar_url": "",
      "bio": "",
      "followed_at": 1782481739250
    }
  ],
  "has_more": false,
  "next_time": 1782481739250,
  "next_id": 8
}
```

#### 可能错误

| 状态码 | 场景 |
| --- | --- |
| `400` | 请求体格式错误 |
| `500` | 服务端内部错误 |
