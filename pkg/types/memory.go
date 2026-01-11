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
	ProjectPathPattern  string         `json:"project_path_pattern,omitempty"`
	Source              KnowledgeSource `json:"source"`
	Confidence          float64        `json:"confidence"`
	AccessCount         int            `json:"access_count"`
	MatchedText         string         `json:"matched_text,omitempty"` // 匹配的文本片段
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
