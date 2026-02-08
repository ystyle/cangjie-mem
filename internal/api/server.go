package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/ystyle/cangjie-mem/internal/store"
	"github.com/ystyle/cangjie-mem/pkg/types"
)

// Server API 服务器
type Server struct {
	store   *store.Store
	webFS   fs.FS // Web 静态文件系统
	apiUser string
	apiPass string
}

// New 创建新的 API 服务器（需要数据库路径，用于兼容）
func New(dbPath string, webFS fs.FS) *Server {
	return &Server{
		store:   nil,
		webFS:   webFS,
		apiUser: os.Getenv("CANGJIE_API_BASIC_AUTH_USERNAME"),
		apiPass: os.Getenv("CANGJIE_API_BASIC_AUTH_PASSWORD"),
	}
}

// NewWithStore 使用现有的 Store 创建 API 服务器
func NewWithStore(store *store.Store, webFS fs.FS) *Server {
	return &Server{
		store:   store,
		webFS:   webFS,
		apiUser: os.Getenv("CANGJIE_API_BASIC_AUTH_USERNAME"),
		apiPass: os.Getenv("CANGJIE_API_BASIC_AUTH_PASSWORD"),
	}
}

// RegisterRoutes 注册 API 路由
func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	// REST API 端点（使用认证 + CORS 包装器）
	mux.HandleFunc("GET /api/health", s.auth(s.cors(s.handleHealth)))
	mux.HandleFunc("GET /api/memories", s.auth(s.cors(s.handleListMemories)))
	mux.HandleFunc("POST /api/memories", s.auth(s.cors(s.handleCreateMemory)))
	mux.HandleFunc("GET /api/memories/", s.auth(s.cors(s.handleMemoryDetail)))
	mux.HandleFunc("PUT /api/memories/", s.auth(s.cors(s.handleUpdateMemory)))
	mux.HandleFunc("DELETE /api/memories/", s.auth(s.cors(s.handleDeleteMemory)))
	mux.HandleFunc("POST /api/search", s.auth(s.cors(s.handleSearch)))
	mux.HandleFunc("GET /api/categories", s.auth(s.cors(s.handleCategories)))
	mux.HandleFunc("POST /api/export", s.auth(s.cors(s.handleExport)))
	mux.HandleFunc("POST /api/import", s.auth(s.cors(s.handleImport)))
	mux.HandleFunc("POST /api/import/confirm", s.auth(s.cors(s.handleImportConfirm)))

	log.Println("✓ REST API 端点已注册: /api/*")
	if s.apiUser != "" && s.apiPass != "" {
		log.Println("✓ REST API Basic Auth 已启用")
	} else {
		log.Println("⚠ REST API 无认证模式（仅适合本地开发）")
	}
}

// auth Basic Auth 中间件包装器
func (s *Server) auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 如果未配置认证，直接通过
		if s.apiUser == "" || s.apiPass == "" {
			next(w, r)
			return
		}

		// 检查 Authorization 头
		auth := r.Header.Get("Authorization")
		if auth == "" {
			s.requestAuth(w)
			return
		}

		// 解析 Basic Auth
		const prefix = "Basic "
		if !strings.HasPrefix(auth, prefix) {
			s.requestAuth(w)
			return
		}

		// 解码 base64
		payload, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
		if err != nil {
			s.requestAuth(w)
			return
		}

		// 分离用户名和密码
		pair := strings.SplitN(string(payload), ":", 2)
		if len(pair) != 2 || pair[0] != s.apiUser || pair[1] != s.apiPass {
			s.requestAuth(w)
			return
		}

		// 认证成功，调用下一个处理器
		next(w, r)
	}
}

// requestAuth 请求认证
func (s *Server) requestAuth(w http.ResponseWriter) {
	w.Header().Set("WWW-Authenticate", `Basic realm="cangjie-mem API"`)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(map[string]string{
		"error": "Unauthorized",
		"code":  "AUTH_REQUIRED",
	})
}

