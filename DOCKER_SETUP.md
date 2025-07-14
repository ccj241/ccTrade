# Docker 环境设置指南

本指南说明如何使用 Docker 运行 Binance Trading 系统的前后端应用。

## 前提条件

- 安装 Docker 和 Docker Compose
- 确保以下端口未被占用：
  - 3000 (前端)
  - 8080 (后端 API)
  - 3306/3307 (MySQL)
  - 6379/6380 (Redis)
  - 80 (Nginx，生产环境)

## 快速启动

### 使用启动脚本

```bash
# 给脚本添加执行权限
chmod +x start-docker.sh

# 运行脚本
./start-docker.sh
```

脚本会提示你选择运行环境（开发或生产）。

### 手动启动

#### 开发环境

```bash
# 构建并启动所有服务
docker-compose -f docker-compose.dev.yml up -d --build

# 查看日志
docker-compose -f docker-compose.dev.yml logs -f

# 停止服务
docker-compose -f docker-compose.dev.yml down
```

#### 生产环境

```bash
# 构建并启动所有服务
docker-compose up -d --build

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

## 访问地址

### 开发环境
- 前端应用: http://localhost:3000
- 后端 API: http://localhost:8080
- PHPMyAdmin: http://localhost:8081
- Redis Commander: http://localhost:8082

### 生产环境
- 前端应用: http://localhost:3000
- 后端 API: http://localhost:8080
- Nginx 代理: http://localhost

## 服务说明

### 前端服务
- 开发环境使用 Vite 开发服务器，支持热重载
- 生产环境使用 Nginx 提供静态文件服务
- API 请求会自动代理到后端服务

### 后端服务
- 提供 RESTful API
- 连接 MySQL 和 Redis
- 开发环境启用调试模式

### 数据库服务
- MySQL 8.0
- 自动创建数据库和表结构
- 数据持久化到 Docker 卷

### 缓存服务
- Redis 7
- 用于会话管理和缓存
- 数据持久化到 Docker 卷

## 常见问题

### 1. localhost:3000 无法访问

**问题原因**：
- Docker 容器未正确启动
- 端口被占用
- 防火墙阻止访问

**解决方案**：
```bash
# 检查容器状态
docker-compose ps

# 查看前端容器日志
docker-compose logs frontend

# 检查端口占用
lsof -i :3000

# 重启前端服务
docker-compose restart frontend
```

### 2. CORS 错误

**问题原因**：
- 后端 CORS 配置不正确
- API 地址配置错误

**解决方案**：
- 确保后端的 CORS 配置允许前端地址
- 检查前端的 API 代理配置

### 3. 数据库连接失败

**问题原因**：
- MySQL 服务未完全启动
- 连接参数错误

**解决方案**：
```bash
# 等待 MySQL 完全启动
docker-compose logs mysql

# 手动连接测试
docker exec -it binance_new_mysql_1 mysql -uroot -prootpassword
```

## 开发建议

1. **前端开发**：
   - 修改代码后会自动热重载
   - 可以直接在 `frontend/src` 目录修改代码

2. **后端开发**：
   - 使用 Air 工具实现热重载（需要在 Dockerfile.dev 中配置）
   - 或者手动重启后端容器：`docker-compose restart app`

3. **数据库管理**：
   - 使用 PHPMyAdmin 进行可视化管理
   - 或使用命令行：`docker exec -it binance_new_mysql_1 mysql -uroot -prootpassword`

4. **调试技巧**：
   - 查看实时日志：`docker-compose logs -f [service_name]`
   - 进入容器：`docker exec -it [container_name] /bin/sh`
   - 检查网络：`docker network ls`

## 性能优化

1. **前端优化**：
   - 生产环境使用 Nginx 提供静态文件
   - 启用 gzip 压缩
   - 设置静态资源缓存

2. **后端优化**：
   - 使用连接池
   - 启用 Redis 缓存
   - 合理设置超时时间

## 安全建议

1. **生产环境**：
   - 修改默认密码
   - 使用 HTTPS
   - 限制数据库端口访问
   - 使用环境变量管理敏感信息

2. **开发环境**：
   - 不要在公网暴露开发端口
   - 定期更新依赖包
   - 使用 .env 文件管理配置

## 部署到生产环境

1. 修改配置文件中的敏感信息
2. 使用 `docker-compose.yml`（生产配置）
3. 配置 SSL 证书
4. 设置防火墙规则
5. 配置日志收集和监控

## 故障排查命令

```bash
# 查看所有容器状态
docker ps -a

# 查看容器日志
docker logs [container_name]

# 检查容器内部
docker exec -it [container_name] /bin/sh

# 查看 Docker 网络
docker network inspect binance_new_default

# 清理并重建
docker-compose down -v
docker-compose up -d --build

# 查看磁盘使用
docker system df

# 清理未使用的资源
docker system prune -a
```