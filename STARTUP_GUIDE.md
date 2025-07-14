# 币安交易系统启动指南

## 项目概述

这是一个完整的币安交易系统，包含前端（React + TypeScript）和后端（Go + Gin）。系统支持现货交易、期货交易、双币投资和自动提现等功能。

## 系统架构

- **前端**：React 18 + TypeScript + Vite + TailwindCSS
- **后端**：Go 1.23 + Gin + GORM + MySQL + Redis
- **部署**：Docker + Docker Compose + Nginx

## 快速启动

### 1. 环境准备

确保已安装以下软件：
- Docker 和 Docker Compose
- Node.js 18+ （开发环境）
- Go 1.23+ （开发环境）

### 2. 配置环境变量

```bash
# 复制环境变量示例文件
cp .env.example .env

# 编辑 .env 文件，设置以下必需的值：
# ENCRYPTION_KEY - 必须是32字符的安全密钥，例如：
# ENCRYPTION_KEY=your32characterencryptionkeyhere

# JWT_SECRET - 设置一个安全的JWT密钥，例如：
# JWT_SECRET=your-super-secret-jwt-key-here

# 数据库密码（可选，使用默认值也可以）
# DB_PASSWORD=yourpassword
```

### 3. 启动方式

#### 方式一：使用 Docker Compose（推荐）

**生产环境：**
```bash
# 启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f
```

**开发环境：**
```bash
# 使用开发环境配置启动（包含 PHPMyAdmin 和 Redis Commander）
docker-compose -f docker-compose.dev.yml up -d

# 服务端口：
# - 应用：http://localhost:3000
# - API：http://localhost:8080
# - PHPMyAdmin：http://localhost:8081
# - Redis Commander：http://localhost:8082
```

#### 方式二：本地开发启动

**启动后端：**
```bash
cd backend

# 安装依赖
go mod download

# 确保 MySQL 和 Redis 正在运行
# 如果使用 docker-compose.dev.yml 启动的数据库：
# MySQL 端口: 3307
# Redis 端口: 6380

# 运行后端
go run main.go
```

**启动前端：**
```bash
cd frontend

# 安装依赖
npm install

# 启动开发服务器
npm run dev

# 访问：http://localhost:3000
```

### 4. 使用 Makefile（快捷方式）

```bash
# 查看所有可用命令
make help

# 启动开发环境
make dev

# 启动生产环境
make prod

# 停止所有服务
make stop

# 清理所有数据
make clean
```

## 默认账户

系统启动后会自动创建默认管理员账户：
- 用户名：`admin`
- 密码：`admin123456`

**重要：生产环境请立即修改默认密码！**

## 功能说明

### 1. 用户管理
- 用户注册需要管理员审批
- 支持普通用户和管理员角色
- 用户需要配置自己的币安API密钥

### 2. 现货交易策略
- 简单策略：基础的买卖策略
- 冰山策略：大单拆分执行
- 慢冰山策略：带时间间隔的冰山策略
- 网格策略：网格交易
- DCA定投策略：定期定额投资

### 3. 期货交易
- 支持USDT永续合约
- 自动同步持仓信息
- 风险管理功能

### 4. 双币投资
- 自动同步币安双币投资产品
- 支持自动认购策略

### 5. 自动提现
- 定时自动提现到指定地址
- 支持多币种配置

## 币安API配置

### 1. 获取币安API密钥
1. 登录[币安](https://www.binance.com)
2. 进入 API管理页面
3. 创建新的API密钥
4. 保存 API Key 和 Secret Key

### 2. 在系统中配置API密钥
1. 登录系统
2. 进入"个人信息"页面
3. 点击"API密钥管理"
4. 输入币安API密钥并保存

### 3. 测试网络（开发测试）
开发环境默认使用币安测试网络：
- [现货测试网](https://testnet.binance.vision/)
- [期货测试网](https://testnet.binancefuture.com/)

## 故障排查

### 1. 数据库连接失败
```bash
# 检查MySQL容器是否运行
docker ps | grep mysql

# 查看MySQL日志
docker logs binance_mysql

# 确认连接配置
# 开发环境：localhost:3307
# 生产环境：localhost:3306
```

### 2. Redis连接失败
```bash
# 检查Redis容器是否运行
docker ps | grep redis

# 查看Redis日志
docker logs binance_redis

# 测试连接
redis-cli -h localhost -p 6379 ping
```

### 3. 前端无法访问后端API
- 检查CORS配置（.env中的CORS_ORIGINS）
- 确认API代理配置（frontend/vite.config.ts）
- 查看浏览器控制台错误信息

### 4. 加密密钥错误
如果出现"encryption key must be 32 bytes"错误：
```bash
# 生成32字符密钥
openssl rand -hex 16

# 或使用任意32字符字符串
echo -n "your32characterencryptionkeyhere" | wc -c
```

## 生产环境部署

### 1. 安全建议
- 修改所有默认密码
- 使用强加密密钥
- 配置HTTPS（修改nginx配置）
- 限制数据库访问权限
- 定期备份数据

### 2. 性能优化
- 调整数据库连接池大小
- 配置Redis缓存策略
- 使用CDN加速前端资源
- 启用Gzip压缩

### 3. 监控建议
- 配置日志收集（ELK Stack）
- 设置系统监控（Prometheus + Grafana）
- 配置错误追踪（Sentry）
- 设置告警通知

## 常见问题

**Q: 如何重置管理员密码？**
A: 可以通过修改数据库或使用提供的重置脚本。

**Q: 支持哪些交易对？**
A: 系统会自动同步币安支持的所有交易对。

**Q: 策略执行频率是多少？**
A: 默认每分钟检查一次策略条件。

**Q: 如何备份数据？**
A: 使用 `docker exec` 执行 mysqldump 备份数据库。

## 技术支持

如遇到问题，请检查：
1. 系统日志：`docker-compose logs -f backend`
2. 数据库状态：使用PHPMyAdmin查看
3. Redis状态：使用Redis Commander查看

## 更新说明

系统支持热更新：
```bash
# 拉取最新代码
git pull

# 重新构建并启动
docker-compose up -d --build
```

---

祝您使用愉快！如有任何问题，请查看项目文档或提交Issue。