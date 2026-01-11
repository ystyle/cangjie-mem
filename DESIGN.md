# 仓颉语言分级记忆库 MCP 服务器设计文档

## 1. 项目概述

**项目名称**：cangjie-mem
**核心目标**：构建一个专用于仓颉编程语言的、支持多级别（语言/项目/公共库）知识智能管理与检索的 MCP 服务器，以解决通用 AI 工具在识别和运用新语言语法时的知识缺失与上下文遗忘问题。
**设计原则**：分级存储、智能检索、主动学习、跨环境共享。

**与 cangjie-docs-mcp 的区别**：

| 特性 | cangjie-docs-mcp | cangjie-mem |
|------|-----------------|-------------|
| **定位** | 官方文档搜索 | 实践经验记忆库 |
| **内容** | 公开的、标准的、权威的 | 个人的、实践的、演进的 |
| **类比** | 教科书/参考手册 | 笔记本/经验库 |
| **更新频率** | 随官方文档更新 | 持续积累和演进 |

## 2. 核心架构与设计

### 2.1 分级知识模型

知识严格分为三个层级，每个层级对应不同的适用范围和权威性：

| 层级 | 标识 | 描述与内容示例 | 存储与匹配关键 |
| :--- | :--- | :--- | :--- |
| **语言级** | `language` | **权威规范**。包括语言语法、关键字、核心语义、编译器特性等（如"仓颉语言的函数定义格式"）。 | 通过 `language_tag` (如 `cangjie`) 标识。 |
| **项目级** | `project` | **具体上下文**。包括特定项目的结构说明、模块约定、调试记录、业务逻辑（如"项目A的启动配置"）。 | 通过 `project_path_pattern` 与当前项目路径匹配。 |
| **公共库级** | `library` | **可复用解决方案**。包括通用工具函数、设计模式、最佳实践、第三方库用法（如"如何用仓颉语言处理JSON"）。 | 通常与特定 `language_tag` 关联，无项目路径限制。 |

### 2.2 技术栈选型

| 组件 | 选型 | 理由 |
| :--- | :--- | :--- |
| **传输协议** | **stdio (本地) / HTTP+SSE (远程)** | stdio 是 MCP 标准方式，HTTP+SSE 支持远程访问和跨环境共享。 |
| **存储引擎** | **SQLite** | 单文件、零配置，支持 FTS5 全文检索，可通过扩展支持向量搜索。 |
| **服务器语言** | **Go** | 用户熟悉，编译为单一二进制，部署简便，性能优异。 |
| **协议** | **Model Context Protocol** | Claude Code 原生支持，实现与 AI 工具的标准集成。 |

### 2.3 系统架构图

```
+-------------------+      stdio/HTTP       +-----------------------+
|   Claude Code     | <------------------> |   Go MCP Server        |
|   (客户端)        |      (工具调用)       |   (记忆库服务端)        |
+-------------------+                       +-----------------------+
         |                                          |
         | (配置项目路径等上下文)                   | (智能检索与存储)
         v                                          v
+-------------------+                     +-----------------------+
|   用户项目环境    |                     |     SQLite 数据库      |
| (如`/path/to/proj`)|                     |  (分级记忆表、FTS5索引) |
+-------------------+                     +-----------------------+
```

## 3. 详细设计

### 3.1 数据表设计

#### 主表：`knowledge_base`

