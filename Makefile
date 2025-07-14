.PHONY: build run test clean docker-dev docker-prod help

# 默认目标
help:
	@echo "可用的命令:"
	@echo "  build      - 编译项目"
	@echo "  run        - 运行项目"
	@echo "  test       - 运行测试"
	@echo "  clean      - 清理构建文件"
	@echo "  docker-dev - 启动开发环境"
	@echo "  docker-prod- 启动生产环境"
	@echo "  lint       - 代码检查"
	@echo "  fmt        - 格式化代码"

# 编译项目
build:
	@echo "编译项目..."
	cd backend && go build -o ../bin/main .

# 运行项目
run:
	@echo "运行项目..."
	cd backend && go run main.go

# 运行测试
test:
	@echo "运行测试..."
	cd backend && go test ./... -v

# 清理构建文件
clean:
	@echo "清理构建文件..."
	rm -f bin/main
	rm -f backend/main

# 代码格式化
fmt:
	@echo "格式化代码..."
	cd backend && go fmt ./...

# 代码检查
lint:
	@echo "代码检查..."
	cd backend && go vet ./...

# 启动开发环境
docker-dev:
	@echo "启动开发环境..."
	docker-compose -f docker-compose.dev.yml up -d

# 启动生产环境  
docker-prod:
	@echo "启动生产环境..."
	docker-compose up -d

# 停止Docker环境
docker-stop:
	@echo "停止Docker环境..."
	docker-compose down

# 安装依赖
deps:
	@echo "安装依赖..."
	cd backend && go mod download && go mod tidy

# 初始化项目
init:
	@echo "初始化项目..."
	cp .env.example backend/.env
	@echo "请编辑 backend/.env 文件配置数据库连接"

# 创建二进制目录
bin:
	mkdir -p bin