// cors CORS 中间件包装器
func (s *Server) cors(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 设置 CORS 头
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Expose-Headers", "WWW-Authenticate")

		// 处理预检请求
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// 调用下一个处理器
		next(w, r)
	}
}

// RegisterStatic 注册静态文件服务
func (s *Server) RegisterStatic(mux *http.ServeMux) {
	if s.webFS == nil {
		log.Println("⚠️  Web FS 未提供，跳过 Web UI 注册")
		return
	}

	// 从传入的文件系统中获取 dist 子目录
	distFS, err := fs.Sub(s.webFS, "dist")
	if err != nil {
		log.Printf("❌ 获取嵌入的 dist 目录失败: %v", err)
		return
	}

	// 创建文件服务器
	fileServer := http.FileServer(http.FS(distFS))

	// SPA 路由：对于非文件路径，返回 index.html
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 尝试打开请求的文件
		path := strings.TrimPrefix(r.URL.Path, "/")
		if _, err := distFS.Open(path); err != nil {
			// 文件不存在，对于 SPA 返回 index.html
			r.URL.Path = "/"
		}
		fileServer.ServeHTTP(w, r)
	}))

	log.Println("✓ Web UI 静态文件已注册: / (从嵌入的文件系统)")
}

// sendJSON 发送 JSON 响应
func (s *Server) sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	var resp interface{}
	switch v := data.(type) {
	case *APIResponse:
		resp = v
	default:
		resp = SuccessResponse(data)
	}

	bytes, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal JSON: %v", err)
		return
	}
	w.Write(bytes)
}

// sendError 发送错误响应
func (s *Server) sendError(w http.ResponseWriter, status int, message string, code ...string) {
	s.sendJSON(w, status, ErrorResponse(message, code...))
}

// parseID 从 URL 路径中解析 ID
func parseID(path string) (int64, error) {
	// 路径格式: /api/memories/:id 或 /api/memories/:id/...
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 3 {
		return 0, fmt.Errorf("invalid path format")
	}
	id, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid id format: %w", err)
	}
	return id, nil
}

// === API 处理器 ===

// handleHealth 健康检查
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.sendJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleListMemories 处理记忆列表
func (s *Server) handleListMemories(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// 解析查询参数
	req := types.ListRequest{
		Level:              r.URL.Query().Get("level"),
		LibraryName:        r.URL.Query().Get("library_name"),
		ProjectPathPattern: r.URL.Query().Get("project_path_pattern"),
		LanguageTag:        r.URL.Query().Get("language_tag"),
		OrderBy:            r.URL.Query().Get("order_by"),
		Brief:              r.URL.Query().Get("brief") == "true",
	}

	// 解析 limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 0 {
			s.sendError(w, http.StatusBadRequest, "Invalid limit parameter")
			return
		}
		if limit > 100 {
			limit = 100 // 最大限制
		}
		req.Limit = limit
	}

	// 解析 offset
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			s.sendError(w, http.StatusBadRequest, "Invalid offset parameter")
			return
		}
		req.Offset = offset
	}

	// 调用 store
	resp, err := s.store.ListMemories(req)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to list memories: %v", err))
		return
	}

	s.sendJSON(w, http.StatusOK, resp)
}

// handleCreateMemory 处理创建记忆
func (s *Server) handleCreateMemory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// 解析请求体
	var req types.StoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
		return
	}

	// 验证必填字段
	if req.Level == "" || req.Title == "" || req.Content == "" {
		s.sendError(w, http.StatusBadRequest, "Missing required fields: level, title, and content are required")
		return
	}

	// 验证层级
	if !req.Level.IsValid() {
		s.sendError(w, http.StatusUnprocessableEntity, fmt.Sprintf("Invalid level: %s. Must be one of: language, project, library", req.Level))
		return
	}

	// 验证特定层级的必填字段
	if req.Level == types.LevelLibrary && req.LibraryName == "" {
		s.sendError(w, http.StatusUnprocessableEntity, "library_name is required for library level")
		return
	}
	if req.Level == types.LevelProject && req.ProjectPathPattern == "" {
		s.sendError(w, http.StatusUnprocessableEntity, "project_path_pattern is required for project level")
		return
	}

	// 存储记忆
	resp, err := s.store.StoreMemory(req)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create memory: %v", err))
		return
	}

	// 设置 Location 头
	w.Header().Set("Location", fmt.Sprintf("/api/memories/%d", resp.ID))
	s.sendJSON(w, http.StatusCreated, resp)
}

