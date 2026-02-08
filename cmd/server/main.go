package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/ystyle/cangjie-mem/internal/api"
	"github.com/ystyle/cangjie-mem/pkg/mcp"
	"github.com/ystyle/cangjie-mem/pkg/version"
)

// getEnvOrDefault 获取环境变量，如果不存在则返回默认值
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvBool 获取布尔环境变量
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return strings.ToLower(value) == "true" || value == "1"
	}
	return defaultValue
}

func main() {
	// 命令行参数
	dbPath := flag.String("db", "", "数据库文件路径（默认 ~/.cangjie-mem/memory.db）")
	showVersion := flag.Bool("version", false, "显示版本信息")

	// HTTP 模式参数
	httpMode := flag.Bool("http", false, "启用 HTTP 模式（Streamable HTTP）")
	httpAddr := flag.String("addr", ":8080", "HTTP 监听地址（默认 :8080）")
	httpEndpoint := flag.String("endpoint", "/mcp", "HTTP MCP 端点路径（默认 /mcp）")
	stateless := flag.Bool("stateless", false, "无状态模式（默认 false）")
	httpToken := flag.String("token", "", "HTTP 认证 Token（留空则不启用认证）")

	// 新增：API 和 UI 功能开关
	enableAPI := flag.Bool("api", false, "启用 REST API（默认 false）")
	enableUI := flag.Bool("ui", false, "启用 Web UI（默认 false）")

	flag.Parse()

	// 环境变量覆盖（优先级高于命令行参数）
	if envDB := getEnvOrDefault("CANGJIE_DB_PATH", *dbPath); envDB != "" {
		dbPath = &envDB
	}
	if envHTTP := getEnvBool("CANGJIE_HTTP", *httpMode); envHTTP {
		httpMode = &envHTTP
	}
	if envAddr := getEnvOrDefault("CANGJIE_ADDR", *httpAddr); envAddr != "" {
		httpAddr = &envAddr
	}
	if envEndpoint := getEnvOrDefault("CANGJIE_ENDPOINT", *httpEndpoint); envEndpoint != "" {
		httpEndpoint = &envEndpoint
	}
	if envStateless := getEnvBool("CANGJIE_STATELESS", *stateless); envStateless {
		stateless = &envStateless
	}
	if envToken := getEnvOrDefault("CANGJIE_TOKEN", *httpToken); envToken != "" {
		httpToken = &envToken
	}
	if envAPI := getEnvBool("CANGJIE_API_ENABLED", *enableAPI); envAPI {
		enableAPI = &envAPI
	}
	if envUI := getEnvBool("CANGJIE_UI_ENABLED", *enableUI); envUI {
		enableUI = &envUI
	}

	if *showVersion {
		fmt.Printf("cangjie-mem %s\n", version.Version)
		fmt.Printf("Git commit: %s\n", version.GitCommit)
		fmt.Printf("Build date: %s\n", version.BuildDate)
		os.Exit(0)
	}

	// 创建 MCP 服务器
	cfg := mcp.Config{
		DBPath:    *dbPath,
		HTTPToken: *httpToken,
	}

	server, err := mcp.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}
	defer server.Close()

	// 检查是否启用多端点模式
	if *enableAPI || *enableUI {
		// 多端点模式：使用统一的 HTTP 服务器
		runMultiEndpointServer(server, *httpAddr, *httpEndpoint, *httpToken, *enableAPI, *enableUI, *stateless)
	} else {
		// 原有模式：仅 MCP
		runLegacyServer(server, *httpMode, *httpAddr, *httpEndpoint, *httpToken, *stateless)
	}
}

// runMultiEndpointServer 运行多端点服务器（MCP + API + UI）
func runMultiEndpointServer(mcpServer *mcp.Server, addr, mcpEndpoint, token string, enableAPI, enableUI, stateless bool) {
	mux := http.NewServeMux()

	// 创建 MCP HTTP 处理器
	mcpHTTPServer := mcpserver.NewStreamableHTTPServer(mcpServer.GetMCPServer(), mcpserver.WithEndpointPath(mcpEndpoint))

	// 注册 MCP 端点
	if token != "" {
		// 需要认证 - 使用 mcpServer 内部的认证逻辑或在这里包装
		mux.Handle(mcpEndpoint, &tokenAuthHandler{
			next:  mcpHTTPServer,
			token: token,
		})
	} else {
		mux.Handle(mcpEndpoint, mcpHTTPServer)
	}
	log.Printf("✓ MCP 端点已注册: %s", mcpEndpoint)

	// 注册 REST API 端点和 Web UI
	if enableAPI || enableUI {
		// 创建 API 服务器（复用 store）
		apiServer := api.NewWithStore(mcpServer.GetStore())

		if enableAPI {
			apiServer.RegisterRoutes(mux)
		}

		if enableUI {
			apiServer.RegisterStatic(mux)
		}
	}

	// 启动 HTTP 服务器
	log.Printf("Starting cangjie-mem HTTP server on %s", addr)
	log.Printf("  - MCP: http://localhost%s%s", addr, mcpEndpoint)
	if enableAPI {
		log.Printf("  - API: http://localhost%s/api/*", addr)
	}
	if enableUI {
		log.Printf("  - UI:  http://localhost%s/", addr)
	}

	if token != "" {
		log.Printf("🔐 Token authentication enabled")
	} else {
		log.Printf("⚠️  WARNING: No authentication configured")
	}

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("HTTP server error: %v", err)
	}
}

// runLegacyServer 运行原有的服务器模式（仅 MCP）
func runLegacyServer(server *mcp.Server, httpMode bool, addr, endpoint, token string, stateless bool) {
	// 根据模式运行服务器
	if httpMode {
		// HTTP 模式
		log.Printf("Starting cangjie-mem HTTP server on %s%s...", addr, endpoint)
		if token != "" {
			log.Printf("🔐 Token authentication enabled - clients must provide X-MCP-Token header")
		} else {
			log.Printf("⚠️  WARNING: No authentication configured - server is open to all requests")
		}

		// 配置 HTTP 选项
		opts := []mcpserver.StreamableHTTPOption{
			mcpserver.WithEndpointPath(endpoint),
		}
		if stateless {
			opts = append(opts, mcpserver.WithStateLess(true))
		}

		if err := server.RunHTTPWithOpts(addr, opts...); err != nil {
			log.Fatalf("HTTP server error: %v", err)
		}
	} else {
		// stdio 模式（默认）
		log.Println("Starting cangjie-mem MCP server in stdio mode...")
		if err := server.Run(); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}
}

// tokenAuthHandler Token 认证中间件
type tokenAuthHandler struct {
	next       http.Handler // 下一个处理器（MCP Server）
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

// newTokenAuthHandler 创建 Token 认证处理器
func newTokenAuthHandler(next http.Handler, token, serverName string) http.Handler {
	return &tokenAuthHandler{
		next:       next,
		token:      token,
		serverName: serverName,
	}
}
