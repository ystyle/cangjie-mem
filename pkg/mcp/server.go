package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/ystyle/cangjie-mem/internal/store"
	"github.com/ystyle/cangjie-mem/pkg/db"
	"github.com/ystyle/cangjie-mem/pkg/types"
)

// Server MCP 服务器
type Server struct {
	server     *server.MCPServer
	store      *store.Store
	httpToken  string // HTTP 认证 Token
}

// Config 服务器配置
type Config struct {
	DBPath string // 数据库路径

	// HTTP 模式配置
	HTTPAddr      string // HTTP 监听地址（如 ":8080"）
	HTTPEndpoint  string // HTTP 端点路径（默认 "/mcp"）
	HTTPStateless bool   // HTTP 无状态模式（默认 false）
	HTTPToken     string // HTTP 认证 Token（空字符串表示不启用认证）
}

// New 创建新的 MCP 服务器
func New(cfg Config) (*Server, error) {
	// 初始化数据库
	dbConfig := db.Config{Path: cfg.DBPath}
	database, err := db.New(dbConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// 创建 Store
	st := store.New(database)

	// 创建 MCP 服务器
	mcpServer := server.NewMCPServer(
		"cangjie-mem",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	s := &Server{
		server:    mcpServer,
		store:     st,
		httpToken: cfg.HTTPToken,
	}

	// 注册工具
	s.registerTools()

	return s, nil
}

// registerTools 注册所有工具
func (s *Server) registerTools() {
	// 工具 1: cangjie_mem_store
	storeTool := mcp.NewTool("cangjie_mem_store",
		mcp.WithDescription("存储仓颉语言的实践经验记忆。支持三级记忆模型：\n"+
			"- language：语言级（语法、关键字、核心语义）\n"+
			"- project：项目级（项目配置、业务逻辑、约定）\n"+
			"- library：公共库级（设计模式、工具函数、最佳实践）"),
		mcp.WithString("level",
			mcp.Required(),
			mcp.Description("记忆层级（必需：language/project/library）"),
			mcp.Enum("language", "project", "library"),
		),
		mcp.WithString("language_tag",
			mcp.Description("语言标签（默认 cangjie）"),
		),
		mcp.WithString("project_path_pattern",
			mcp.Description("项目路径模式（project 层级必需，如：/path/to/project/*）"),
		),
		mcp.WithString("title",
			mcp.Required(),
			mcp.Description("记忆标题（简短描述，如：接口定义方式、日志配置位置）"),
		),
		mcp.WithString("content",
			mcp.Required(),
			mcp.Description("记忆内容（详细的实践经验、代码示例等）"),
		),
		mcp.WithString("summary",
			mcp.Description("简短摘要（可选，快速浏览时显示）"),
		),
		mcp.WithString("source",
			mcp.Description("来源（manual 手动记录 或 auto_captured AI 捕获，默认 manual）"),
			mcp.Enum("manual", "auto_captured"),
		),
	)
	s.server.AddTool(storeTool, s.handleStoreMemory)

	// 工具 2: cangjie_mem_recall
	recallTool := mcp.NewTool("cangjie_mem_recall",
		mcp.WithDescription("智能回忆仓颉语言实践经验（基于关键词全文搜索）。\n\n"+
			"📌 搜索模式：使用**空格分隔的 AND 匹配**模式\n"+
			"- 多个关键词必须**同时出现**才会匹配\n"+
			"- 关键词越多，结果越精准\n"+
			"- 建议：使用记忆中的核心关键词查询\n\n"+
			"✅ 查询示例：\n"+
			"- 「interface 接口 定义」→ 匹配同时包含这 3 个词的记忆\n"+
			"- 「var 变量 声明」→ 匹配同时包含这 3 个词的记忆\n"+
			"- 「log 日志 配置」→ 匹配包含这些词的配置相关记忆\n\n"+
			"🎯 使用场景：\n"+
			"1. 查询仓颉语法/关键字 → 不传 project_context，自动使用 language 级别\n"+
			"2. 查询项目特定配置 → 传 project_context，自动使用 project 级别\n"+
			"3. 通用设计模式/最佳实践 → 不传 project_context，使用 library 级别\n\n"+
			"💡 提示：通常只需传 query，让 AI 自动判断层级！"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("查询内容（使用空格分隔的关键词，如：interface 接口 定义、var 变量 声明）"),
		),
		mcp.WithString("level",
			mcp.Description("记忆层级（通常不需要传，让 AI 自动判断。强制指定时可选：language/project/library）"),
			mcp.Enum("language", "project", "library"),
		),
		mcp.WithString("language_tag",
			mcp.Description("语言标签（默认 cangjie，通常不需要传）"),
		),
		mcp.WithString("project_context",
			mcp.Description("项目路径（可选。不传时 AI 自动判断层级：通用问题→language，项目特定问题→project）"),
		),
		mcp.WithNumber("max_results",
			mcp.Description("最大返回数量（默认 10）"),
		),
		mcp.WithNumber("min_confidence",
			mcp.Description("最小置信度阈值（默认 0.5）"),
		),
	)
	s.server.AddTool(recallTool, s.handleRecallMemories)
}

// handleStoreMemory 处理存储记忆请求
func (s *Server) handleStoreMemory(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 解析参数
	var req types.StoreRequest
	if err := s.parseRequest(request, &req); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid parameters: %v", err)), nil
	}

	// 存储记忆
	resp, err := s.store.StoreMemory(req)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to store memory: %v", err)), nil
	}

	// 返回结果
	return s.toolResult(resp)
}

