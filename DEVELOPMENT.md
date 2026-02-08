# 开发指南

本文档面向 cangjie-mem 的维护者，记录开发、测试和发版流程。

## 项目架构

cangjie-mem 采用单体架构，一个 Go 进程同时服务：
- **MCP 协议**（stdio 或 HTTP）
- **REST API**
- **Web UI**（嵌入的静态文件）

```
┌─────────────────────────────────────────────┐
│              cangjie-mem 进程               │
│                                             │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  │
│  │   MCP    │  │   API    │  │  Web UI  │  │
│  │  Server  │  │  Handler │  │  Static  │  │
│  └──────────┘  └──────────┘  └──────────┘  │
│                                             │
│              ┌──────────────┐               │
│              │   SQLite DB  │               │
│              └──────────────┘               │
└─────────────────────────────────────────────┘
```

## 开发环境

### 前置要求

- **Go**: 1.23 或更高版本
- **Node.js**: 20 或更高版本（前端开发）
- **pnpm**: 8 或更高版本（前端包管理）
- **SQLite**: 支持 FTS5 扩展（系统自带或通过 CGo 编译）
- **Docker**: 用于容器化部署和测试

### 依赖安装

```bash
# 克隆仓库
git clone https://github.com/ystyle/cangjie-mem.git
cd cangjie-mem

# 下载 Go 依赖
go mod download

# 安装前端依赖
cd web
pnpm install
cd ..
```

## 本地开发

### 方式 1：同时运行前端开发服务器和 Go API

**推荐用于前端开发**，支持热更新。

**终端 1 - 启动 Go API 服务器**：
```bash
go run -tags="sqlite_fts5" ./cmd/server -http -api -addr :8080
```

**终端 2 - 启动前端开发服务器**：
```bash
cd web
pnpm dev
```

然后访问 http://localhost:5173

### 方式 2：构建并运行完整应用

**推荐用于后端开发和测试**。

```bash
# 构建前端
cd web
pnpm build
cd ..

# 运行 Go 服务器（会嵌入前端构建产物）
go run -tags="sqlite_fts5" ./cmd/server -http -api -ui -addr :8080
```

然后访问 http://localhost:8080

### 环境变量

| 变量 | 说明 | 默认值 |
|-----|------|--------|
| `CANGJIE_HTTP` | 启用 HTTP 模式 | `false` |
| `CANGJIE_ADDR` | HTTP 监听地址 | `:8080` |
| `CANGJIE_API_ENABLED` | 启用 REST API | `false` |
| `CANGJIE_UI_ENABLED` | 启用 Web UI | `false` |
| `CANGJIE_TOKEN` | MCP 认证 Token | 空 |
| `CANGJIE_API_BASIC_AUTH_USERNAME` | API Basic Auth 用户名 | 空 |
| `CANGJIE_API_BASIC_AUTH_PASSWORD` | API Basic Auth 密码 | 空 |

### API 认证

`/api/*` REST API 端点使用独立的 Basic Auth 认证，与 MCP 的 Token 认证完全分离：

**认证架构**：
- `/mcp` → 使用 `CANGJIE_TOKEN`（MCP 协议）
- `/api/*` → 使用 `CANGJIE_API_BASIC_AUTH_USERNAME/PASSWORD`（Basic Auth）
- `/` (Web UI) → 静态文件，无认证

**本地开发**：
- 默认无需认证，所有端点开放
- Web UI 直接访问即可

**测试认证**：
```bash
# 设置环境变量
export CANGJIE_API_BASIC_AUTH_USERNAME=admin
export CANGJIE_API_BASIC_AUTH_PASSWORD=test123

# 测试 API（应该返回 401）
curl http://localhost:8080/api/memories

# 使用 Basic Auth（应该成功）
curl -u admin:test123 http://localhost:8080/api/memories
```

## 前端开发

### 目录结构

```
web/
├── src/
│   ├── api/          # API 客户端
│   ├── components/   # Vue 组件
│   ├── stores/       # Pinia 状态管理
│   ├── types/        # TypeScript 类型定义
│   ├── views/        # 页面组件
│   ├── App.vue       # 根组件
│   └── main.ts       # 入口文件
├── public/           # 静态资源
├── index.html        # HTML 模板
├── vite.config.ts    # Vite 配置
├── tsconfig.json     # TypeScript 配置
├── package.json      # 依赖配置
└── pnpm-lock.yaml    # 锁文件
```

### 常用命令

```bash
cd web

# 开发模式（热更新）
pnpm dev

# 构建生产版本
pnpm build

# 预览构建结果
pnpm preview

# 类型检查
pnpm type-check

# 代码检查
pnpm lint
```

