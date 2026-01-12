# 开发指南

本文档面向 cangjie-mem 的维护者，记录开发、测试和发版流程。

## 开发环境

### 前置要求

- **Go**: 1.21 或更高版本
- **SQLite**: 支持 FTS5 扩展（系统自带或通过 CGo 编译）
- **Docker**: 用于容器化部署和测试
- **Make**: 可选，用于简化构建流程

### 依赖安装

```bash
# 克隆仓库
git clone https://github.com/ystyle/cangjie-mem.git
cd cangjie-mem

# 下载依赖
go mod download

# 整理依赖
go mod tidy
```

## 测试

### 运行测试

**重要**: 测试需要 `sqlite_fts5` 编译标签，以启用 FTS5 全文搜索功能。

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

### 测试数据

测试使用项目目录下的 `./test-data/` 目录，不会影响全局数据：
- 测试数据库文件名格式：`<测试名>.db`
- 测试结束后自动清理
- 使用独特前缀（如 "TEST_FTS5:"）确保 FTS5 搜索结果准确

### 数据库迁移测试

测试旧数据库升级场景：

```bash
# 在本地创建旧数据库
sqlite3 test-old.db "CREATE TABLE knowledge_base (id INTEGER PRIMARY KEY, level TEXT, title TEXT, content TEXT);"

# 启动服务，挂载旧数据库
docker run -v $(pwd)/test-old.db:/home/cangjie/.cangjie-mem/memory.db cangjie-mem:latest

# 检查日志，确认迁移执行
docker logs <container-id>
```

## 构建和部署

### 本地构建

```bash
# 构建当前平台的二进制文件
go build -tags="sqlite_fts5" -o cangjie-mem ./cmd/server

# 交叉编译（需要对应平台的 CGo 工具链）
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -tags="sqlite_fts5" -o cangjie-mem-linux ./cmd/server
```

### Docker 构建

```bash
# 构建本地镜像
docker build -t cangjie-mem:local .

# 构建多平台镜像（需要 buildx）
docker buildx build --platform linux/amd64,linux/arm64 -t cangjie-mem:local .
```

## 发版流程

### 版本号规则

遵循语义化版本（Semantic Versioning）：

- **主版本号**（如 1.x.x）：破坏性变更，不兼容的 API 修改
- **次版本号**（如 x.1.x）：新功能，向后兼容
- **修订号**（如 x.x.1）：Bug 修复，向后兼容

示例：
- `v1.0.0` → `v1.1.0`: 新增 HTTP 模式支持
- `v1.1.0` → `v1.2.0`: 添加 library_name 字段
- `v1.2.0` → `v1.2.1`: 修复数据库迁移 Bug

### 发版步骤

#### 1. 准备工作

确保所有测试通过：
```bash
go test ./... -tags="sqlite_fts5"
```

#### 2. 更新版本信息

更新以下文件中的版本号：
- `pkg/version/version.go`: 修改 `Version` 变量
- README.md（如果需要）：更新功能说明
- DEVELOPMENT.md（本文档）：更新版本历史

#### 3. 提交代码

```bash
git add .
git commit -m "<type>: <description>

<详细说明>"

# 提交类型（type）：
# - feat: 新功能
# - fix: Bug 修复
# - docs: 文档更新
# - refactor: 重构
# - test: 测试相关
# - chore: 构建/工具链相关
```

#### 4. 打 Tag

```bash
# 创建带注释的 tag
git tag v<version> -a -m "v<version>: <description>

# 示例
git tag v1.3.4 -a -m "v1.3.4: 修复数据库迁移 Bug，确保旧版本平滑升级

问题:
- 从旧版本升级时，数据库迁移逻辑静默失败
- 导致服务启动时出现 'no such column: library_name' 错误

修复:
- 改进错误处理：查询失败时强制执行迁移
- 添加迁移日志：输出清晰的迁移进度和结果
- 调整索引创建时机：将 library_name 索引从 init() 移至 migrateLibraryName()

测试:
- 新数据库创建：✓ 字段和索引正常创建
- 旧数据库迁移：✓ 字段和索引成功添加，旧数据完好
- 迁移幂等性：✓ 重复执行不会报错
- Docker 升级测试：✓ 模拟真实升级场景通过"
```

#### 5. 推送到 GitHub

```bash
# 推送代码和 tags
git push
git push --tags
```