```sql
CREATE TABLE knowledge_base (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    -- 分级核心字段
    level TEXT NOT NULL CHECK (level IN ('language', 'project', 'library')),
    language_tag TEXT NOT NULL DEFAULT 'cangjie',
    project_path_pattern TEXT, -- 支持通配符，如 `/projects/cangjie-*`

    -- 内容字段
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    summary TEXT, -- 自动生成的摘要

    -- 来源与元数据
    source TEXT CHECK (source IN ('manual', 'auto_captured')) DEFAULT 'manual',

    -- 统计与排序
    access_count INTEGER DEFAULT 0,
    confidence REAL DEFAULT 1.0, -- manual: 1.0, auto_captured: 0.5-0.8

    -- 时间戳
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_accessed_at TIMESTAMP
);

-- 索引
CREATE INDEX idx_knowledge_level ON knowledge_base(level);
CREATE INDEX idx_knowledge_language ON knowledge_base(language_tag);
CREATE INDEX idx_knowledge_project_pattern ON knowledge_base(project_path_pattern);
CREATE INDEX idx_knowledge_created_at ON knowledge_base(created_at DESC);

-- FTS5 全文搜索虚拟表
CREATE VIRTUAL TABLE knowledge_base_fts USING fts5(
    title,
    content,
    summary,
    content=knowledge_base,
    content_rowid=rowid
);

-- 自动同步触发器
CREATE TRIGGER knowledge_base_ai AFTER INSERT ON knowledge_base BEGIN
  INSERT INTO knowledge_base_fts(rowid, title, content, summary)
  VALUES (new.id, new.title, new.content, new.summary);
END;

CREATE TRIGGER knowledge_base_ad AFTER DELETE ON knowledge_base BEGIN
  INSERT INTO knowledge_base_fts(knowledge_base_fts, rowid, title, content, summary)
  VALUES ('delete', old.id, old.title, old.content, old.summary);
END;

CREATE TRIGGER knowledge_base_au AFTER UPDATE ON knowledge_base BEGIN
  INSERT INTO knowledge_base_fts(knowledge_base_fts, rowid, title, content, summary)
  VALUES ('delete', old.id, old.title, old.content, old.summary);
  INSERT INTO knowledge_base_fts(rowid, title, content, summary)
  VALUES (new.id, new.title, new.content, new.summary);
END;
```

#### 标签表（可选扩展）

```sql
CREATE TABLE knowledge_tags (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    knowledge_id INTEGER NOT NULL,
    tag TEXT NOT NULL,
    FOREIGN KEY (knowledge_id) REFERENCES knowledge_base(id) ON DELETE CASCADE
);

CREATE INDEX idx_tags_knowledge_id ON knowledge_tags(knowledge_id);
CREATE INDEX idx_tags_tag ON knowledge_tags(tag);
```

### 3.2 MCP 工具接口设计

#### 工具 1：`cangjie_mem_store` - 存储记忆

**描述**：存储仓颉语言的实践经验记忆，支持语言级/项目级/公共库级三个层级。

**参数**：
```json
{
  "level": "language",          // 必需：'language' | 'project' | 'library'
  "language_tag": "cangjie",    // 可选，默认 'cangjie'
  "project_path_pattern": null, // level='project' 时必需
  "title": "变量声明语法",       // 必需：记忆标题
  "content": "详细说明...",      // 必需：记忆内容
  "summary": "简短摘要",         // 可选，自动生成
  "source": "manual"            // 可选，'manual' | 'auto_captured'
}
```

**返回**：
```json
{
  "success": true,
  "id": 123,
  "message": "记忆已成功存储"
}
```

---

#### 工具 2：`cangjie_mem_recall` - 回忆记忆（核心）

**描述**：根据当前项目上下文和问题，智能检索最相关的仓颉语言记忆。这是获取语言规范、项目特定信息或公共库解决方案的首要工具。

**参数**：
```json
{
  "query": "如何定义接口？",           // 必需：查询内容
  "level": null,                      // 可选：显式指定层级（null=自动判断）
  "language_tag": "cangjie",          // 可选，默认 'cangjie'
  "project_context": "/home/user/project", // 可选，由 Claude Code 自动传入
  "max_results": 10,                  // 可选，最大返回数量
  "min_confidence": 0.5               // 可选，最小置信度阈值
}
```

**返回**：
```json
{
  "total": 5,
  "results": [
    {
      "id": 123,
      "level": "language",
      "title": "接口定义语法",
      "content": "仓颉语言中...",
      "confidence": 0.95,
      "access_count": 42
    }
  ],
  "search_strategy": "auto_determined_language" // 使用的检索策略
}
```

**智能检索逻辑**（服务端自动执行，无需客户端指定）：

1. **精确匹配**（置信度 1.0）：
   - 标题或内容的精确匹配

2. **项目上下文匹配**（置信度 0.9-1.0）：
   - 如果 `project_context` 匹配某 `project_path_pattern`
   - 且查询包含项目相关词（如"我项目"、"这里的配置"）
   - 优先返回项目级记忆

3. **语言级匹配**（置信度 0.7-0.9）：
   - 查询核心语法、关键字
   - 返回语言级记忆

