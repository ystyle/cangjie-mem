package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/ystyle/cangjie-mem/pkg/types"
)

// importPreviewStore 存储导入预览数据（生产环境应使用 Redis 或数据库）
type importPreviewStore struct {
	previews map[string]*importPreviewData
}

type importPreviewData struct {
	Memories []types.StoreRequest
	Expiry   time.Time
}

// 全局预览存储
var importStore = &importPreviewStore{
	previews: make(map[string]*importPreviewData),
}

// cleanupPreviews 定期清理过期的预览数据
func init() {
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			importStore.cleanup()
		}
	}()
}

func (s *importPreviewStore) save(memories []types.StoreRequest) string {
	id := uuid.New().String()
	s.previews[id] = &importPreviewData{
		Memories: memories,
		Expiry:   time.Now().Add(30 * time.Minute),
	}
	return id
}

func (s *importPreviewStore) get(id string) ([]types.StoreRequest, bool) {
	data, ok := s.previews[id]
	if !ok {
		return nil, false
	}
	if time.Now().After(data.Expiry) {
		delete(s.previews, id)
		return nil, false
	}
	return data.Memories, true
}

func (s *importPreviewStore) cleanup() {
	now := time.Now()
	for id, data := range s.previews {
		if now.After(data.Expiry) {
			delete(s.previews, id)
		}
	}
}

// handleExport 处理导出
func (s *Server) handleExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// 解析请求体
	var req types.ExportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
		return
	}

	// 导出记忆
	memories, err := s.store.ExportMemories(req)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to export memories: %v", err))
		return
	}

	// 构建知识包
	pkg := types.KnowledgePackage{
		Version: "1.0",
		Package: types.PackageInfo{
			Name:    "cangjie-mem export",
			Version: time.Now().Format("2006.01.02.150405"),
		},
		Memories: memories,
	}

	// 设置响应头
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"cangjie-mem-%s.json\"", time.Now().Format("20060102-150405")))

	// 返回 JSON
	s.sendJSON(w, http.StatusOK, pkg)
}

// handleImport 处理导入（上传和预览）
func (s *Server) handleImport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// 解析请求体
	var pkg types.KnowledgePackage
	if err := json.NewDecoder(r.Body).Decode(&pkg); err != nil {
		s.sendError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
		return
	}

	// 验证格式
	if pkg.Version == "" {
		s.sendError(w, http.StatusBadRequest, "Invalid package format: missing version")
		return
	}

	if len(pkg.Memories) == 0 {
		s.sendError(w, http.StatusBadRequest, "No memories found in package")
		return
	}

	// 预览导入
	preview, err := s.store.PreviewImport(pkg.Memories)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to preview import: %v", err))
		return
	}

	// 保存预览数据到内存存储
	importID := importStore.save(pkg.Memories)
	preview.ImportID = importID

	s.sendJSON(w, http.StatusOK, preview)
}

// handleImportConfirm 处理导入确认
func (s *Server) handleImportConfirm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.sendError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// 解析请求体
	var req types.ImportConfirmRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.sendError(w, http.StatusBadRequest, fmt.Sprintf("Invalid request body: %v", err))
		return
	}

	// 获取预览数据
	memories, ok := importStore.get(req.ImportID)
	if !ok {
		s.sendError(w, http.StatusNotFound, fmt.Sprintf("Import preview not found or expired: %s", req.ImportID))
		return
	}

	// 执行导入
	result, err := s.store.ImportMemories(memories)
	if err != nil {
		s.sendError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to import memories: %v", err))
		return
	}

	log.Printf("✓ Import completed: %d added, %d updated", result.Added, result.Updated)
	s.sendJSON(w, http.StatusOK, result)
}
