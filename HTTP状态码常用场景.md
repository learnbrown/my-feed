# HTTP 状态码常用场景

这份速查表用于后端接口开发。原则很简单：状态码描述“请求在 HTTP 层面的处理结果”，响应体里的 `error` 或 `message` 再描述业务细节。

## 2xx：请求成功

| 状态码 | 名称 | 常用场景 | 项目示例 |
| --- | --- | --- | --- |
| 200 | OK | 请求成功并返回数据 | 登录成功、查询 `/account/me` 成功 |
| 201 | Created | 创建资源成功 | 注册账号成功、后续发布视频成功 |
| 204 | No Content | 请求成功但不需要返回响应体 | 删除、登出、取消点赞成功时可选 |

## 4xx：客户端请求有问题

| 状态码 | 名称 | 常用场景 | 项目示例 |
| --- | --- | --- | --- |
| 400 | Bad Request | 请求格式错误、缺少必填字段、JSON 绑定失败 | 注册缺少 `username` 或 `password` |
| 401 | Unauthorized | 未登录、token 缺失、token 过期、账号密码错误 | `/account/me` 未带 `Authorization` |
| 403 | Forbidden | 已登录，但没有权限操作该资源 | 删除别人的评论或视频 |
| 404 | Not Found | 请求的资源不存在 | 查询不存在的用户、视频、评论 |
| 405 | Method Not Allowed | 请求方法不对 | 用 `GET` 请求只支持 `POST` 的接口 |
| 409 | Conflict | 请求与当前资源状态冲突 | 用户名已存在、重复关注、重复点赞 |
| 415 | Unsupported Media Type | 请求体格式不支持 | 上传接口收到不支持的文件类型 |
| 422 | Unprocessable Entity | JSON 格式正确，但业务校验不通过 | 密码太短、用户名包含非法字符 |
| 429 | Too Many Requests | 请求过于频繁，被限流 | 登录失败次数过多、频繁点赞或评论 |

## 5xx：服务端处理失败

| 状态码 | 名称 | 常用场景 | 项目示例 |
| --- | --- | --- | --- |
| 500 | Internal Server Error | 服务端未知错误、数据库异常、代码异常 | 数据库写入失败、生成 token 失败 |
| 502 | Bad Gateway | 网关或反向代理收到异常响应 | Nginx 代理后端失败 |
| 503 | Service Unavailable | 服务暂时不可用 | MySQL、Redis、RabbitMQ 短暂不可用 |
| 504 | Gateway Timeout | 上游服务超时 | 网关等待 API 或 Worker 相关服务超时 |

## 在当前项目中的建议

注册接口：

```text
201 Created      注册成功
400 Bad Request  JSON 错误或缺少字段
409 Conflict     用户名已存在
500 Internal Server Error  数据库异常或密码哈希失败
```

登录接口：

```text
200 OK           登录成功
400 Bad Request  JSON 错误或缺少字段
401 Unauthorized 用户名或密码错误
500 Internal Server Error  token 生成或保存失败
```

鉴权接口：

```text
200 OK           token 有效，查询成功
401 Unauthorized token 缺失、格式错误、过期、已登出
404 Not Found    token 中的用户 ID 对应的账号不存在
500 Internal Server Error  数据库查询异常
```

写操作接口：

```text
200 OK / 201 Created  操作成功
400 Bad Request       参数错误
401 Unauthorized      未登录
403 Forbidden         没有权限
404 Not Found         目标资源不存在
409 Conflict          重复操作或状态冲突
500 Internal Server Error  服务端异常
```

## 几条实战规则

1. 参数绑定失败用 `400`，不要用 `500`。
2. 未登录、token 失效、账号密码错误用 `401`。
3. 登录了但不能操作别人的资源，用 `403`。
4. 用户名已存在、重复点赞、重复关注，用 `409`。
5. 数据库、缓存、消息队列异常才考虑 `500` 或 `503`。
6. 不要把所有错误都塞进 `200 OK`，这会让前端和调用方很痛苦。
