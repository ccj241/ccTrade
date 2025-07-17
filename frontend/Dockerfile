# 前端Dockerfile - 生产优化构建
FROM node:18-alpine AS builder

# 安装依赖
RUN apk add --no-cache git

# 设置工作目录
WORKDIR /app

# 复制package文件
COPY package*.json ./

# 安装依赖（包括开发依赖，因为需要用于构建）
RUN npm ci && npm cache clean --force

# 复制源代码
COPY . .

# 构建应用（跳过类型检查以进行生产部署）
RUN npm run build:prod

# 生产阶段
FROM nginx:1.25-alpine

# 安装必要工具
RUN apk add --no-cache curl

# 复制构建产物
COPY --from=builder /app/dist /usr/share/nginx/html

# 复制nginx配置
COPY nginx.conf /etc/nginx/conf.d/default.conf

# 创建nginx用户（如果不存在）
RUN (addgroup -g 1001 -S nginx 2>/dev/null || true) && \
    (adduser -S -D -H -u 1001 -h /var/cache/nginx -s /sbin/nologin -G nginx -g nginx nginx 2>/dev/null || true)

# 设置权限
RUN chown -R nginx:nginx /usr/share/nginx/html && \
    chown -R nginx:nginx /var/cache/nginx && \
    chown -R nginx:nginx /var/log/nginx && \
    chown -R nginx:nginx /etc/nginx/conf.d && \
    touch /var/run/nginx.pid && \
    chown nginx:nginx /var/run/nginx.pid

# 暴露端口
EXPOSE 80

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost/ || exit 1

# 使用非root用户
USER nginx

# 启动nginx
CMD ["nginx", "-g", "daemon off;"]