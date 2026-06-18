# My Feed System

[Golang找到实习需要什么？](https://www.bilibili.com/video/BV1njAwzcEeJ)

根据以上视频，在AI指导下，自行实现视频feed流系统

## V0.1X 账号注册 / 登录 / JWT / 基础分层

将数据库、路由、中间件、JWT令牌方法模型定义处理放入不同模块中实现

### 路由Router
仅进行路由注册，分组的操作，不进行任何业务逻辑

### JWT生成与解析方法
将用户名与id加密生成token
解密token获得用户名与id

### Login路由
登录成功后生成token，更新数据库字段

### JWT中间件
验证请求头中的token
与数据库token字段比对
解析获取用户名与id，放入gin上下文

### 需鉴权路由
应用JWT中间件，在访问时验证token有效性

### handler -> service -> repository 三层分明

#### Model 

定义表结构模型

#### Handler 层：只做输入输出

接收请求参数，调用service，处理错误，返回信息

不进行业务逻辑和数据库操作

#### Service 层：只做业务逻辑

接收handler的参数，实现注册、登录、登出业务逻辑，

#### Repository 层：只做数据库读写 (CRUD)

实现对数据库的增删改查