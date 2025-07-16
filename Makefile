# ccTrade币安智能交易系统 - Makefile
# 作者：ccTrade团队
# 版本：1.0.0

.PHONY: help deploy update rollback backup restore clean status logs health monitor test

# 变量定义
COMPOSE_FILE = docker-compose.prod.yml
BACKUP_DIR = ./backups
PROJECT_NAME = cctrade

# 部署相关命令
deploy: ## 一键部署应用
	@echo "🚀 开始部署ccTrade系统..."
	@chmod +x deployment/scripts/deploy.sh
	@./deployment/scripts/deploy.sh

deploy-dev: ## 部署开发环境
	@echo "🔧 部署开发环境..."
	@docker-compose up -d

update: ## 更新应用到最新版本
	@echo "🔄 更新应用..."
	@chmod +x deployment/scripts/update.sh
	@./deployment/scripts/update.sh

rollback: ## 回滚到上一个版本
	@echo "⏪ 回滚应用..."
	@chmod +x deployment/scripts/rollback.sh
	@./deployment/scripts/rollback.sh

# 数据库相关命令
backup: ## 备份数据库
	@echo "💾 备份数据库..."
	@mkdir -p $(BACKUP_DIR)
	@docker exec cctrade-mysql-prod mysqldump -u root -p$$(grep DB_PASSWORD .env | cut -d'=' -f2) binance_trading > $(BACKUP_DIR)/backup_$$(date +%Y%m%d_%H%M%S).sql
	@echo "✅ 备份完成: $(BACKUP_DIR)/backup_$$(date +%Y%m%d_%H%M%S).sql"

restore: ## 恢复数据库 (用法: make restore file=backup_file.sql)
	@echo "🔄 恢复数据库..."
	@if [ -z "$(file)" ]; then echo "请指定备份文件: make restore file=backup_file.sql"; exit 1; fi
	@docker exec -i cctrade-mysql-prod mysql -u root -p$$(grep DB_PASSWORD .env | cut -d'=' -f2) binance_trading < $(file)
	@echo "✅ 数据库恢复完成"

# 运维相关命令
status: ## 检查服务状态
	@echo "📊 服务状态:"
	@docker-compose -f $(COMPOSE_FILE) ps

logs: ## 查看服务日志
	@docker-compose -f $(COMPOSE_FILE) logs -f

logs-backend: ## 查看后端日志
	@docker-compose -f $(COMPOSE_FILE) logs -f backend

logs-frontend: ## 查看前端日志
	@docker-compose -f $(COMPOSE_FILE) logs -f frontend

logs-nginx: ## 查看nginx日志
	@docker-compose -f $(COMPOSE_FILE) logs -f nginx

health: ## 运行健康检查
	@echo "🏥 运行健康检查..."
	@chmod +x deployment/scripts/health-check.sh
	@./deployment/scripts/health-check.sh

monitor: ## 查看系统监控
	@echo "📈 系统监控信息:"
	@echo "CPU使用率:"
	@top -bn1 | grep "Cpu(s)" | awk '{print $2 + $4 "%"}'
	@echo "内存使用率:"
	@free -h
	@echo "磁盘使用率:"
	@df -h /
	@echo "Docker容器状态:"
	@docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.MemPerc}}"

# 测试相关命令
test: ## 运行完整测试套件
	@echo "🧪 运行测试..."
	@chmod +x deployment/scripts/test.sh
	@./deployment/scripts/test.sh

test-api: ## 测试API接口
	@echo "🔍 测试API接口..."
	@curl -f http://localhost/api/health || echo "API测试失败"

test-frontend: ## 测试前端
	@echo "🎭 测试前端..."
	@curl -f http://localhost/health || echo "前端测试失败"

# 维护相关命令
clean: ## 清理系统资源
	@echo "🧹 清理系统资源..."
	@docker system prune -f
	@docker volume prune -f
	@echo "✅ 清理完成"

stop: ## 停止所有服务
	@echo "🛑 停止所有服务..."
	@docker-compose -f $(COMPOSE_FILE) down

restart: ## 重启所有服务
	@echo "🔄 重启所有服务..."
	@docker-compose -f $(COMPOSE_FILE) restart

# 开发相关命令
dev: ## 启动开发环境
	@echo "🔧 启动开发环境..."
	@docker-compose up -d

dev-stop: ## 停止开发环境
	@echo "🛑 停止开发环境..."
	@docker-compose down

dev-rebuild: ## 重建开发环境
	@echo "🔨 重建开发环境..."
	@docker-compose down
	@docker-compose build --no-cache
	@docker-compose up -d

# 安全相关命令
security-scan: ## 运行安全扫描
	@echo "🛡️ 运行安全扫描..."
	@chmod +x deployment/scripts/security-scan.sh
	@./deployment/scripts/security-scan.sh

ssl-renew: ## 更新SSL证书
	@echo "🔐 更新SSL证书..."
	@chmod +x deployment/scripts/ssl-renew.sh
	@./deployment/scripts/ssl-renew.sh

# 性能相关命令
performance: ## 性能测试
	@echo "⚡ 性能测试..."
	@chmod +x deployment/scripts/performance-test.sh
	@./deployment/scripts/performance-test.sh

optimize: ## 系统优化
	@echo "🔧 系统优化..."
	@chmod +x deployment/scripts/optimize.sh
	@./deployment/scripts/optimize.sh

# 帮助信息
help: ## 显示帮助信息
	@echo "ccTrade币安智能交易系统 - 可用命令:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "使用示例:"
	@echo "  make deploy          # 部署生产环境"
	@echo "  make status          # 查看服务状态"
	@echo "  make logs            # 查看所有服务日志"
	@echo "  make backup          # 备份数据库"
	@echo "  make health          # 运行健康检查"
	@echo ""

# 默认目标
.DEFAULT_GOAL := help