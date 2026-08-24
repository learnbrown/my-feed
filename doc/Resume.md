# 丁昊天

**手机：** 158xxxx1991  
**邮箱：** reinerbrown.dev@gmail.com  
**GitHub：** https://github.com/learnbrown/my-feed  
**求职意向：** Go 后端开发实习  
**期望城市：** 北京  
**实习形式：** 可线下实习  

---

## 教育背景

**长安大学（211工程 / 双一流）｜软件工程｜本科**  
2023.09 - 2027.07

- GPA：约 3.8
- 专业排名：4 / 70
- 英语：CET-6 460
- 荣誉：校级学业优秀奖学金 2 次；蓝桥杯 C++ A 组省级二等奖；校内 ACM 竞赛二等奖 2 次
- 主修课程：数据结构、操作系统、计算机网络、数据库系统、软件工程

---

## 专业技能

- **Go：** 掌握 Go 基础语法，能够使用 `slice`、`map`、`struct`、`interface`、`defer` 等完成业务开发；了解 `goroutine`、`channel` 基础用法。
- **Web 开发：** 使用 Gin 完成路由分组、参数绑定、JWT 中间件、统一响应、文件上传等接口开发。
- **数据库：** 使用 GORM + MySQL 完成基础 CRUD、事务、唯一索引约束、游标分页查询等功能。
- **缓存与鉴权：** 使用 JWT 和服务端 token 校验实现登录鉴权与登出失效；使用 Redis 缓存登录 token 和视频详情，了解缓存 miss 回源和写操作后删除缓存的处理方式。
- **工具：** 熟悉 Linux 基础命令和 Git 基础操作；使用 Bruno 进行接口联调；了解 Docker、Go testing 和 `httptest` 基础用法。

---

## 项目经历

### 短视频 Feed 后端系统

**技术栈：** Go、Gin、GORM、MySQL、Redis、JWT、bcrypt  
**项目地址：** https://github.com/learnbrown/my-feed

个人 Go 后端项目，围绕短视频应用的基础业务场景，实现账号、视频发布、Feed 浏览、点赞、评论、关注、私信和用户主页等功能。

- 采用 `handler / service / repository` 分层组织代码，分别处理请求参数与响应、业务逻辑、数据库访问，降低接口层与数据访问层的耦合。
- 实现用户注册、登录、登出和个人信息接口，使用 `bcrypt` 存储密码摘要；使用 JWT 进行登录态校验，并在 Gin 中间件中解析 token、写入当前用户上下文。
- 在服务端保存当前有效 token，鉴权时结合 Redis token cache 和 MySQL token 字段进行校验，支持用户登出后旧 token 失效。
- 实现视频上传、封面上传、视频发布、详情查询和作者视频列表接口；发布视频时提取 `#tag`，并在事务中写入视频、标签和视频标签关联数据。
- 使用 DTO 控制接口返回字段，避免直接暴露 GORM Model 中的 `password_hash`、`token`、`DeletedAt` 等内部字段。
- 实现最新视频 Feed、标签 Feed、作者作品等列表接口，使用 `created_at + id` 复合游标分页，并通过查询 `limit + 1` 条记录判断是否存在下一页，避免 `offset` 分页在新增数据场景下可能出现的重复或漏读问题。
- 实现点赞、取消点赞、评论发布、评论删除等功能；在涉及关系记录和视频计数字段变更时使用 GORM 事务，保证多表更新一致性。
- 为点赞表设置“用户 ID + 视频 ID”唯一索引，为关注表设置 `follower_id + vlogger_id` 唯一索引；重复点赞或重复关注时识别唯一索引冲突并返回成功，保证接口幂等。
- 使用 Redis cache-aside 缓存账号 token 和视频详情信息；token 缓存 TTL 为 2 小时，视频详情缓存 TTL 为 5 分钟，点赞和评论变更后删除对应视频详情缓存。
- 使用 Bruno 整理接口请求并进行联调，尝试为部分 service 逻辑和 HTTP 接口补充 Go testing / `httptest` 测试。

---

## 其他

- 具备数据结构、操作系统、计算机网络、数据库系统等计算机基础课程学习经历。
- 了解 C++、Python 基础，曾完成 Qt 计算器、Crow 教学管理系统等课程练习。