// handleRecallMemories 处理回忆记忆请求
func (s *Server) handleRecallMemories(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 解析参数
	var req types.RecallRequest
	if err := s.parseRequest(request, &req); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("invalid parameters: %v", err)), nil
	}

	// 检索记忆
	resp, err := s.store.RecallMemories(req)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to recall memories: %w", err)), nil
	}

	// 返回结果
	return s.toolResult(resp)
}

// parseRequest 解析请求参数
func (s *Server) parseRequest(request mcp.CallToolRequest, dest interface{}) error {
	data, err := json.Marshal(request.Params.Arguments)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

// toolResult 将结果转换为工具响应
func (s *Server) toolResult(result interface{}) (*mcp.CallToolResult, error) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to marshal result: %v", err)), nil
	}
	return mcp.NewToolResultText(string(data)), nil
}

// Run 运行服务器（stdio 模式）
func (s *Server) Run() error {
	return server.ServeStdio(s.server)
}

// RunHTTP 运行 HTTP 服务器（Streamable HTTP 模式）
func (s *Server) RunHTTP(addr string) error {
	// 创建 HTTP 服务器
	httpServer := server.NewStreamableHTTPServer(s.server)

	// 启动服务器
	return httpServer.Start(addr)
}

// RunHTTPWithOpts 使用自定义选项运行 HTTP 服务器
func (s *Server) RunHTTPWithOpts(addr string, opts ...server.StreamableHTTPOption) error {
	// 创建 HTTP 服务器并应用选项
	httpServer := server.NewStreamableHTTPServer(s.server, opts...)

	// 如果设置了 Token，添加认证中间件
	if s.httpToken != "" {
		handler := &tokenAuthHandler{
			next:       httpServer,
			token:      s.httpToken,
			serverName: "cangjie-mem",
		}
		return s.startServerWithHandler(addr, handler)
	}

	// 启动服务器
	return httpServer.Start(addr)
}

// tokenAuthHandler Token 认证中间件
type tokenAuthHandler struct {
	next       http.Handler // 下一个处理器（StreamableHTTPServer）
	token      string       // 期望的 Token
	serverName string       // 服务器名称（用于日志）
}

// ServeHTTP 实现 http.Handler 接口
func (h *tokenAuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 获取客户端提供的 Token
	clientToken := r.Header.Get("X-MCP-Token")

	// 验证 Token
	if clientToken != h.token {
		// Token 验证失败，返回 401
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "Unauthorized", "message": "Invalid or missing X-MCP-Token header"}`))
		return
	}

	// Token 验证成功，转发到下一个处理器
	h.next.ServeHTTP(w, r)
}

// startServerWithHandler 启动带有自定义 handler 的 HTTP 服务器
func (s *Server) startServerWithHandler(addr string, handler http.Handler) error {
	mux := http.NewServeMux()
	mux.Handle("/mcp", handler)

	httpServer := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	return httpServer.ListenAndServe()
}

// Close 关闭服务器
func (s *Server) Close() error {
	// TODO: 关闭数据库连接
	return nil
}
