# cangjie-mem

> 仓颉语言分级记忆库 MCP 服务器

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://golang.org/)

**cangjie-mem** 是一个专用于仓颉编程语言的、支持多级别（语言/项目/公共库）知识智能管理与检索的 MCP 服务器。它解决了通用 AI 工具在识别和运用新语言语法时的知识缺失与上下文遗忘问题。

## 🎯 核心特性

### 分级记忆模型

- **语言级（language）**：权威规范，包括语法、关键字、核心语义等
- **公共库级（library）**：可复用方案，包括工具函数、设计模式、最佳实践等
- **项目级（project）**：具体上下文，包括项目结构、配置、业务逻辑等

### 智能检索

- ✅ **自动层级判断**：根据查询内容和项目上下文智能选择最佳记忆层级
- ✅ **置信度评分**：基于匹配度、来源可信度、访问热度计算相关性
- ✅ **全文搜索**：基于 SQLite FTS5 的高效全文检索
- ✅ **上下文感知**：结合项目路径进行精准匹配

## 📦 与 cangjie-docs-mcp 的区别

| 特性 | cangjie-docs-mcp | cangjie-mem |
|------|-----------------|-------------|
| **定位** | 官方文档搜索 | 实践经验记忆库 |
| **内容** | 公开的、标准的、权威的 | 个人的、实践的、演进的 |
| **类比** | 教科书/参考手册 | 笔记本/经验库 |
| **更新** | 随官方文档更新 | 持续积累和演进 |

**两者互补，协同使用！** 🎯

## 🚀 快速开始

cangjie-mem 支持两种启动模式，根据使用场景选择：

---

### 模式 1：stdio 模式（本地使用）

**适用场景**：个人开发，本地使用

#### 1. 下载预编译二进制