4. **公共库级匹配**（置信度 0.5-0.7）：
   - 查询通用功能、模式
   - 返回公共库级记忆

5. **全文搜索补充**：
   - 使用 FTS5 对所有层级进行全文搜索
   - 作为补充结果返回

---

#### 工具 3：`cangjie_mem_suggest` - 建议补充（可选）

**描述**：当记忆库中找不到相关内容时，建议补充缺失的知识。

**参数**：
```json
{
  "query": "如何实现泛型约束？",
  "suggested_title": "泛型约束实现方法",
  "suggested_content": "应该在仓颉语言中...",
  "suggested_level": "language",
  "reason": "未找到相关记忆，建议补充"
}
```

**返回**：
```json
{
  "success": true,
  "suggestion_id": 456,
  "status": "pending_review", // 待审核
  "message": "建议已记录，待审核后加入记忆库"
}
```

### 3.3 数据模型（Go）

```go
package types

// KnowledgeLevel 记忆层级
type KnowledgeLevel string

const (
    LevelLanguage KnowledgeLevel = "language" // 语言级
    LevelProject  KnowledgeLevel = "project"  // 项目级
    LevelLibrary  KnowledgeLevel = "library"  // 公共库级
)

// KnowledgeSource 记忆来源
type KnowledgeSource string

const (
    SourceManual         KnowledgeSource = "manual"          // 手动录入
    SourceAutoCaptured   KnowledgeSource = "auto_captured"   // 自动捕获
)

// Memory 记忆条目
type Memory struct {
    ID                 int64            `json:"id"`
    Level              KnowledgeLevel   `json:"level"`
    LanguageTag        string           `json:"language_tag"`
    ProjectPathPattern string           `json:"project_path_pattern,omitempty"`
    Title              string           `json:"title"`
    Content            string           `json:"content"`
    Summary            string           `json:"summary,omitempty"`
    Source             KnowledgeSource  `json:"source"`
    AccessCount        int              `json:"access_count"`
    Confidence         float64          `json:"confidence"`
    CreatedAt          time.Time        `json:"created_at"`
    UpdatedAt          time.Time        `json:"updated_at"`
    LastAccessedAt     *time.Time       `json:"last_accessed_at,omitempty"`
}

// StoreRequest 存储请求
type StoreRequest struct {
    Level              KnowledgeLevel  `json:"level" required:"true"`
    LanguageTag        string          `json:"language_tag"`
    ProjectPathPattern string          `json:"project_path_pattern,omitempty"`
    Title              string          `json:"title" required:"true"`
    Content            string          `json:"content" required:"true"`
    Summary            string          `json:"summary,omitempty"`
    Source             KnowledgeSource `json:"source"`
}

// RecallRequest 回忆请求
type RecallRequest struct {
    Query          string  `json:"query" required:"true"`
    Level          string  `json:"level,omitempty"` // 空字符串表示自动判断
    LanguageTag    string  `json:"language_tag"`
    ProjectContext string  `json:"project_context,omitempty"`
    MaxResults     int     `json:"max_results"`
    MinConfidence  float64 `json:"min_confidence"`
}

// RecallResult 回忆结果
type RecallResult struct {
    ID            int64             `json:"id"`
    Level         KnowledgeLevel    `json:"level"`
    Title         string            `json:"title"`
    Content       string            `json:"content"`
    Summary       string            `json:"summary,omitempty"`
    Confidence    float64           `json:"confidence"`
    AccessCount   int               `json:"access_count"`
    MatchedText   string            `json:"matched_text,omitempty"` // 匹配的文本片段
}

// RecallResponse 回忆响应
type RecallResponse struct {
    Total          int            `json:"total"`
    Results        []RecallResult `json:"results"`
    SearchStrategy string         `json:"search_strategy"` // 使用的检索策略
}
```

## 4. 智能检索算法

### 4.1 置信度计算

