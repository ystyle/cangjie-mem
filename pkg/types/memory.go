package types

import "time"

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
	SourceManual       KnowledgeSource = "manual"        // 手动录入
	SourceAutoCaptured KnowledgeSource = "auto_captured" // 自动捕获
)

// Memory 记忆条目
type Memory struct {
	ID                 int64            `json:"id"`
	Level              KnowledgeLevel   `json:"level"`
	LanguageTag        string           `json:"language_tag"`
	LibraryName        string           `json:"library_name,omitempty"`
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
	Level              KnowledgeLevel  `json:"level" mcp:"required"`
	LanguageTag        string          `json:"language_tag"`
	LibraryName        string          `json:"library_name,omitempty"`
	ProjectPathPattern string          `json:"project_path_pattern,omitempty"`
	Title              string          `json:"title" mcp:"required"`
	Content            string          `json:"content" mcp:"required"`
	Summary            string          `json:"summary,omitempty"`
	Source             KnowledgeSource `json:"source"`
}

// StoreResponse 存储响应
type StoreResponse struct {
	Success bool   `json:"success"`
	ID      int64  `json:"id"`
	Message string `json:"message"`
}

// RecallRequest 回忆请求
type RecallRequest struct {
	Query          string  `json:"query" mcp:"required"`
	Level          string  `json:"level,omitempty"` // 空字符串表示自动判断
	LanguageTag    string  `json:"language_tag"`
	LibraryName    string  `json:"library_name,omitempty"` // 库名筛选（仅对 library 层级有效）
	ProjectContext string  `json:"project_context,omitempty"`
	MaxResults     int     `json:"max_results"`
	MinConfidence  float64 `json:"min_confidence"`
}

// RecallResult 回忆结果
type RecallResult struct {
	ID                  int64          `json:"id"`
	Level               KnowledgeLevel `json:"level"`
	Title               string         `json:"title"`
	Content             string         `json:"content"`
	Summary             string         `json:"summary,omitempty"`
	LibraryName         string         `json:"library_name,omitempty"`
	ProjectPathPattern  string         `json:"project_path_pattern,omitempty"`
	Source              KnowledgeSource `json:"source"`
	Confidence          float64        `json:"confidence"`
	AccessCount         int            `json:"access_count"`
	MatchedText         string         `json:"matched_text,omitempty"` // 匹配的文本片段
	CreatedAt           string         `json:"created_at,omitempty"`   // 创建时间
	UpdatedAt           string         `json:"updated_at,omitempty"`   // 更新时间
}

// RecallResponse 回忆响应
type RecallResponse struct {
	Total          int            `json:"total"`
	Results        []RecallResult `json:"results"`
	SearchStrategy string         `json:"search_strategy"` // 使用的检索策略
}

// SuggestRequest 建议补充请求
type SuggestRequest struct {
	Query           string         `json:"query" mcp:"required"`
	SuggestedTitle  string         `json:"suggested_title" mcp:"required"`
	SuggestedContent string        `json:"suggested_content" mcp:"required"`
	SuggestedLevel  KnowledgeLevel `json:"suggested_level"`
	Reason          string         `json:"reason"`
}

// SuggestResponse 建议补充响应
type SuggestResponse struct {
	Success       bool   `json:"success"`
	SuggestionID  int64  `json:"suggestion_id"`
	Status        string `json:"status"` // pending_review, approved, rejected
	Message       string `json:"message"`
}

// IsValid 验证记忆层级是否有效
func (l KnowledgeLevel) IsValid() bool {
	switch l {
	case LevelLanguage, LevelProject, LevelLibrary:
		return true
	default:
		return false
	}
}

// IsValid 验证记忆来源是否有效
func (s KnowledgeSource) IsValid() bool {
	switch s {
	case SourceManual, SourceAutoCaptured:
		return true
	default:
		return false
	}
}

// ListRequest 列出请求
type ListRequest struct {
	Level              string `json:"level,omitempty"`                // 可选：language/project/library
	LibraryName        string `json:"library_name,omitempty"`         // 可选：库名筛选
	ProjectPathPattern string `json:"project_path_pattern,omitempty"` // 可选：项目路径筛选
	LanguageTag        string `json:"language_tag,omitempty"`         // 可选：语言标签
	Limit              int    `json:"limit,omitempty"`                // 可选：返回数量，默认20
	Offset             int    `json:"offset,omitempty"`               // 可选：分页偏移
	OrderBy            string `json:"order_by,omitempty"`             // 可选：排序字段
	Brief              bool   `json:"brief,omitempty"`                // 可选：简洁模式，默认false。true时仅返回标题和摘要，不返回完整内容
}

// ListResponse 列出响应
type ListResponse struct {
	Total   int      `json:"total"`
	Results []Memory `json:"results"`
}

// CategoryInfo 分类信息
type CategoryInfo struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// ListCategoriesRequest 分类列表请求
type ListCategoriesRequest struct {
	LanguageTag string `json:"language_tag,omitempty"` // 可选：按语言筛选
}

// ListCategoriesResponse 分类列表响应
type ListCategoriesResponse struct {
	Libraries []CategoryInfo `json:"libraries"` // 所有库及其记忆数
	Projects  []CategoryInfo `json:"projects"`  // 所有项目及其记忆数
}

// DeleteRequest 删除请求
type DeleteRequest struct {
	ID int64 `json:"id" mcp:"required"`
}

// DeleteResponse 删除响应
type DeleteResponse struct {
	Success bool   `json:"success"`
	ID      int64  `json:"id"`
	Message string `json:"message"`
}

// PackageInfo 知识包信息
type PackageInfo struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Author      string   `json:"author,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Version     string   `json:"version"`
}

// KnowledgePackage 知识包（导入/导出格式）
type KnowledgePackage struct {
	Version  string            `json:"version"`  // 格式版本
	Package  PackageInfo       `json:"package"`  // 包信息
	Memories []StoreRequest    `json:"memories"` // 记忆列表
}

// ExportRequest 导出请求
type ExportRequest struct {
	Level              string `json:"level,omitempty"`
	LibraryName        string `json:"library_name,omitempty"`
	ProjectPathPattern string `json:"project_path_pattern,omitempty"`
	LanguageTag        string `json:"language_tag,omitempty"`
}

// ImportPreview 导入预览
type ImportPreview struct {
	ImportID  string         `json:"import_id"`  // 预览 ID
	Total     int            `json:"total"`      // 总记忆数
	ToAdd     int            `json:"to_add"`     // 将新增
	ToUpdate  int            `json:"to_update"`  // 将更新
	Conflicts []ConflictInfo `json:"conflicts"`  // 冲突列表
}

// ConflictInfo 冲突信息
type ConflictInfo struct {
	ExistingID  int64          `json:"existing_id"`  // 已存在记录的 ID
	Title       string         `json:"title"`        // 标题
	LibraryName string         `json:"library_name"` // 库名
	Level       KnowledgeLevel `json:"level"`        // 层级
}

// ImportConfirmRequest 导入确认请求
type ImportConfirmRequest struct {
	ImportID string `json:"import_id"` // 预览 ID
}

// ImportResult 导入结果
type ImportResult struct {
	Added   int `json:"added"`   // 新增数量
	Updated int `json:"updated"` // 更新数量
	Total   int `json:"total"`   // 总数
}
