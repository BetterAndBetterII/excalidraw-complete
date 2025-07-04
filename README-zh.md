# Excalidraw Complete - 增强版

基于原有的 Excalidraw Complete 自托管解决方案，新增了以下功能：

## 🆕 新增功能

### 1. 用户管理系统
- **GitHub OAuth 登录**：支持使用 GitHub 账户安全登录
- **用户会话管理**：安全的会话管理和自动过期清理
- **多用户支持**：每个用户都有独立的数据空间

### 2. 画布管理
- **保存画布到后端**：画布数据持久化存储
- **切换画布**：轻松在多个画布之间切换
- **画布列表**：查看和管理所有个人画布
- **公开/私有设置**：控制画布的可见性

### 3. 现代化前端界面
- **响应式设计**：适配桌面和移动设备
- **美观的用户界面**：现代化的设计风格
- **流畅的用户体验**：直观的操作流程

## 🚀 快速开始

### 环境要求
- Go 1.21 或更高版本
- SQLite（用于数据存储）
- GitHub OAuth 应用（用于用户认证）

### 1. 创建 GitHub OAuth 应用

1. 访问 [GitHub Developer Settings](https://github.com/settings/developers)
2. 点击 "New OAuth App"
3. 填写应用信息：
   - **Application name**: `Excalidraw Complete`
   - **Homepage URL**: `http://localhost:3002`
   - **Authorization callback URL**: `http://localhost:3002/auth/github/callback`
4. 创建应用后，记录 Client ID 和 Client Secret

### 2. 环境配置

创建 `.env` 文件（或设置环境变量）：

```bash
# 存储配置
STORAGE_TYPE=sqlite
DATA_SOURCE_NAME=./excalidraw.db

# GitHub OAuth 配置
GITHUB_CLIENT_ID=your_github_client_id
GITHUB_CLIENT_SECRET=your_github_client_secret

# 应用配置
BASE_URL=http://localhost:3002
```

### 3. 运行应用

```bash
# 克隆仓库
git clone https://github.com/PatWie/excalidraw-complete.git --recursive
cd excalidraw-complete

# 构建前端（如果需要）
cd excalidraw
git apply ../frontend.patch
npm install
cd excalidraw-app
npm run build:app:docker
cd ../../
cp -r excalidraw/excalidraw-app/build frontend/

# 运行应用
STORAGE_TYPE=sqlite DATA_SOURCE_NAME=./excalidraw.db GITHUB_CLIENT_ID=your_client_id GITHUB_CLIENT_SECRET=your_client_secret go run main.go --loglevel debug
```

### 4. 访问应用

打开浏览器访问 `http://localhost:3002`

## 🔧 配置选项

### 存储配置

支持多种存储方式：

```bash
# SQLite（推荐用于单机部署）
STORAGE_TYPE=sqlite
DATA_SOURCE_NAME=./excalidraw.db

# 文件系统
STORAGE_TYPE=filesystem
LOCAL_STORAGE_PATH=/tmp/excalidraw/

# AWS S3
STORAGE_TYPE=s3
S3_BUCKET_NAME=your-bucket-name

# 内存存储（仅用于测试）
STORAGE_TYPE=memory
```

### GitHub OAuth 配置

```bash
# 必需的 OAuth 配置
GITHUB_CLIENT_ID=your_github_client_id
GITHUB_CLIENT_SECRET=your_github_client_secret

# 可选的基础URL配置（默认为 http://localhost:3002）
BASE_URL=https://your-domain.com
```

## 📚 API 文档

### 认证端点

- `GET /auth/github` - 启动 GitHub OAuth 登录
- `GET /auth/github/callback` - OAuth 回调处理
- `POST /auth/logout` - 用户登出
- `GET /auth/me` - 获取当前用户信息

### 画布管理端点

- `GET /api/canvas/` - 获取用户画布列表
- `POST /api/canvas/` - 创建新画布
- `GET /api/canvas/{id}` - 获取特定画布
- `PUT /api/canvas/{id}` - 更新画布
- `DELETE /api/canvas/{id}` - 删除画布

### 兼容性端点

保留原有的 Firebase 兼容性端点和文档端点，确保向后兼容。

## 🛠️ 开发指南

### 项目结构

```
excalidraw-complete/
├── main.go                      # 主应用程序
├── core/
│   └── entity.go               # 核心实体定义
├── stores/
│   ├── storage.go              # 存储抽象层
│   ├── sqlite/                 # SQLite 实现
│   ├── memory/                 # 内存存储实现
│   └── ...
├── handlers/
│   ├── auth/                   # 认证处理器
│   │   ├── oauth.go
│   │   └── middleware.go
│   └── api/
│       ├── canvas/             # 画布API
│       └── documents/          # 文档API
└── frontend/                   # 前端文件
    ├── index.html
    ├── styles.css
    └── app.js
```

### 添加新功能

1. **添加新的存储接口**：在 `core/entity.go` 中定义
2. **实现存储**：在相应的存储目录中实现接口
3. **添加API端点**：在 `handlers/api/` 中创建新的处理器
4. **更新路由**：在 `main.go` 中添加路由

## 🔒 安全特性

- **安全的会话管理**：使用安全的 Cookie 和会话过期
- **CSRF 保护**：内置的 CSRF 保护机制
- **数据隔离**：每个用户只能访问自己的数据
- **OAuth 安全**：使用 GitHub OAuth 进行安全认证

## 📦 部署

### Docker 部署

```dockerfile
# 使用现有的 Dockerfile
docker build -t excalidraw-complete -f excalidraw-complete.Dockerfile .

# 运行容器
docker run -d \
  -p 3002:3002 \
  -e STORAGE_TYPE=sqlite \
  -e DATA_SOURCE_NAME=/app/data/excalidraw.db \
  -e GITHUB_CLIENT_ID=your_client_id \
  -e GITHUB_CLIENT_SECRET=your_client_secret \
  -v /path/to/data:/app/data \
  excalidraw-complete
```

### 生产环境建议

1. **使用 HTTPS**：配置 SSL 证书
2. **设置环境变量**：不要在代码中硬编码敏感信息
3. **定期备份**：备份数据库文件
4. **监控日志**：使用适当的日志级别

## 🤝 贡献

欢迎贡献代码！请遵循以下步骤：

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

## 📄 许可证

本项目基于原有的 Excalidraw Complete 项目，保持相同的许可证。

## 🙏 致谢

- 感谢 [Excalidraw](https://excalidraw.com) 项目
- 感谢原始的 [Excalidraw Complete](https://github.com/PatWie/excalidraw-complete) 项目
- 感谢所有贡献者