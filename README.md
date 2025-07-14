# 📈 ccTrade - 币安智能交易系统

[![Go Version](https://img.shields.io/badge/Go-1.23+-blue.svg)](https://golang.org)
[![React Version](https://img.shields.io/badge/React-18+-blue.svg)](https://reactjs.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Docker](https://img.shields.io/badge/Docker-Ready-brightgreen.svg)](https://www.docker.com/)

> 🚀 **一个功能完整的币安交易系统，支持现货、期货和双币投资等多种交易模式**

## ✨ 功能特性

### 🎯 核心交易功能

#### 现货交易策略
- **📊 简单策略**: 基础价格触发订单
- **🧊 冰山策略**: 大单拆分执行，隐藏真实交易意图
- **⏱️ 慢冰山策略**: 带时间间隔的冰山订单
- **🕸️ 网格策略**: 在价格区间内自动低买高卖
- **💰 DCA定投**: 定时定额投资
- **🔀 套利策略**: 跨市场套利机会捕捉
- **📈 突破策略**: 价格突破关键技术位时自动交易
- **⚡ 动量策略**: 基于RSI、MACD等指标的趋势跟踪
- **📊 均值回归**: 价格偏离均线时交易
- **🌊 波段交易**: 捕捉短期价格波动
- **🔺 金字塔加仓**: 盈利时逐步加仓
- **🎯 信号策略**: 技术指标组合信号交易

#### 期货交易
- **⚡ 杠杆交易**: 支持1-125倍杠杆
- **🛡️ 止盈止损**: 自动风险控制
- **🧊 期货冰山**: 大额杠杆订单拆分
- **🔄 自动重启**: 策略完成后自动重新开始
- **💼 逐仓/全仓**: 支持两种保证金模式

#### 双币投资
- **🎯 单次投资**: 手动选择产品投资
- **🔄 自动复投**: 到期自动再投资
- **📊 梯度投资**: 根据深度分配投资
- **💰 价格触发**: 达到目标价格时投资

### 🛠️ 系统管理

#### 👥 用户管理
- **🔐 身份认证**: JWT + 角色权限控制
- **🔑 API密钥管理**: AES-256加密存储
- **👤 用户状态**: pending/active/disabled

#### 📋 订单管理
- **📝 订单创建**: 支持多种订单类型
- **🗂️ 批量管理**: 批量取消订单
- **📊 历史记录**: 完整交易记录
- **⏰ 自动取消**: 超时自动取消功能

#### 💎 资产管理
- **💰 余额查询**: 实时资产余额
- **🏦 提币管理**: 自动提币规则设置
- **📈 收益统计**: 详细收益分析

#### 🤖 后台任务
- **📊 价格监控**: 30秒更新一次价格
- **🔍 订单检查**: 30秒检查一次订单状态
- **🏦 提币检查**: 5分钟检查一次提币规则
- **💎 双币投资**: 1小时检查一次投资机会
- **⚡ 期货监控**: 1分钟监控一次期货策略

## 🏗️ 技术架构

### 后端技术栈
- **🔧 语言**: Go 1.23.0
- **🌐 框架**: Gin 1.10.1
- **🗄️ 数据库**: MySQL 8.0 + GORM 1.30.0
- **⚡ 缓存**: Redis 7.0
- **🔐 认证**: JWT (golang-jwt/jwt/v5)
- **🔒 加密**: AES-256
- **📡 交易所API**: github.com/adshao/go-binance/v2
- **🔌 WebSocket**: gorilla/websocket
- **🧪 测试**: testify

### 前端技术栈
- **⚛️ 框架**: React 18.2.0 + TypeScript 5.3.3
- **🏗️ 构建工具**: Vite 5.0.4
- **🎨 样式**: TailwindCSS
- **📡 数据获取**: React Query (TanStack Query)
- **🧭 路由**: React Router v6
- **🏪 状态管理**: Zustand
- **📝 表单**: React Hook Form + Zod
- **🎯 图标**: Lucide React

### 部署与运维
- **🐳 容器化**: Docker + Docker Compose
- **🌐 代理**: Nginx
- **📊 监控**: 健康检查端点
- **📝 日志**: 结构化JSON日志
- **🔒 安全**: SSL/TLS, CORS, 输入验证

## 🚀 快速开始

### 环境要求
- **Go**: 1.23+
- **Node.js**: 18+
- **MySQL**: 8.0+
- **Redis**: 7.0+
- **Docker**: 20.10+ (可选)

### 📁 项目结构
```
ccTrade/
├── backend/              # 后端Go代码
│   ├── controllers/      # 控制器层
│   ├── services/         # 业务逻辑层
│   ├── models/           # 数据模型
│   ├── middleware/       # 中间件
│   ├── routes/           # 路由定义
│   ├── utils/            # 工具函数
│   ├── tasks/            # 定时任务
│   └── main.go           # 主入口
├── frontend/             # 前端React代码
│   ├── src/
│   │   ├── components/   # 组件
│   │   ├── pages/        # 页面
│   │   ├── hooks/        # 自定义Hooks
│   │   ├── stores/       # 状态管理
│   │   ├── api/          # API客户端
│   │   └── types/        # 类型定义
│   ├── public/           # 静态资源
│   └── package.json      # 依赖配置
├── nginx/                # Nginx配置
├── docker-compose.yml    # Docker编排
├── Makefile             # 构建脚本
└── README.md
```

### 💻 本地开发

1. **克隆项目**
   ```bash
   git clone https://github.com/ccj241/ccTrade.git
   cd ccTrade
   ```

2. **环境配置**
   ```bash
   cp .env.example .env
   # 编辑.env文件，配置数据库和API密钥
   ```

3. **后端启动**
   ```bash
   cd backend
   go mod download
   go run main.go
   ```

4. **前端启动**
   ```bash
   cd frontend
   npm install
   npm run dev
   ```

### 🐳 Docker快速部署

#### 开发环境
```bash
make docker-dev
# 或者
docker-compose -f docker-compose.dev.yml up -d
```

#### 生产环境
```bash
make docker-prod
# 或者
docker-compose up -d
```

### 🛠️ Make命令

```bash
# 构建项目
make build

# 运行后端
make run

# 运行测试
make test

# 开发环境
make docker-dev

# 生产环境
make docker-prod

# 清理
make clean
```

## 📡 API接口文档

### 🔐 认证接口
| 方法 | 路径 | 描述 |
|-----|------|------|
| POST | `/api/register` | 用户注册 |
| POST | `/api/login` | 用户登录 |
| GET | `/api/profile` | 获取用户信息 |
| PUT | `/api/profile` | 更新用户信息 |
| POST | `/api/change-password` | 修改密码 |
| POST | `/api/api-keys` | 设置API密钥 |

### 📊 策略管理
| 方法 | 路径 | 描述 |
|-----|------|------|
| GET | `/api/strategies` | 获取策略列表 |
| POST | `/api/strategies` | 创建策略 |
| GET | `/api/strategies/:id` | 获取策略详情 |
| PUT | `/api/strategies/:id` | 更新策略 |
| POST | `/api/strategies/:id/toggle` | 启用/禁用策略 |
| DELETE | `/api/strategies/:id` | 删除策略 |

### 📋 订单管理
| 方法 | 路径 | 描述 |
|-----|------|------|
| GET | `/api/orders` | 获取订单列表 |
| POST | `/api/order` | 创建订单 |
| DELETE | `/api/order/:id` | 取消订单 |
| POST | `/api/batch-cancel-orders` | 批量取消订单 |

### ⚡ 期货交易
| 方法 | 路径 | 描述 |
|-----|------|------|
| GET | `/api/futures/strategies` | 期货策略列表 |
| POST | `/api/futures/strategy` | 创建期货策略 |
| GET | `/api/futures/positions` | 持仓查询 |
| GET | `/api/futures/stats` | 统计信息 |

### 💎 双币投资
| 方法 | 路径 | 描述 |
|-----|------|------|
| GET | `/api/dual/products` | 产品列表 |
| POST | `/api/dual/strategy` | 创建投资策略 |
| GET | `/api/dual/orders` | 订单列表 |
| GET | `/api/dual/stats` | 收益统计 |

### 💰 资产管理
| 方法 | 路径 | 描述 |
|-----|------|------|
| GET | `/api/balance` | 账户余额 |
| GET | `/api/withdrawals` | 提币规则 |
| POST | `/api/withdrawals` | 创建提币规则 |
| GET | `/api/withdrawals/history` | 提币历史 |

## ⚙️ 配置说明

### 🌍 环境变量

| 变量名 | 描述 | 默认值 |
|--------|------|--------|
| `SERVER_HOST` | 服务器主机 | `0.0.0.0` |
| `SERVER_PORT` | 服务器端口 | `8080` |
| `DB_HOST` | 数据库主机 | `localhost` |
| `DB_PORT` | 数据库端口 | `3306` |
| `DB_USERNAME` | 数据库用户名 | `root` |
| `DB_PASSWORD` | 数据库密码 | `` |
| `DB_DATABASE` | 数据库名 | `binance_trading` |
| `REDIS_HOST` | Redis主机 | `localhost` |
| `REDIS_PORT` | Redis端口 | `6379` |
| `JWT_SECRET` | JWT密钥 | `your-secret-key` |
| `ENCRYPTION_KEY` | 加密密钥(32字符) | `your-32-character-key-here!!` |
| `BINANCE_TESTNET` | 是否使用测试网 | `false` |

### 🐳 Docker环境变量
```yaml
services:
  backend:
    environment:
      - DB_HOST=mysql
      - REDIS_HOST=redis
      - SERVER_PORT=8080
```

## 🔒 安全特性

- **🔐 API密钥加密**: 使用AES-256加密存储
- **🎫 JWT认证**: 安全的令牌认证机制
- **🛡️ CORS保护**: 可配置的跨域策略
- **✅ 输入验证**: 中间件层面的参数验证
- **👥 权限控制**: 基于角色的访问控制(RBAC)
- **🛡️ SQL注入防护**: GORM预编译语句
- **📝 审计日志**: 完整的活动日志记录
- **🔒 HTTPS**: 强制SSL/TLS加密传输
- **🚫 限流保护**: API请求频率限制

## 🧪 测试

### 运行测试套件
```bash
# 后端测试
cd backend
go test ./...

# 前端测试
cd frontend
npm test

# E2E测试
npm run test:e2e
```

### 测试覆盖率
```bash
# 生成覆盖率报告
make test-coverage

# 查看覆盖率
open coverage.html
```

## 🚀 生产部署

### 📋 部署检查清单

- [ ] 🔐 配置生产环境变量
- [ ] 📜 设置SSL证书
- [ ] 🌐 配置Nginx反向代理
- [ ] 📊 设置监控和日志
- [ ] 💾 配置数据库备份
- [ ] 🔒 设置防火墙规则
- [ ] 📈 配置性能监控
- [ ] 🚨 设置告警系统

### 🐳 Docker生产部署

1. **克隆项目**
   ```bash
   git clone https://github.com/ccj241/ccTrade.git
   cd ccTrade
   ```

2. **配置环境**
   ```bash
   cp .env.example .env
   # 编辑生产环境配置
   ```

3. **启动服务**
   ```bash
   docker-compose up -d
   ```

4. **健康检查**
   ```bash
   curl http://localhost/health
   ```

### 📊 监控指标

- **🔍 系统监控**: CPU、内存、磁盘使用率
- **📡 API监控**: 响应时间、错误率、QPS
- **🗄️ 数据库监控**: 连接数、查询性能
- **⚡ 交易监控**: 订单成功率、延迟统计
- **💰 资产监控**: 余额变化、盈亏统计

## 🤝 贡献指南

1. **Fork项目**
2. **创建特性分支** (`git checkout -b feature/AmazingFeature`)
3. **提交更改** (`git commit -m 'Add some AmazingFeature'`)
4. **推送到分支** (`git push origin feature/AmazingFeature`)
5. **创建Pull Request**

### 📝 代码规范

- **Go**: 遵循Go官方代码规范
- **React**: 使用ESLint + Prettier
- **提交信息**: 使用约定式提交格式
- **测试**: 新功能必须包含测试用例

## 📄 许可证

本项目使用 [MIT License](LICENSE) 开源协议。

## 🆘 支持与反馈

- **🐛 Bug报告**: [提交Issue](https://github.com/ccj241/ccTrade/issues)
- **💡 功能建议**: [功能请求](https://github.com/ccj241/ccTrade/issues)
- **📧 邮箱**: support@cctrade.com
- **💬 讨论**: [GitHub Discussions](https://github.com/ccj241/ccTrade/discussions)

## 🎯 路线图

- [ ] 🔄 **v2.1.0**: 增加更多技术指标
- [ ] 📊 **v2.2.0**: 优化用户界面
- [ ] 🤖 **v2.3.0**: 增加AI交易建议
- [ ] 📱 **v2.4.0**: 移动端应用
- [ ] 🔗 **v2.5.0**: 支持更多交易所

## ⭐ 如果这个项目对您有帮助，请给我们一个Star！

---

<div align="center">
  <b>🚀 开始您的智能交易之旅吧！</b>
</div>