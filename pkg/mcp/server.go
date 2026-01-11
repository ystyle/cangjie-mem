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
		mcp.WithDescription("存储仓颉语言的实践经验记忆（语言级/项目级/公共库级）"),
		mcp.WithString("level",
			mcp.Required(),
			mcp.Description("记忆层级"),
			mcp.Enum("language", "project", "library"),
		),
		mcp.WithString("language_tag",
			mcp.Description("语言标签（默认 cangjie）"),
		),
		mcp.WithString("project_path_pattern",
			mcp.Description("项目路径模式（project 层级必需）"),
		),
		mcp.WithString("title",
			mcp.Required(),
			mcp.Description("记忆标题"),
		),
		mcp.WithString("content",
			mcp.Required(),
			mcp.Description("记忆内容"),
		),
		mcp.WithString("summary",
			mcp.Description("简短摘要（可选）"),
		),
		mcp.WithString("source",
			mcp.Description("来源（manual 或 auto_captured，默认 manual）"),
			mcp.Enum("manual", "auto_captured"),
		),
	)
	s.server.AddTool(storeTool, s.handleStoreMemory)

	// 工具 2: cangjie_mem_recall
	recallTool := mcp.NewTool("cangjie_mem_recall",
		mcp.WithDescription("回忆仓颉语言相关的实践经验，根据项目上下文智能检索最相关记忆"),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("查询内容"),
		),
		mcp.WithString("level",
			mcp.Description("记忆层级（可选，留空自动判断）"),
			mcp.Enum("language", "project", "library"),
		),
		mcp.WithString("language_tag",
			mcp.Description("语言标签（默认 cangjie）"),
		),
		mcp.WithString("project_context",
			mcp.Description("项目上下文路径（由 Claude Code 自动传入）"),
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