### 技术栈

- **框架**: Vue 3 (Composition API)
- **构建工具**: Vite
- **UI 组件**: Naive UI
- **状态管理**: Pinia
- **路由**: Vue Router
- **语言**: TypeScript
- **图标**: @vicons/material

## 测试

### 运行测试

```bash
# 运行所有测试
go test ./... -tags="sqlite_fts5"

# 运行特定包的测试
go test ./pkg/db -tags="sqlite_fts5"

# 运行测试并显示详细输出
go test ./... -tags="sqlite_fts5" -v

# 查看测试覆盖率
go test ./... -tags="sqlite_fts5" -cover
```

### 前端测试

```bash
cd web

# 运行单元测试（如果有）
pnpm test

# 运行端到端测试（如果有）
pnpm test:e2e
```

## 构建和部署

### 本地构建

**完整构建（包含前端）**：
```bash
# 1. 构建前端
cd web && pnpm build && cd ..

# 2. 构建 Go 二进制（会嵌入前端产物）
go build -tags="sqlite_fts5" -o cangjie-mem ./cmd/server

# 3. 运行
./cangjie-mem -http -api -ui
```

**仅 Go 构建（使用已构建的前端）**：
```bash
# 前端已构建在 web/dist/ 目录
go build -tags="sqlite_fts5" -o cangjie-mem ./cmd/server
```

### Docker 构建

```bash
# 构建本地镜像
docker build -t cangjie-mem:local .

# 构建多平台镜像（需要 buildx）
docker buildx build --platform linux/amd64,linux/arm64 -t cangjie-mem:local .

# 运行容器
docker run -d --name cangjie-mem -p 8080:8080 cangjie-mem:local
```

### 生产部署

**使用 Docker Compose**：
```yaml
version: '3.8'

services:
  cangjie-mem:
    image: ghcr.io/ystyle/cangjie-mem:latest
    container_name: cangjie-mem
    restart: unless-stopped
    ports:
      - "8080:8080"
    environment:
      - CANGJIE_HTTP=true
      - CANGJIE_API_ENABLED=true
      - CANGJIE_UI_ENABLED=true
    volumes:
      - cangjie-data:/home/cangjie/.cangjie-mem

volumes:
  cangjie-data:
```

## 发版流程

### 版本号规则

遵循语义化版本（Semantic Versioning）：

- **主版本号**（如 1.x.x）：破坏性变更，不兼容的 API 修改
- **次版本号**（如 x.1.x）：新功能，向后兼容
- **修订号**（如 x.x.1）：Bug 修复，向后兼容

### 发版步骤

#### 1. 准备工作

确保所有测试通过：
```bash
go test ./... -tags="sqlite_fts5"
cd web && pnpm build && cd ..
```

#### 2. 更新版本信息

更新以下文件：
- `pkg/version/version.go`: 修改 `Version` 变量
- `web/package.json`: 修改版本号
- README.md 和 DEVELOPMENT.md: 更新功能说明

#### 3. 提交代码

```bash
git add .
git commit -m "<type>: <description>"
```

#### 4. 打 Tag

```bash
git tag v<version> -a -m "v<version>: <description>"
git push
git push --tags
```

#### 5. GitHub Actions 自动构建

推送 tag 后，GitHub Actions 会自动：
1. 构建前端（pnpm build）
2. 构建多平台 Go 二进制文件
3. 构建 Docker 镜像（包含嵌入的前端）
4. 创建 GitHub Release
5. 上传所有构建产物

## 常见问题

### 前端构建失败

**错误**：`Cannot find module 'xxx'`

**解决**：
```bash
cd web
rm -rf node_modules pnpm-lock.yaml
pnpm install
```

### Go 构建找不到前端文件

**错误**：`no such file or directory: web/dist`

**解决**：
```bash
cd web && pnpm build && cd ..
```

### Docker 构建失败

**错误**：前端构建阶段失败

**解决**：确保 `web/package.json` 和 `web/pnpm-lock.yaml` 存在

### 热更新不工作

**检查**：
1. Go API 服务器运行在 http://localhost:8080
2. 前端开发服务器运行在 http://localhost:5173
3. 浏览器访问的是 http://localhost:5173（不是 8080）

## 参考资源

- [Go 官方文档](https://golang.org/doc/)
- [Vue 3 文档](https://vuejs.org/)
- [Naive UI 文档](https://www.naiveui.com/)
- [Vite 文档](https://vitejs.dev/)
- [SQLite FTS5 文档](https://www.sqlite.org/fts5.html)
- [Model Context Protocol](https://modelcontextprotocol.io/)