#### 6. GitHub Actions 自动构建

推送 tag 后，GitHub Actions 会自动：
1. 构建多平台二进制文件（linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64）
2. 创建 GitHub Release
3. 上传二进制文件作为 Release Assets
4. 构建 Docker 镜像并推送到 ghcr.io

#### 7. 验证 Release

访问 GitHub Release 页面：
- 检查 Release Notes 是否正确
- 下载各平台二进制文件测试
- 验证 Docker 镜像可用：
  ```bash
  docker pull ghcr.io/ystyle/cangjie-mem:v1.3.4
  docker run ghcr.io/ystyle/cangjie-mem:v1.3.4
  ```

### 回滚流程

如果发现严重问题需要回滚：

1. **删除有问题的 tag**：
   ```bash
   git tag -d v<bad-version>
   git push origin :refs/tags/v<bad-version>
   ```

2. **在 GitHub 上删除 Release**：
   - 进入 Release 页面
   - 点击 "Delete release"

3. **发布修复版本**：
   - 修复问题
   - 打新版本号（如 v1.3.4 → v1.3.5）
   - 重新发版

## Git 工作流

### 分支策略

- **master**: 主分支，保持稳定可发布状态
- **feature/<name>**: 功能分支（如果需要）
- **fix/<name>**: Bug 修复分支（如果需要）

### 提交规范

使用约定式提交（Conventional Commits）：

```
<type>(<scope>): <subject>

<body>

<footer>
```

**类型（type）**：
- `feat`: 新功能
- `fix`: Bug 修复
- `docs`: 文档更新
- `style`: 代码格式（不影响功能）
- `refactor`: 重构
- `perf`: 性能优化
- `test`: 测试相关
- `chore`: 构建/工具链相关

**示例**：
```bash
git commit -m "feat(db): 添加 library_name 字段支持

- 添加 library_name 列到 knowledge_base 表
- 创建 idx_knowledge_library 索引
- 更新 Store 和 Recall 方法支持库级别筛选

Closes #123"
```

### 版本历史

| 版本 | 日期 | 说明 |
|------|------|------|
| v1.3.4 | 2026-01-12 | 修复数据库迁移 Bug，确保旧版本平滑升级 |
| v1.3.3 | 2026-01-12 | 恢复 extract-binaries-and-release job 中的 Create Release 步骤 |
| v1.3.2 | 2026-01-12 | 修复 Release 构建问题 |
| v1.3.1 | 2026-01-12 | 添加 macOS 构建支持 |
| v1.3.0 | 2026-01-11 | 添加 library_name 字段支持，新增 list/delete 工具 |
| v1.2.0 | 2026-01-10 | 添加 HTTP/SSE 远程模式 |
| v1.1.0 | 2026-01-09 | 添加 FTS5 全文搜索，空格分隔 AND 匹配 |
| v1.0.0 | 2026-01-08 | 初始版本，基础记忆存储和检索 |

## 常见问题

### FTS5 编译错误

**错误**：`no such table: knowledge_base_fts` 或 FTS5 相关错误

**解决**：确保使用 `-tags="sqlite_fts5"` 编译：
```bash
go build -tags="sqlite_fts5" ./cmd/server
go test ./... -tags="sqlite_fts5"
```

### 跨平台编译失败

**错误**：CGo 跨平台编译需要对应的工具链

**解决**：
- Linux: 使用 `gcc`
- macOS: 使用 Xcode Command Line Tools
- Windows: 使用 MinGW

或者使用 GitHub Actions 自动构建。

### Docker 镜像构建慢

**原因**：每次构建都下载依赖

**优化**：利用 Docker 缓存层级，将依赖下载放在前面：
```dockerfile
COPY go.mod go.sum ./
RUN go mod download
COPY . .
```

### 迁移测试失败

**原因**：旧数据库权限问题

**解决**：确保数据库文件权限正确：
```bash
chmod 644 test-old.db
chown 1000:1000 test-old.db  # Docker 容器中的 cangjie 用户
```

## 参考资源

- [Go 官方文档](https://golang.org/doc/)
- [SQLite FTS5 文档](https://www.sqlite.org/fts5.html)
- [Model Context Protocol](https://modelcontextprotocol.io/)
- [语义化版本](https://semver.org/lang/zh-CN/)
- [约定式提交](https://www.conventionalcommits.org/zh-hans/)
