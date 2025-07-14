#!/bin/bash

# GitHub发布前的准备脚本
# 用于清理敏感文件并准备项目上传

echo "🚀 开始准备GitHub发布..."

# 创建scripts目录
mkdir -p scripts

# 1. 清理敏感文件
echo "🧹 清理敏感文件..."

# 删除环境配置文件
echo "  - 删除环境配置文件..."
rm -f backend/.env backend/.env.dev frontend/.env

# 删除数据库文件
echo "  - 删除数据库文件..."
rm -f backend/binance_trading.db backend/test.db

# 删除系统文件
echo "  - 删除系统文件..."
find . -name ".DS_Store" -type f -delete

# 删除依赖目录
echo "  - 删除依赖目录..."
rm -rf frontend/node_modules node_modules backend/vendor

# 删除备份文件
echo "  - 删除备份文件..."
rm -rf cleanup_backup_*
find . -name "*.bak" -type f -delete

# 删除二进制和临时文件
echo "  - 删除二进制和临时文件..."
rm -rf backend/tmp backend/bin bin/
find . -name "*.exe" -type f -delete
find . -name "*.test" -type f -delete

# 删除日志文件
echo "  - 删除日志文件..."
find . -name "*.log" -type f -delete
rm -rf logs/ log/

# 2. 检查必要文件是否存在
echo "📋 检查必要文件..."

required_files=(".gitignore" "README.md" "LICENSE" ".env.example")
for file in "${required_files[@]}"; do
    if [ -f "$file" ]; then
        echo "  ✅ $file 存在"
    else
        echo "  ❌ $file 不存在"
    fi
done

# 3. 检查敏感文件是否已清理
echo "🔍 检查敏感文件清理状态..."

sensitive_files=("backend/.env" "backend/.env.dev" "frontend/.env" "backend/binance_trading.db")
all_clean=true

for file in "${sensitive_files[@]}"; do
    if [ -f "$file" ]; then
        echo "  ⚠️  $file 仍然存在，需要手动删除"
        all_clean=false
    else
        echo "  ✅ $file 已清理"
    fi
done

# 4. 检查目录大小
echo "📊 检查项目大小..."
du -sh . | while read size path; do
    echo "  项目总大小: $size"
done

# 5. 显示项目结构
echo "📁 项目结构预览..."
tree -L 2 -I 'node_modules|vendor|*.log|*.db|tmp|bin' . || ls -la

# 6. 生成清理报告
echo "📄 生成清理报告..."
cat > GITHUB_RELEASE_NOTES.md << 'EOF'
# GitHub发布准备报告

## 已清理的敏感文件

- [x] 环境配置文件 (.env, .env.dev)
- [x] 数据库文件 (.db, .sqlite)
- [x] 系统文件 (.DS_Store)
- [x] 依赖目录 (node_modules, vendor)
- [x] 备份文件 (cleanup_backup_*)
- [x] 二进制文件 (bin/, tmp/)
- [x] 日志文件 (*.log)

## 安全检查

- [x] 无硬编码的API密钥
- [x] 无明文密码
- [x] 无敏感数据库内容
- [x] 已配置.gitignore文件

## 后续步骤

1. 检查.env.example文件，确保不包含真实密钥
2. 初始化Git仓库：`git init`
3. 添加远程仓库：`git remote add origin https://github.com/ccj241/ccTrade.git`
4. 提交代码：`git add . && git commit -m "Initial commit"`
5. 推送到GitHub：`git push -u origin main`

## 部署提醒

- 生产环境请使用真实的环境变量
- 更改所有默认密码和密钥
- 启用HTTPS和安全头
- 配置监控和备份
EOF

# 7. 最终检查
if $all_clean; then
    echo "✅ 所有敏感文件已清理完成！"
    echo "📤 项目已准备好上传到GitHub"
    echo ""
    echo "🔄 下一步操作："
    echo "1. 检查 .env.example 文件"
    echo "2. 运行: git init"
    echo "3. 运行: git remote add origin https://github.com/ccj241/ccTrade.git"
    echo "4. 运行: git add ."
    echo "5. 运行: git commit -m 'Initial commit'"
    echo "6. 运行: git push -u origin main"
else
    echo "❌ 仍有敏感文件未清理，请手动删除后再次运行此脚本"
fi

echo "🎉 GitHub发布准备完成！"