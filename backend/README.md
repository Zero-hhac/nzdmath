# math-backend

数学协会网站后端服务 — Go + Gin + GORM + MySQL + Redis + JWT

## 技术栈

- **语言**: Go 1.23+
- **Web 框架**: Gin
- **ORM**: GORM
- **数据库**: MySQL 8.x
- **缓存 / Session**: Redis 7.x
- **认证**: JWT + Redis
- **日志**: log/slog (标准库)
- **配置**: Viper

## 目录结构

```
math-backend/
├── cmd/server/main.go             # 程序入口
├── configs/
│   ├── config.yaml                # 当前配置（不入 git）
│   └── config.example.yaml        # 配置示例
├── internal/
│   ├── cache/                     # Redis 缓存封装
│   ├── config/                    # viper 配置加载
│   ├── consts/                    # 常量
│   ├── dto/                       # 通用 DTO
│   ├── handler/                   # HTTP 处理器
│   ├── middleware/                # 中间件
│   │   ├── cors.go                # CORS
│   │   ├── jwt.go                 # JWT 鉴权（支持 user/admin 前缀）
│   │   ├── admin_auth.go          # 管理员角色校验
│   │   └── ratelimit.go           # 限流
│   ├── model/                     # 数据库模型
│   ├── repository/                # Repository 层
│   ├── response/                  # 统一响应格式
│   ├── router/                    # 路由注册
│   ├── service/                   # 业务逻辑
│   └── utils/                     # 工具（JWT 等）
├── storage/uploads/               # 上传文件目录
└── go.mod
```

## 快速开始

### 1. 准备环境

```bash
# MySQL 已启动，创建一个数据库
mysql -u root -p
> CREATE DATABASE math_top DEFAULT CHARACTER SET utf8mb4;
> exit

# Redis 已启动
redis-cli ping  # 返回 PONG
```

### 2. 配置

```bash
cp configs/config.example.yaml configs/config.yaml
# 编辑 configs/config.yaml，填入 MySQL 用户名/密码、JWT secret 等
```

### 3. 启动

```bash
go mod tidy
go run cmd/server/main.go
```

服务监听 `:8080`。

## 默认管理员

服务首次启动时，若 `admins` 表为空，会自动创建：

- 用户名: `admin`
- 密码: `admin123`

**生产环境请立即修改！**

## 主要接口

### 公开接口（无需鉴权）

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/api/v1/home` | 首页聚合数据（最近活动、推荐资源、热门作品等，Redis 缓存 5 分钟） |
| GET | `/api/v1/events` | 活动列表 |
| GET | `/api/v1/events/:id` | 活动详情 |
| GET | `/api/v1/news` | 资讯列表 |
| GET | `/api/v1/news/:id` | 资讯详情 |
| GET | `/api/v1/resources` | 资源列表 |
| GET | `/api/v1/resources/:id` | 资源详情 |
| GET | `/api/v1/resources/download/:id` | 下载资源 |
| GET | `/api/v1/showcases` | 作品列表 |
| GET | `/api/v1/showcases/:id` | 作品详情 |
| GET | `/api/v1/comments?target_type=X&target_id=Y` | 查看评论（含一级回复） |
| POST | `/api/v1/auth/register` | 用户注册 |
| POST | `/api/v1/auth/login` | 用户登录 |
| POST | `/api/v1/auth/logout` | 用户退出 |

### 鉴权接口（需 user token）

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/api/v1/profile` | 获取个人资料 |
| PUT | `/api/v1/profile` | 更新个人资料 |
| POST | `/api/v1/user/avatar` | 上传头像（multipart） |
| POST | `/api/v1/auth/change-password` | 修改密码 |
| GET/POST/DELETE | `/api/v1/member/favorites` | 收藏 |
| GET | `/api/v1/member/downloads` | 我的下载历史 |
| POST/DELETE | `/api/v1/comments` | 发评论/回复/删自己的 |

### 后台接口（需 admin token）

| 方法 | 路径 | 说明 |
|---|---|---|
| POST | `/api/v1/admin/auth/login` | 管理员登录 |
| POST | `/api/v1/admin/auth/logout` | 管理员退出 |
| GET | `/api/v1/admin/dashboard` | 后台仪表盘 |
| POST | `/api/v1/admin/homepage/invalidate` | 失效首页缓存 |
| CRUD | `/api/v1/admin/{events,news,resources,showcases}` | 4 个内容模块的完整 CRUD |
| PATCH | `/api/v1/admin/events/:id/feature` | 切换活动推荐 |
| GET | `/api/v1/admin/users` | 用户列表 |
| PATCH | `/api/v1/admin/users/:id/status` | 启用/禁用 |
| POST | `/api/v1/admin/users/:id/reset-password` | 强制重置密码（管理员在表单中填新密码） |
| GET/DELETE | `/api/v1/admin/comments` | 评论管理 |

## 响应格式

所有接口统一返回：

```json
{
  "code": 200,
  "msg": "success",
  "data": { ... }
}
```

分页接口额外包含 `total`、`page`、`pageSize`。

错误码与 HTTP 状态码一致：400 参数错、401 未登录、403 无权限、404 不存在、429 限流、500 服务异常。

## 架构要点

- **软删除**: 所有业务表都有 `deleted_at`，GORM 默认忽略已删除记录
- **Repository 层**: 在 `internal/repository/`，按模型拆分
- **Redis 缓存**:
  - 首页聚合 5 分钟
  - 后台写操作触发 `cache:invalidate` 失效
- **双 Token**: 用户和管理员用不同的 Redis 前缀（`token:` / `admin_token:`），互不干扰
- **限流**: 登录/注册 10 次/分钟，管理员登录 20 次/分钟
- **JWT**: 密钥和过期时间全部从 config 读取
