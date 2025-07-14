#!/bin/bash

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 打印带颜色的消息
print_message() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查 Docker 是否安装
if ! command -v docker &> /dev/null; then
    print_error "Docker 未安装。请先安装 Docker。"
    exit 1
fi

# 检查 Docker Compose 是否安装
if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    print_error "Docker Compose 未安装。请先安装 Docker Compose。"
    exit 1
fi

# 选择环境
echo "请选择运行环境："
echo "1) 开发环境 (Development)"
echo "2) 生产环境 (Production)"
read -p "请输入选择 (1 或 2): " choice

case $choice in
    1)
        COMPOSE_FILE="docker-compose.dev.yml"
        ENV_NAME="开发环境"
        ;;
    2)
        COMPOSE_FILE="docker-compose.yml"
        ENV_NAME="生产环境"
        ;;
    *)
        print_error "无效的选择"
        exit 1
        ;;
esac

print_message "正在启动 $ENV_NAME..."

# 停止现有容器
print_message "停止现有容器..."
docker-compose -f $COMPOSE_FILE down

# 构建镜像
print_message "构建 Docker 镜像..."
docker-compose -f $COMPOSE_FILE build

# 启动服务
print_message "启动服务..."
docker-compose -f $COMPOSE_FILE up -d

# 等待服务启动
print_message "等待服务启动..."
sleep 10

# 检查服务状态
print_message "检查服务状态..."
docker-compose -f $COMPOSE_FILE ps

# 显示访问信息
echo ""
print_message "服务已启动！"
echo ""
echo "访问地址："
echo "- 前端应用: http://localhost:3000"
echo "- 后端 API: http://localhost:8080"

if [ "$choice" == "1" ]; then
    echo "- PHPMyAdmin: http://localhost:8081"
    echo "- Redis Commander: http://localhost:8082"
fi

echo ""
echo "查看日志："
echo "- 所有服务: docker-compose -f $COMPOSE_FILE logs -f"
echo "- 前端服务: docker-compose -f $COMPOSE_FILE logs -f frontend"
echo "- 后端服务: docker-compose -f $COMPOSE_FILE logs -f app"
echo ""
echo "停止服务: docker-compose -f $COMPOSE_FILE down"