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