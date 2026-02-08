package store

import (
	"fmt"
	"math"
	"path/filepath"
	"strings"
	"time"

	"github.com/ystyle/cangjie-mem/pkg/db"
	"github.com/ystyle/cangjie-mem/pkg/types"
)

// Store 记忆存储
type Store struct {
	db *db.Database
}

// New 创建新的 Store
func New(database *db.Database) *Store {
	return &Store{db: database}
}

// StoreMemory 存储记忆
func (s *Store) StoreMemory(req types.StoreRequest) (*types.StoreResponse, error) {
	return s.db.Store(req)
}

// RecallMemories 智能检索记忆
func (s *Store) RecallMemories(req types.RecallRequest) (*types.RecallResponse, error) {
	// 设置默认值
	if req.LanguageTag == "" {
		req.LanguageTag = "cangjie"
	}
	if req.MaxResults <= 0 {
		req.MaxResults = 10
	}
	if req.MinConfidence <= 0 {
		req.MinConfidence = 0.5
	}

	// 自动判断层级
	var level types.KnowledgeLevel
	var strategy string

	if req.Level != "" {
		// 用户显式指定层级
		level = types.KnowledgeLevel(req.Level)
		if !level.IsValid() {
			return nil, fmt.Errorf("invalid level: %s", req.Level)
		}
		strategy = fmt.Sprintf("user_specified_%s", level)
	} else {
		// 自动判断层级
		level = s.determineLevel(req.Query, req.ProjectContext)
		strategy = fmt.Sprintf("auto_determined_%s", level)
	}

	// 构建查询字符串（空格分隔的 AND 模式）
	ftsQuery := s.buildFTSQuery(req.Query)

	// 执行查询
	results, err := s.db.Recall(ftsQuery, level, req.LanguageTag, req.ProjectContext, req.MaxResults*2)
	if err != nil {
		return nil, fmt.Errorf("failed to recall memories: %w", err)
	}

	// 计算置信度并排序
	for i := range results {
		results[i].Confidence = s.calculateConfidence(results[i], req.Query, req.ProjectContext)

		// 提取匹配的文本片段
		results[i].MatchedText = s.extractMatchedText(results[i].Content, req.Query, 100)
	}

	// 过滤低置信度结果并排序
	filtered := s.filterAndSortResults(results, req.MinConfidence)

	// 限制返回数量
	if len(filtered) > req.MaxResults {
		filtered = filtered[:req.MaxResults]
	}

	// 更新访问计数
	for _, r := range filtered {
		_ = s.db.UpdateAccessCount(r.ID)
	}

	return &types.RecallResponse{
		Total:          len(filtered),
		Results:        filtered,
		SearchStrategy: strategy,
	}, nil
}

// determineLevel 自动判断记忆层级
func (s *Store) determineLevel(query string, projectPath string) types.KnowledgeLevel {
	queryLower := strings.ToLower(query)

	// 1. 项目相关关键词
	projectKeywords := []string{
		"我项目", "这里的", "我们", "当前项目",
		"配置文件", "项目结构", "我们项目",
	}

	for _, kw := range projectKeywords {
		if strings.Contains(queryLower, kw) && projectPath != "" {
			return types.LevelProject
		}
	}

	// 2. 语言级关键词
	langKeywords := []string{
		"语法", "定义", "关键字", "类型", "接口",
		"函数", "变量", "类", "结构体",
		"如何定义", "怎么声明", "语法是什么",
	}

	for _, kw := range langKeywords {
		if strings.Contains(queryLower, kw) {
			return types.LevelLanguage
		}
	}

	// 3. 查询是否涉及特定项目路径
	if projectPath != "" {
		// 检查是否有项目级记忆匹配这个路径
		// 这里简化处理，实际可以查询数据库
	}

	// 4. 默认返回公共库级
	return types.LevelLibrary
}

// calculateConfidence 计算置信度
func (s *Store) calculateConfidence(result types.RecallResult, query string, projectPath string) float64 {
	base := 0.5
	queryLower := strings.ToLower(query)

	// 1. 精确匹配加分
	if strings.Contains(strings.ToLower(result.Title), queryLower) {
		base = 1.0
	} else if strings.Contains(strings.ToLower(result.Content), queryLower) {
		base = 0.9
	}

	// 2. 项目上下文匹配
	if result.Level == types.LevelProject && projectPath != "" {
		if s.matchesProjectPattern(projectPath, result.ProjectPathPattern) {
			base += 0.3
		}
	}

	// 3. 语言级权威性
	if result.Level == types.LevelLanguage {
		base += 0.2
	}

	// 4. 来源可信度
	if result.Source == types.SourceManual {
		base += 0.1
	}

	// 5. 访问热度（轻微影响）
	if result.AccessCount > 10 {
		base += 0.05
	}

	return math.Min(base, 1.0)
}

