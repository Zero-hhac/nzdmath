# 数学协会全栈应用 — Docker 容器化一键部署指南

本项目已完成完整的 Docker 容器化及 Nginx 反向代理配置。您可以通过以下步骤，在安装有 Docker 和 Docker Compose 的环境中实现一键本地打包与部署运行。

## 目录结构说明

```
/nzdmath
├── docker-compose.yml        # 统一的多容器编排配置文件
├── backend/                  # 后端 Go 项目
│   ├── Dockerfile            # 后端多阶段构建 Dockerfile
│   └── configs/
│       └── config.docker.yaml # 容器专用的数据库与缓存配置文件
└── frontend/                 # 前端 React 项目
    ├── Dockerfile            # 前端打包与 Nginx 托管的多阶段 Dockerfile
    └── nginx.conf            # Nginx 托管规则与 API 反向代理配置文件
```

---

## 快速启动步骤

### 1. 启动 Docker Desktop
确保您本地的 Docker 已经启动并正常运行。

### 2. 构建并启动服务
在终端中进入 `/nzdmath` 目录，执行以下一键启动命令：

```bash
cd nzdmath
docker compose up --build -d
```

该命令将自动：
1. 下载并运行 **MySQL 8.0** 数据库容器（并挂载持久化卷 `mysql_data`）。
2. 下载并运行 **Redis 7** 缓存服务容器（并挂载持久化卷 `redis_data`）。
3. 编译并运行 **Go 后端** 容器（使用 `config.docker.yaml` 自动连接 MySQL 和 Redis）。
4. 打包 **React 前端** 静态文件，拷贝到 **Nginx** 容器中，并监听本机的 `80` 端口。

### 3. 访问项目
服务拉起后，直接在浏览器中打开：
- 门户网站及管理后台：[http://localhost](http://localhost)
- 后端服务接口地址：`http://localhost/api/v1`

---

## 运维与技术说明

1. **前后端接口联调**：
   - 所有的 API 请求（形如 `/api/v1/...`）在前端发起后，均会由 Nginx 自动反向代理至后端的 `http://backend:8080`，无需在前端代码中硬编码任何后端端口。
2. **文件上传路径持久化**：
   - 容器化时，后端的上传文件存放于 `/app/storage/uploads`。我们在 `docker-compose` 中挂载了命名卷 `uploads_data`，确保即使容器被销毁重建，用户上传的头像和作品文件也不会丢失。
3. **数据库自动迁移**：
   - 后端容器启动并等 MySQL 就绪后，会借助 GORM 自动执行迁移，自动创建所有必要的数据表，无需手动执行 SQL 初始化。

---

## ☁️ 阿里云 ECS 服务器部署特别注意事项

如果您的目标部署环境是 **阿里云 ECS 服务器（Linux 系统）**，请在部署前注意以下事项：

### 1. 将项目上传至服务器
您可以通过 `git` 或 `scp` 命令将本机的 `nzdmath` 目录上传至阿里云服务器中。例如通过 `scp` 上传：
```bash
scp -r ./nzdmath root@您的阿里云公网IP:/root/nzdmath
```

### 2. 配置阿里云安全组规则
默认情况下，阿里云服务器的端口是关闭的。请登录 **阿里云控制台 -> 云服务器 ECS -> 安全组**，添加入方向规则：
- **目的端口**：`80` (HTTP)
- **源地址**：`0.0.0.0/0` (允许所有人访问)
- *(注意：请勿向公网开放 3306 和 6379 端口，数据库与缓存只需在容器内网通过桥接网络安全互通即可)*

### 3. 国内镜像加速器配置
由于国内云服务器拉取 Docker Hub 镜像（如 `mysql:8.0` 和 `redis:7-alpine`）可能受限，建议在阿里云服务器上配置**镜像加速器**。
登录您的阿里云控制台，搜索 **“容器镜像服务 ACR”**，找到 **“镜像工具 -> 镜像加速器”**，根据页面提供的加速器地址，在服务器上执行以下命令：
```bash
sudo mkdir -p /etc/docker
sudo tee /etc/docker/daemon.json <<-'EOF'
{
  "registry-mirrors": ["您的阿里云专属加速器地址"]
}
EOF
sudo systemctl daemon-reload
sudo systemctl restart docker
```

### 4. 域名绑定（可选）
如果您购买了域名并解析到了您的阿里云服务器 IP，只需直接访问域名即可。如果需要将 Nginx 配置与域名绑定，可以修改 [nginx.conf](file:///Users/Admin/copy/test/math-top/nzdmath/frontend/nginx.conf) 中的 `server_name localhost;` 为 `server_name 您的域名;`，然后重新构建并启动前端容器。