访问 [GitHub Releases](https://github.com/ystyle/cangjie-mem/releases) 下载对应平台的二进制文件。

**Linux**：
```bash
# 下载（以 Linux AMD64 为例）
wget https://github.com/ystyle/cangjie-mem/releases/download/v1.3.0/cangjie-mem-linux-amd64.tar.gz

# 解压
tar xzf cangjie-mem-linux-amd64.tar.gz

# 运行（测试）
./cangjie-mem-linux-amd64 -version
```

**Windows**：
```powershell
# 下载（使用 PowerShell）
Invoke-WebRequest -Uri "https://github.com/ystyle/cangjie-mem/releases/download/v1.3.0/cangjie-mem-windows-amd64.tar.gz" -OutFile "cangjie-mem-windows-amd64.tar.gz"

# 解压（需要 tar 工具，Windows 10+ 内置）
tar xzf cangjie-mem-windows-amd64.tar.gz

# 运行（测试）
.\cangjie-mem-windows-amd64.exe -version
```

**macOS**：
```bash
# 下载（Apple Silicon M1/M2/M3）
curl -LO https://github.com/ystyle/cangjie-mem/releases/download/v1.3.0/cangjie-mem-darwin-arm64.tar.gz

# 或 Intel Mac
# curl -LO https://github.com/ystyle/cangjie-mem/releases/download/v1.3.0/cangjie-mem-darwin-amd64.tar.gz

# 解压
tar xzf cangjie-mem-darwin-arm64.tar.gz

# 运行（测试）
./cangjie-mem-darwin-arm64 -version
```

#### 2. 配置 Claude Code

在 Claude Code 的配置文件中添加：

**Linux**: `~/.config/Claude/claude_desktop_config.json`
**macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
**Windows**: `%APPDATA%/Claude/claude_desktop_config.json`

```json
{
  "mcpServers": {
    "cangjie-mem": {
      "command": "/path/to/cangjie-mem-linux-amd64",
      "env": {
        "CANGJIE_DB_PATH": "/path/to/.cangjie-mem/memory.db"
      }
    }
  }
}
```

> **提示**：将 `/path/to/cangjie-mem-linux-amd64` 替换为实际的二进制文件路径

配置完成后，重启 Claude Code 即可开始使用！

---

### 模式 2：HTTP 模式（远程/服务器部署）

**适用场景**：团队协作、多设备共享、远程访问

#### 方式 1：Docker Compose（推荐）

```bash
# 创建 docker-compose.yml
cat > docker-compose.yml <<'EOF'
version: '3.8'

services:
  cangjie-mem:
    image: ghcr.io/ystyle/cangjie-mem:v1.3.0
    container_name: cangjie-mem
    restart: unless-stopped
    ports:
      - "8080:8080"
    environment:
      - CANGJIE_HTTP=true
      # - CANGJIE_TOKEN=your-secret-token  # 可选：启用认证
    volumes:
      - cangjie-data:/home/cangjie/.cangjie-mem

volumes:
  cangjie-data:
EOF

# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

#### 方式 2：Docker Run

```bash
docker run -d \
  --name cangjie-mem \
  -p 8080:8080 \
  -v cangjie-data:/home/cangjie/.cangjie-mem \
  -e CANGJIE_HTTP=true \
  ghcr.io/ystyle/cangjie-mem:v1.3.0
```

**⚠️ 安全提示**：
- **无 Token**：任何人都能访问，仅适合本地开发
- **有 Token**：需要认证才能访问，适合内网使用
- **生产环境**：建议启用 Token 并使用 HTTPS

#### 3. 配置 Claude Code

在 Claude Code 的配置文件中添加：

```json
{
  "mcpServers": {
    "cangjie-mem": {
      "transport": "http",
      "url": "http://localhost:8080/mcp",
      "headers": {
        "X-MCP-Token": "your-secret-token"
      }
    }
  }
}
```

> **提示**：如果启用了 Token 认证，需要在 `headers` 中添加 `X-MCP-Token`

---

**两种模式对比**：

| 特性 | stdio 模式 | HTTP 模式 |
|------|-----------|-----------|
| **使用场景** | 个人本地开发 | 团队协作、远程访问 |
| **启动方式** | Claude Code 自动启动 | Docker/手动启动服务 |
| **配置难度** | 简单 | 需要配置服务器 |
| **网络访问** | 本地 | 可远程访问 |
| **数据共享** | 本地独占 | 多设备共享 |

---

## 💻 开发者指南

### 从源码编译

项目使用 [Task](https://taskfile.dev/) 作为构建工具：

**安装 Task**（可选）：
```bash
# Linux/macOS
brew install go-task/tap/go-task

# 或使用安装脚本
sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b /usr/local/bin

# Windows
scoop install task
# 或
choco install go-task
```

**编译项目**：
```bash
# 克隆仓库
git clone https://github.com/ystyle/cangjie-mem.git
cd cangjie-mem

# 使用 Task 编译
task build

# 或使用 Go 命令
go build -o cangjie-mem ./cmd/server

# 运行
task run
# 或
./cangjie-mem
```

**可用命令**：
```bash
task build        # 构建当前平台
task test         # 运行测试
task clean        # 清理构建文件
task deps         # 下载依赖
task run          # 运行服务器（stdio 模式）
task run-http     # 运行服务器（HTTP 模式）
```

**运行测试**：
```bash
# 使用 Task（推荐）
task test

# 或使用 Go 命令（需要 sqlite_fts5 编译标签）
go test ./... -tags="sqlite_fts5"

# 运行特定包的测试
go test ./pkg/db -tags="sqlite_fts5"
go test ./internal/store -tags="sqlite_fts5"

# 查看详细测试输出
go test ./... -tags="sqlite_fts5" -v
```

> **⚠️ 注意**：项目使用 SQLite FTS5 全文搜索功能，测试时需要添加 `-tags="sqlite_fts5"` 编译标签。

**Docker 本地构建**：
```bash
# 构建镜像
docker build -t cangjie-mem:latest .

# 运行容器
docker run -d --name cangjie-mem -p 8080:8080 \
  -e CANGJIE_HTTP=true \
  cangjie-mem:latest
```

**Docker 环境变量**：

| 环境变量 | 说明 | 默认值 | 必需 |
|---------|------|--------|------|
| `CANGJIE_HTTP` | 启用 HTTP 模式 | `false` | ✅ Docker 部署必需 |
| `CANGJIE_ADDR` | HTTP 监听地址 | `:8080` | 否 |
| `CANGJIE_ENDPOINT` | HTTP 端点路径 | `/mcp` | 否 |
| `CANGJIE_TOKEN` | HTTP 认证 Token | 空 | 否 |
| `CANGJIE_STATELESS` | 无状态模式 | `false` | 否 |
| `CANGJIE_DB_PATH` | 数据库文件路径 | `~/.cangjie-mem/memory.db` | 否 |

## 🛠️ MCP 工具

### 工具 1：存储记忆

**`cangjie_mem_store`** - 存储仓颉语言的实践经验记忆

**支持三级记忆模型**：
- `language`：语言级（语法、关键字、核心语义）
- `project`：项目级（项目配置、业务逻辑、约定）
- `library`：公共库级（设计模式、工具函数、第三方库用法）

**参数**：
- `level`（必需）：记忆层级（language/project/library）
- `title`（必需）：记忆标题
- `content`（必需）：记忆内容
- `library_name`（可选）：库名（用于 library 层级，如：tang、http-client）
- `project_path_pattern`（可选）：项目路径模式（project 层级必需）
- `summary`（可选）：简短摘要
- `source`（可选）：来源（manual 或 auto_captured）

**示例**：

```
请帮我记录：仓颉语言中接口定义使用 'interface' 关键字，类似于 Java 的接口
```

```
存储：Tang 框架使用 RouterGroup 配置路由分组
```

### 工具 2：回忆记忆（核心）

**`cangjie_mem_recall`** - 基于关键词智能检索仓颉语言记忆

**搜索模式**：使用**空格分隔的 AND 匹配**
- 多个关键词必须**同时出现**才会匹配
- 关键词越多，结果越精准

**参数**：
- `query`（必需）：查询内容（空格分隔的关键词）
- `level`（可选）：记忆层级（通常让 AI 自动判断）
- `project_context`（可选）：项目路径
- `max_results`（可选）：最大返回数量（默认 10）
- `min_confidence`（可选）：最小置信度（默认 0.5）

**示例**：

```
我项目中之前是怎么处理泛型约束的？
```

```
仓颉语言中如何定义接口？
```

```
怎么用仓颉处理 JSON 数据？
```

```
tang 路由 中间件  # 搜索同时包含 "tang"、"路由"、"中间件" 的记忆
```

### 工具 3：列出记忆

**`cangjie_mem_list`** - 浏览记忆，支持按层级、库名、项目路径筛选

**使用场景**：
- 浏览特定库的所有知识点（如：tang 库的所有记忆）
- 浏览特定项目的所有记忆
- 浏览特定层级的所有记忆

**参数**：
- `level`（可选）：记忆层级（language/project/library）
- `library_name`（可选）：库名筛选（如：tang）
- `project_path_pattern`（可选）：项目路径模式筛选
- `limit`（可选）：返回数量（默认 20）
- `offset`（可选）：分页偏移（默认 0）
- `order_by`（可选）：排序字段（created_at/access_count/updated_at）

**示例**：

```
列出所有 tang 库相关的记忆
```

```
列出所有项目级的记忆
```

```
列出最近访问最多的 10 条记忆
```

### 工具 4：列出分类

**`cangjie_mem_list_categories`** - 列出所有库和项目分类（仅返回名称和统计）

**使用场景**：
- 查看都记录了哪些第三方库及其知识点数量
- 查看都有哪些项目及其记忆数量
- 快速浏览知识库的整体结构

**参数**：
- `language_tag`（可选）：语言标签（默认 cangjie）

**返回格式**：
```json
{
  "libraries": [
    {"name": "tang", "count": 12},
    {"name": "http-client", "count": 5}
  ],
  "projects": [
    {"name": "/home/user/tang-web/*", "count": 15}
  ]
}
```

### 工具 5：删除记忆

**`cangjie_mem_delete`** - 删除指定 ID 的记忆

**使用场景**：
- 删除错误的记忆
- 配合 `cangjie_mem_list` 实现"更新"效果（先删除旧记忆，再插入新记忆）
- 提炼项目记忆为库级记忆后，删除原始项目记忆

**参数**：
- `id`（必需）：记忆 ID

**示例**：

```
删除 ID 为 123 的记忆
```

## 🎨 使用示例

### 场景 1：记录项目约定

```
请存储记忆：我们的项目使用 three-tier 架构，
所有 API 接口都放在 /api 目录下，使用仓颉的 struct 定义数据模型
```

### 场景 2：查询项目配置

```
我项目的日志配置文件在哪里？
```

AI 会自动识别这是项目级问题，并返回相关的项目记忆。

### 场景 3：学习语言语法

```
仓颉语言中如何定义泛型函数？
```

AI 会自动识别这是语言级问题，优先返回语言级记忆。

## 🛠️ 技术架构

### 技术栈

- **语言**：Go 1.21+
- **协议**：Model Context Protocol (MCP)
- **存储**：SQLite (支持 FTS5 全文搜索)
- **传输**：stdio (本地模式)

### 项目结构

```
cangjie-mem/
├── cmd/
│   └── server/          # 主入口
├── pkg/
│   ├── db/              # 数据库层
│   ├── mcp/             # MCP 服务器
│   └── types/           # 类型定义
├── internal/
│   ├── config/          # 配置管理
│   └── store/           # 智能检索逻辑
├── DESIGN.md            # 详细设计文档
└── README.md            # 本文件
```

## 📊 数据模型

### 记忆层级

```sql
CREATE TABLE knowledge_base (
    id INTEGER PRIMARY KEY,
    level TEXT NOT NULL,              -- 'language' | 'project' | 'library'
    language_tag TEXT NOT NULL,       -- 'cangjie'
    project_path_pattern TEXT,        -- 项目路径模式（通配符支持）
    title TEXT NOT NULL,              -- 标题
    content TEXT NOT NULL,            -- 内容
    summary TEXT,                     -- 摘要
    source TEXT,                      -- 'manual' | 'auto_captured'
    access_count INTEGER DEFAULT 0,   -- 访问次数
    confidence REAL DEFAULT 1.0,      -- 置信度
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    last_accessed_at TIMESTAMP
);
```

## 🔧 配置选项

### 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `DB_PATH` | 数据库文件路径 | `~/.cangjie-mem/memory.db` |
| `LOG_LEVEL` | 日志级别 | `info` |

### 命令行参数

```bash
cangjie-mem [options]

Options:
  -db string
        数据库文件路径（默认 ~/.cangjie-mem/memory.db）
  -version
        显示版本信息
  -http
        启用 HTTP 模式（Streamable HTTP）
  -addr string
        HTTP 监听地址（默认 :8080）
  -endpoint string
        HTTP 端点路径（默认 /mcp）
  -stateless
        无状态模式（默认 false）
  -token string
        HTTP 认证 Token（留空则不启用认证）
```

## 🚧 开发计划

### Phase 1：MVP ✅

- [x] 基础数据模型
- [x] SQLite 存储
- [x] 两个核心工具（store/recall）
- [x] 智能层级判断
- [x] 置信度评分

### Phase 2：功能增强 🚧

- [x] HTTP/SSE 远程模式（Streamable HTTP）
- [x] HTTP Token 认证
- [x] Docker 部署支持
- [ ] CLI 管理工具
- [ ] 自动摘要生成
- [ ] 访问统计和热度排序

### Phase 3：智能化

- [ ] 向量语义搜索
- [ ] 自动知识提取
- [ ] 知识图谱关联
- [ ] 多语言支持

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件

## 🙏 致谢

- [cangjie-docs-mcp](https://github.com/ystyle/cangjie-docs-mcp) - 仓颉语言文档检索系统
- [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go) - MCP Go SDK

---

**Made with ❤️ by ystyle**