```go
func calculateConfidence(result RecallResult, query string, projectPath string) float64 {
    base := 0.5

    // 1. 精确匹配加分
    if strings.Contains(result.Title, query) || strings.Contains(result.Content, query) {
        base = 1.0
    }

    // 2. 项目上下文匹配
    if result.Level == LevelProject && projectPath != "" {
        if matchesProjectPattern(projectPath, result.ProjectPathPattern) {
            base += 0.3
        }
    }

    // 3. 语言级权威性
    if result.Level == LevelLanguage {
        base += 0.2
    }

    // 4. 来源可信度
    if result.Source == SourceManual {
        base += 0.1
    }

    // 5. 访问热度（轻微影响）
    if result.AccessCount > 10 {
        base += 0.05
    }

    return math.Min(base, 1.0)
}
```

### 4.2 自动层级判断

```go
func determineLevel(query string, projectPath string) KnowledgeLevel {
    // 1. 项目相关关键词
    projectKeywords := []string{"我项目", "这里的", "我们", "当前项目", "配置文件"}
    for _, kw := range projectKeywords {
        if strings.Contains(query, kw) && projectPath != "" {
            return LevelProject
        }
    }

    // 2. 语言级关键词
    langKeywords := []string{"语法", "定义", "关键字", "类型", "接口", "函数", "变量"}
    for _, kw := range langKeywords {
        if strings.Contains(query, kw) {
            return LevelLanguage
        }
    }

    // 3. 默认返回公共库级
    return LevelLibrary
}
```

## 5. 配置与部署

### 5.1 Claude Code 客户端配置

#### 本地模式（stdio）
```json
{
  "mcpServers": {
    "cangjie-mem": {
      "command": "/usr/local/bin/cangjie-mem",
      "env": {
        "DB_PATH": "/Users/ystyle/.cangjie-mem/memory.db",
        "LOG_LEVEL": "info"
      }
    }
  }
}
```

#### 远程模式（HTTP/SSE）
```json
{
  "mcpServers": {
    "cangjie-mem-remote": {
      "command": "http://home-server:8080/sse",
      "headers": {
        "Authorization": "Bearer ${CANGJIE_MEM_TOKEN}"
      }
    }
  }
}
```

### 5.2 服务器配置

```yaml
# config.yaml
database:
  path: ~/.cangjie-mem/memory.db

server:
  mode: stdio # stdio | http
  host: "0.0.0.0"
  port: 8080
  auth_token: "${CANGJIE_MEM_TOKEN}"

memory:
  default_language: "cangjie"
  max_results: 10
  min_confidence: 0.5
  enable_auto_summary: true
```

## 6. 工作流程

### 6.1 知识积累流程

1. **手动录入**：通过 `cangjie_mem_store` 添加权威语言知识
2. **自动捕获**：（可选）通过开发钩子自动记录项目上下文
3. **持续演进**：自动捕获的内容可被提炼为权威知识

### 6.2 智能查询流程

1. 用户在 Claude Code 中提出仓颉语言相关问题
2. AI 根据工具描述，优先调用 `cangjie_mem_recall`
3. 服务器接收查询及 `project_context`
4. 智能判断最佳层级并检索
5. 返回按置信度排序的结果
6. 如果找不到相关内容，AI 可以调用 `cangjie_mem_suggest` 记录缺失

## 7. 扩展性设计

### 7.1 未来可扩展功能

- **向量语义搜索**：集成 Embedding 模型进行语义匹配
- **知识图谱**：建立记忆之间的关联关系
- **多语言支持**：扩展到其他编程语言（Rust、Go 等）
- **自动摘要**：使用 LLM 自动生成内容摘要
- **智能去重**：检测并合并相似的记忆

### 7.2 数据导出/导入

```json
{
  "version": "1.0",
  "exported_at": "2026-01-11T12:00:00Z",
  "memories": [
    {
      "level": "language",
      "title": "接口定义",
      "content": "...",
      "tags": ["语法", "接口"]
    }
  ]
}
```

## 8. 设计亮点总结

1. **真·分级智能**：通过 `level`、`language_tag`、`project_path_pattern` 实现上下文感知
2. **主动引导 AI**：通过 `cangjie_mem_` 前缀命名，引导 Claude 优先使用专用工具
3. **跨环境中心化**：HTTP/SSE 协议支持家庭服务器部署，一处维护，多处同步
4. **扩展性基础**：清晰的分层架构，为未来功能扩展预留空间
5. **互补设计**：与 `cangjie-docs-mcp` 形成互补，文档查规范，记忆查经验

---

*设计版本：v1.0*
*最后更新：2026-01-11*
*作者：ystyle*