// handleMemoryDetail 处理单个记忆详情
func (s *Server) handleMemoryDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// 解析 ID
	id, err := parseID(r.URL.Path)
	if err != nil {
		s.sendError(w, http.StatusBadRequest, fmt.Sprintf("Invalid ID: %v", err))
		return
	}

	// 获取记忆
	memory, err := s.store.GetMemory(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "no rows") {
			s.sendError(w, http.StatusNotFound, fmt.Sprintf("Memory not found: id=%d", id))
		} else {
			s.sendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get memory: %v", err))
		}
		return
	}

	s.sendJSON(w, http.StatusOK, memory)
}

// handleUpdateMemory 处理更新记忆
func (s *Server) handleUpdateMemory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		s.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// 解析 ID
	id, err := parseID(r.URL.Path)
	if err != nil {
		s.sendError(w, http.StatusBadRequest, fmt.Sprintf("Invalid ID: %v", err))
		return
	}

	// 解析请求体
	var req types.StoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
		return
	}

	// 验证必填字段（PUT 需要所有字段，PATCH 可以部分更新）
	if r.Method == http.MethodPut {
		if req.Level == "" || req.Title == "" || req.Content == "" {
			s.sendError(w, http.StatusBadRequest, "Missing required fields: level, title, and content are required")
			return
		}
	}

	// 验证层级（如果提供）
	if req.Level != "" && !req.Level.IsValid() {
		s.sendError(w, http.StatusUnprocessableEntity, fmt.Sprintf("Invalid level: %s. Must be one of: language, project, library", req.Level))
		return
	}

	// 更新记忆
	memory, err := s.store.UpdateMemory(id, req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			s.sendError(w, http.StatusNotFound, fmt.Sprintf("Memory not found: id=%d", id))
		} else {
			s.sendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to update memory: %v", err))
		}
		return
	}

	s.sendJSON(w, http.StatusOK, memory)
}

// handleDeleteMemory 处理删除记忆
func (s *Server) handleDeleteMemory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		s.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// 解析 ID
	id, err := parseID(r.URL.Path)
	if err != nil {
		s.sendError(w, http.StatusBadRequest, fmt.Sprintf("Invalid ID: %v", err))
		return
	}

	// 删除记忆
	_, err = s.store.DeleteMemory(types.DeleteRequest{ID: id})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			s.sendError(w, http.StatusNotFound, fmt.Sprintf("Memory not found: id=%d", id))
		} else {
			s.sendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to delete memory: %v", err))
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleSearch 处理搜索
func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// 解析请求体
	var req types.RecallRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
		return
	}

	// 验证查询
	if req.Query == "" {
		s.sendError(w, http.StatusBadRequest, "Query cannot be empty")
		return
	}

	// 执行搜索
	resp, err := s.store.RecallMemories(req)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to search memories: %v", err))
		return
	}

	s.sendJSON(w, http.StatusOK, resp)
}

// handleCategories 处理分类统计
func (s *Server) handleCategories(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// 解析查询参数
	req := types.ListCategoriesRequest{
		LanguageTag: r.URL.Query().Get("language_tag"),
	}

	// 获取分类
	resp, err := s.store.ListCategories(req)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to list categories: %v", err))
		return
	}

	s.sendJSON(w, http.StatusOK, resp)
}
