package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/ystyle/cangjie-mem/pkg/mcp"
	"github.com/ystyle/cangjie-mem/pkg/version"
)

func main() {
	// 命令行参数
	dbPath := flag.String("db", "", "数据库文件路径（默认 ~/.cangjie-mem/memory.db）")
	showVersion := flag.Bool("version", false, "显示版本信息")

	// HTTP 模式参数
	httpMode := flag.Bool("http", false, "启用 HTTP 模式（Streamable HTTP）")
	httpAddr := flag.String("addr", ":8080", "HTTP 监听地址（默认 :8080）")
	httpEndpoint := flag.String("endpoint", "/mcp", "HTTP 端点路径（默认 /mcp）")
	stateless := flag.Bool("stateless", false, "无状态模式（默认 false）")
	httpToken := flag.String("token", "", "HTTP 认证 Token（留空则不启用认证）")

	flag.Parse()

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

	// 根据模式运行服务器
	if *httpMode {
		// HTTP 模式
		log.Printf("Starting cangjie-mem HTTP server on %s%s...", *httpAddr, *httpEndpoint)
		if *httpToken != "" {
			log.Printf("🔐 Token authentication enabled - clients must provide X-MCP-Token header")
		} else {
			log.Printf("⚠️  WARNING: No authentication configured - server is open to all requests")
		}

		// 配置 HTTP 选项
		opts := []mcpserver.StreamableHTTPOption{
			mcpserver.WithEndpointPath(*httpEndpoint),
		}
		if *stateless {
			opts = append(opts, mcpserver.WithStateLess(true))
		}

		if err := server.RunHTTPWithOpts(*httpAddr, opts...); err != nil {
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
