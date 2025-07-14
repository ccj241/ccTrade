# 贡献指南

感谢您对ccTrade项目的贡献！本文档将指导您如何参与项目开发。

## 🚀 开始贡献

### 前置条件

- **Go**: 1.23+
- **Node.js**: 18+
- **MySQL**: 8.0+
- **Redis**: 7.0+
- **Docker**: 20.10+ (可选)
- **Git**: 2.0+

### 开发环境设置

1. **Fork仓库**
   ```bash
   # 在GitHub上Fork项目
   # 然后克隆你的Fork
   git clone https://github.com/你的用户名/ccTrade.git
   cd ccTrade
   ```

2. **配置环境**
   ```bash
   # 复制环境配置文件
   cp .env.example .env
   # 编辑.env文件，配置数据库等信息
   ```

3. **安装依赖**
   ```bash
   # 后端依赖
   cd backend
   go mod download
   
   # 前端依赖
   cd ../frontend
   npm install
   ```

4. **启动开发环境**
   ```bash
   # 使用Docker Compose
   make docker-dev
   
   # 或者手动启动
   # 后端
   cd backend && go run main.go
   # 前端
   cd frontend && npm run dev
   ```

## 🔄 贡献流程

### 1. 创建Issue

在开始工作之前，请先创建一个Issue来讨论你的想法：

- 🐛 **Bug报告**: 使用Bug报告模板
- 💡 **功能建议**: 使用功能请求模板
- 📚 **文档改进**: 直接创建Issue说明

### 2. 分支策略

```bash
# 创建特性分支
git checkout -b feature/功能名称

# 创建修复分支
git checkout -b fix/问题描述

# 创建文档分支
git checkout -b docs/文档更新
```

### 3. 开发规范

#### 后端开发 (Go)

- **代码风格**: 遵循Go官方代码规范
- **命名约定**: 使用驼峰命名法
- **错误处理**: 始终处理错误，不要忽略
- **测试**: 为新功能编写测试
- **注释**: 为公共函数和复杂逻辑添加注释

```go
// 良好的示例
func CreateUser(ctx context.Context, user *models.User) error {
    if err := validateUser(user); err != nil {
        return fmt.Errorf("用户验证失败: %w", err)
    }
    
    // 业务逻辑
    return nil
}
```

#### 前端开发 (React + TypeScript)

- **代码风格**: 使用ESLint和Prettier
- **组件命名**: 使用PascalCase
- **文件命名**: 使用kebab-case
- **类型定义**: 为所有组件和函数添加类型
- **Hooks**: 优先使用函数组件和Hooks

```typescript
// 良好的示例
interface UserProps {
  user: User;
  onUpdate: (user: User) => void;
}

const UserProfile: React.FC<UserProps> = ({ user, onUpdate }) => {
  const [editing, setEditing] = useState(false);
  
  // 组件逻辑
  return <div>...</div>;
};
```

### 4. 测试要求

#### 后端测试

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./controllers

# 运行覆盖率测试
go test -cover ./...
```

#### 前端测试

```bash
# 运行单元测试
npm test

# 运行E2E测试
npm run test:e2e

# 运行覆盖率测试
npm run test:coverage
```

### 5. 提交规范

使用约定式提交格式：

```bash
# 功能
git commit -m "feat: 添加用户管理功能"

# 修复
git commit -m "fix: 修复登录验证问题"

# 文档
git commit -m "docs: 更新API文档"

# 样式
git commit -m "style: 格式化代码"

# 重构
git commit -m "refactor: 重构用户服务"

# 测试
git commit -m "test: 添加用户测试用例"

# 构建
git commit -m "build: 更新依赖版本"
```

### 6. Pull Request

1. **推送分支**
   ```bash
   git push origin feature/功能名称
   ```

2. **创建PR**
   - 在GitHub上创建Pull Request
   - 使用描述性的标题
   - 详细描述变更内容
   - 关联相关Issue

3. **PR模板**
   ```markdown
   ## 变更描述
   简要描述这个PR的内容

   ## 变更类型
   - [ ] 功能(feature)
   - [ ] 修复(fix)
   - [ ] 文档(docs)
   - [ ] 样式(style)
   - [ ] 重构(refactor)
   - [ ] 测试(test)

   ## 测试
   - [ ] 单元测试通过
   - [ ] 集成测试通过
   - [ ] 手动测试通过

   ## 相关Issue
   关闭 #issue_number
   ```

## 🔍 代码审查

### 审查标准

- **功能正确性**: 代码实现符合需求
- **代码质量**: 遵循最佳实践
- **测试覆盖**: 有足够的测试用例
- **文档完整**: 代码有适当的注释
- **性能考虑**: 没有明显的性能问题
- **安全性**: 没有安全漏洞

### 审查流程

1. **自动检查**: CI/CD流水线自动运行
2. **代码审查**: 至少一个维护者审查
3. **测试验证**: 确保所有测试通过
4. **合并代码**: 审查通过后合并

## 📝 文档规范

### API文档

- 使用OpenAPI/Swagger规范
- 包含请求/响应示例
- 说明错误码和错误信息

### 代码注释

```go
// UserService 用户服务接口
type UserService interface {
    // CreateUser 创建新用户
    // 参数:
    //   ctx: 上下文
    //   user: 用户信息
    // 返回:
    //   error: 错误信息，nil表示成功
    CreateUser(ctx context.Context, user *models.User) error
}
```

### README更新

- 新功能添加后更新README
- 保持安装和使用说明的准确性
- 更新API文档链接

## 🐛 Bug报告

### 报告格式

```markdown
## Bug描述
简要描述问题

## 复现步骤
1. 步骤1
2. 步骤2
3. 步骤3

## 预期行为
描述预期的正确行为

## 实际行为
描述实际发生的错误行为

## 环境信息
- 操作系统: [e.g., macOS 14.0]
- 浏览器: [e.g., Chrome 118]
- 项目版本: [e.g., v2.0.0]

## 附加信息
- 错误日志
- 截图
- 其他相关信息
```

## 💡 功能建议

### 建议格式

```markdown
## 功能描述
简要描述建议的功能

## 使用场景
描述这个功能的使用场景

## 实现思路
简要描述可能的实现方式

## 其他考虑
- 性能影响
- 兼容性
- 安全性
```

## 🎯 发版流程

### 版本号规则

使用语义化版本：`主版本.次版本.修订版本`

- **主版本**: 不兼容的API修改
- **次版本**: 向后兼容的功能新增
- **修订版本**: 向后兼容的问题修正

### 发版检查清单

- [ ] 所有测试通过
- [ ] 文档更新完成
- [ ] 版本号更新
- [ ] 变更日志更新
- [ ] 安全审查通过
- [ ] 性能测试通过

## 📞 联系方式

如果您有任何问题或建议，请通过以下方式联系我们：

- 📧 **邮箱**: dev@cctrade.com
- 💬 **讨论**: [GitHub Discussions](https://github.com/ccj241/ccTrade/discussions)
- 🐛 **Issue**: [GitHub Issues](https://github.com/ccj241/ccTrade/issues)

## 🙏 感谢

感谢所有为ccTrade项目做出贡献的开发者！

---

再次感谢您的贡献！让我们一起打造更好的交易系统！ 🚀