// matchesProjectPattern 检查项目路径是否匹配模式
func (s *Store) matchesProjectPattern(projectPath, pattern string) bool {
	if pattern == "" {
		return false
	}

	// 简单的 GLOB 匹配
	matched, err := filepath.Match(pattern, projectPath)
	if err != nil {
		return false
	}

	return matched
}

// extractMatchedText 提取匹配的文本片段
func (s *Store) extractMatchedText(content, query string, maxLen int) string {
	if len(content) <= maxLen {
		return content
	}

	queryLower := strings.ToLower(query)
	contentLower := strings.ToLower(content)

	// 查找查询词在内容中的位置
	idx := strings.Index(contentLower, queryLower)
	if idx == -1 {
		// 没找到，返回前 maxLen 个字符
		return content[:maxLen] + "..."
	}

	// 找到了，提取周围的文本
	start := idx - 50
	if start < 0 {
		start = 0
	}
	end := idx + len(query) + 50
	if end > len(content) {
		end = len(content)
	}

	text := content[start:end]
	if start > 0 {
		text = "..." + text
	}
	if end < len(content) {
		text = text + "..."
	}

	return text
}

// filterAndSortResults 过滤并排序结果
func (s *Store) filterAndSortResults(results []types.RecallResult, minConfidence float64) []types.RecallResult {
	// 过滤
	var filtered []types.RecallResult
	for _, r := range results {
		if r.Confidence >= minConfidence {
			filtered = append(filtered, r)
		}
	}

	// 排序（按置信度降序）
	for i := 0; i < len(filtered); i++ {
		for j := i + 1; j < len(filtered); j++ {
			if filtered[j].Confidence > filtered[i].Confidence {
				filtered[i], filtered[j] = filtered[j], filtered[i]
			}
		}
	}

	return filtered
}

// buildFTSQuery 构建 FTS5 查询字符串（空格分隔的 AND 模式）
func (s *Store) buildFTSQuery(query string) string {
	// 按空格分割查询
	words := strings.Fields(query)

	// 如果没有关键词或只有一个，直接返回
	if len(words) <= 1 {
		return query
	}

	// 用空格连接所有关键词（FTS5 默认 AND 模式）
	cleanQuery := strings.Join(words, " ")

	return cleanQuery
}

// ListMemories 列出记忆
func (s *Store) ListMemories(req types.ListRequest) (*types.ListResponse, error) {
	// 设置默认值
	if req.LanguageTag == "" {
		req.LanguageTag = "cangjie"
	}
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.OrderBy == "" {
		req.OrderBy = "created_at"
	}

	return s.db.List(req)
}

// ListCategories 列出所有库和项目分类
func (s *Store) ListCategories(req types.ListCategoriesRequest) (*types.ListCategoriesResponse, error) {
	if req.LanguageTag == "" {
		req.LanguageTag = "cangjie"
	}

	return s.db.ListCategories(req.LanguageTag)
}

// DeleteMemory 删除记忆
func (s *Store) DeleteMemory(req types.DeleteRequest) (*types.DeleteResponse, error) {
	err := s.db.Delete(req.ID)
	if err != nil {
		return &types.DeleteResponse{
			Success: false,
			ID:      req.ID,
			Message: fmt.Sprintf("删除记忆失败: %v", err),
		}, err
	}

	return &types.DeleteResponse{
		Success: true,
		ID:      req.ID,
		Message: "记忆已成功删除",
	}, nil
}

// GetMemory 获取单个记忆
func (s *Store) GetMemory(id int64) (*types.Memory, error) {
	return s.db.GetByID(id)
}

// UpdateMemory 更新记忆
func (s *Store) UpdateMemory(id int64, req types.StoreRequest) (*types.Memory, error) {
	return s.db.Update(id, req)
}

// ExportMemories 导出记忆
func (s *Store) ExportMemories(req types.ExportRequest) ([]types.StoreRequest, error) {
	return s.db.ExportForImport(req)
}

// PreviewImport 预览导入（检测冲突）
func (s *Store) PreviewImport(memories []types.StoreRequest) (*types.ImportPreview, error) {
	// 检测冲突
	conflicts, err := s.db.FindConflicts(memories)
	if err != nil {
		return nil, fmt.Errorf("failed to find conflicts: %w", err)
	}

	// 生成预览 ID
	importID := fmt.Sprintf("import-%d", time.Now().Unix())

	return &types.ImportPreview{
		ImportID:  importID,
		Total:     len(memories),
		ToAdd:     len(memories) - len(conflicts),
		ToUpdate:  len(conflicts),
		Conflicts: conflicts,
	}, nil
}

// ImportMemories 导入记忆
func (s *Store) ImportMemories(memories []types.StoreRequest) (*types.ImportResult, error) {
	return s.db.ImportMemories(memories)
}

// Close 关闭数据库连接
func (s *Store) Close() error {
	return s.db.Close()
}
