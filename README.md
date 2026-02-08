# cangjie-mem

> 仓颉语言分级记忆库 - 支持智能检索与 Web 管理界面

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?logo=go)](https://golang.org/)

**cangjie-mem** 是一个专用于仓颉编程语言的智能记忆管理系统，支持多级别（语言/项目/公共库）知识存储、智能检索，并提供现代化的 Web 管理界面。

## ✨ 核心特性

### 🗂️ 分级记忆模型

- **语言级**：语法规范、关键字、核心语义
- **公共库级**：工具函数、设计模式、最佳实践
- **项目级**：项目配置、业务逻辑、开发约定

### 🔍 智能检索

- **自动层级判断**：根据查询内容智能选择最佳记忆层级
- **全文搜索**：基于 SQLite FTS5 的高效全文检索（AND 匹配）
- **置信度评分**：基于匹配度、来源可信度、访问热度排序

### 🌐 Web 管理界面

- **可视化浏览**：现代化 UI，按层级/库/项目分组浏览
- **CRUD 操作**：创建、编辑、删除记忆的完整支持
- **搜索筛选**：实时搜索、多维度筛选（层级/库/项目）
- **导入/导出**：支持知识包的 JSON 导入导出

## 🚀 快速开始

### Docker 部署（推荐）

```bash
docker run -d \
  --name cangjie-mem \
  -p 8080:8080 \
  -v cangjie-data:/home/cangjie/.cangjie-mem \
  ghcr.io/ystyle/cangjie-mem:latest

# 访问 Web UI
open http://localhost:8080
```

### Docker Compose

```yaml
services:
  cangjie-mem:
    image: ghcr.io/ystyle/cangjie-mem:latest
    ports:
      - "8080:8080"
    volumes:
      - cangjie-data:/home/cangjie/.cangjie-mem

volumes:
  cangjie-data:
```

### 二进制部署

访问 [GitHub Releases](https://github.com/ystyle/cangjie-mem/releases) 下载对应平台的最新版本二进制文件。

```bash
# 下载最新版本（自动跳转到最新版本）
wget https://github.com/ystyle/cangjie-mem/releases/latest/download/cangjie-mem-linux-amd64.tar.gz

# 或指定版本（如 v1.5.0）
wget https://github.com/ystyle/cangjie-mem/releases/download/v1.5.0/cangjie-mem-linux-amd64.tar.gz

# 解压
tar xzf cangjie-mem-linux-amd64.tar.gz

# 启动服务（启用 Web UI）
./cangjie-mem -http -api -ui

# 访问 Web UI
open http://localhost:8080
```

## 📖 使用方法

### Web 界面使用

访问 `http://localhost:8080` 后，你可以：

1. **浏览记忆**：按层级、库、项目分组浏览所有记忆
2. **搜索记忆**：使用搜索框实时检索标题和内容
3. **创建记忆**：点击"新建记忆"添加新内容
4. **编辑/删除**：点击记忆卡片进行编辑或删除
5. **导入/导出**：批量导入导出 JSON 格式的知识包

### MCP 集成（Claude Code）

在 Claude Code 配置中添加：

**stdio 模式**（本地）：
```json
{
  "mcpServers": {
    "cangjie-mem": {
      "command": "/path/to/cangjie-mem",
      "env": {
        "CANGJIE_DB_PATH": "/path/to/.cangjie-mem/memory.db"
      }
    }
  }
}
```

**HTTP 模式**（远程）：
```json
{
  "mcpServers": {
    "cangjie-mem": {
      "transport": "http",
      "url": "http://your-server:8080/mcp"
    }
  }
}
```

### MCP 工具

| 工具 | 说明 | 参数 |
|-----|------|------|
| `cangjie_mem_store` | 存储记忆 | level, title, content, library_name?, project_path_pattern? |
| `cangjie_mem_recall` | 检索记忆（核心） | query（空格分隔关键词）, level?, max_results? |
| `cangjie_mem_list` | 列出记忆 | level?, library_name?, brief?, limit?, offset? |
| `cangjie_mem_list_categories` | 列出分类 | 无 |
| `cangjie_mem_delete` | 删除记忆 | id |

### 使用示例

```
# 存储记忆
仓颉语言中接口定义使用 'interface' 关键字

# 检索记忆
Tang 框架如何处理路由中间件？

# 列出特定库的记忆
列出所有 tang 库相关的记忆
```

## 🔗 最佳实践

查看[最佳实践文档](https://github.com/ystyle/cangjie-mem/blob/master/best-practices.md) 理解使用方法

## 🛠️ 开发

### 从源码构建

```bash
# 克隆仓库
git clone https://github.com/ystyle/cangjie-mem.git
cd cangjie-mem

# 安装前端依赖
cd web && pnpm install && cd ..

# 构建前端
cd web && pnpm build && cd ..

# 构建 Go 二进制
go build -tags="sqlite_fts5" -o cangjie-mem ./cmd/server

# 运行
./cangjie-mem -http -api -ui
```

### 本地开发

**终端 1 - 启动 Go API**：
```bash
go run -tags="sqlite_fts5" ./cmd/server -http -api -addr :8080
```

**终端 2 - 启动前端开发服务器**：
```bash
cd web && pnpm dev
```

访问 http://localhost:5173

详细开发指南请查看 [DEVELOPMENT.md](DEVELOPMENT.md)

## 🏗️ 项目结构

```
cangjie-mem/
├── cmd/server/       # 主入口
├── pkg/
│   ├── db/           # 数据库层（SQLite + FTS5）
│   ├── mcp/          # MCP 服务器
│   └── types/        # 类型定义
├── internal/
│   ├── api/          # REST API 处理器
│   ├── config/       # 配置管理
│   └── store/        # 智能检索逻辑
├── web/              # Vue 3 前端
│   ├── src/
│   │   ├── api/      # API 客户端
│   │   ├── components/  # Vue 组件
│   │   ├── stores/   # Pinia 状态管理
│   │   └── views/    # 页面组件
│   └── package.json
├── Dockerfile        # 多阶段构建（前端 + Go）
└── README.md
```

## 🎯 技术栈

- **后端**：Go 1.23+, SQLite (FTS5)
- **前端**：Vue 3, Vite, Naive UI, Pinia, TypeScript
- **协议**：Model Context Protocol (MCP)
- **部署**：Docker (多平台镜像)

## 📋 环境变量

| 变量 | 说明 | 默认值 |
|-----|------|--------|
| `CANGJIE_HTTP` | 启用 HTTP 模式 | `false` |
| `CANGJIE_ADDR` | HTTP 监听地址 | `:8080` |
| `CANGJIE_API_ENABLED` | 启用 REST API | `false` |
| `CANGJIE_UI_ENABLED` | 启用 Web UI | `false` |
| `CANGJIE_TOKEN` | MCP 认证 Token | 空 |
| `CANGJIE_API_BASIC_AUTH_USERNAME` | API Basic Auth 用户名 | 空 |
| `CANGJIE_API_BASIC_AUTH_PASSWORD` | API Basic Auth 密码 | 空 |

## 🔒 API 认证配置

`/api/*` REST API 端点支持独立的 Basic Auth 认证，与 MCP 的 Token 认证分离。

### 配置方法

**Docker Compose**：
```yaml
environment:
  - CANGJIE_API_BASIC_AUTH_USERNAME=admin
  - CANGJIE_API_BASIC_AUTH_PASSWORD=your-secret-password
```

**Docker Run**：
```bash
docker run -d \
  -e CANGJIE_API_BASIC_AUTH_USERNAME=admin \
  -e CANGJIE_API_BASIC_AUTH_PASSWORD=your-secret-password \
  ghcr.io/ystyle/cangjie-mem:latest
```

**前端配置**（如果使用独立的前端开发服务器）：
```bash
# web/.env.local
VITE_API_USERNAME=admin
VITE_API_PASSWORD=your-secret-password
```

### 安全说明

- **本地开发**：无需配置认证，直接访问即可
- **内网部署**：建议配置 Basic Auth
- **公网部署**：必须配置认证，并使用 HTTPS

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件

## 🙏 致谢

- [cangjie-docs-mcp](https://github.com/ystyle/cangjie-docs-mcp) - 仓颉语言文档检索系统
- [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go) - MCP Go SDK

---

**Made with ❤️ by ystyle**
