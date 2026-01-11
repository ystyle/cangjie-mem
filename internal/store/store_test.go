//go:build sqlite_fts5
// +build sqlite_fts5

package store

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/ystyle/cangjie-mem/pkg/db"
	"github.com/ystyle/cangjie-mem/pkg/types"
)

// getTestStore 获取测试 Store 实例
func getTestStore(t *testing.T) *Store {
	t.Helper()

	// 在项目目录下创建测试数据库
	testDir := "./test-data"
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	// 使用测试文件名（基于测试名称）
	tmpPath := filepath.Join(testDir, fmt.Sprintf("store-%s.db", t.Name()))

	// 清理旧数据库文件（如果存在）
	os.Remove(tmpPath)

	// 创建数据库实例
	database, err := db.New(db.Config{Path: tmpPath})
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}

	// 测试结束后清理
	t.Cleanup(func() {
		database.Close()
		os.Remove(tmpPath)
		// 如果目录为空，删除目录
		if entries, _ := os.ReadDir(testDir); len(entries) == 0 {
			os.Remove(testDir)
		}
	})

	return New(database)
}

func TestStoreMemory(t *testing.T) {
	store := getTestStore(t)

	tests := []struct {
		name    string
		req     types.StoreRequest
		wantErr bool
	}{
		{
			name: "存储语言级记忆",
			req: types.StoreRequest{
				Level:     types.LevelLanguage,
				Title:     "接口定义",
				Content:   "使用 interface 关键字定义接口",
				Source:    types.SourceManual,
			},
			wantErr: false,
		},
		{
			name: "存储库级记忆（带库名）",
			req: types.StoreRequest{
				Level:       types.LevelLibrary,
				LibraryName: "tang",
				Title:       "Tang 路由",
				Content:     "使用 RouterGroup 配置路由",
				Source:      types.SourceManual,
			},
			wantErr: false,
		},
		{
			name: "存储项目级记忆",
			req: types.StoreRequest{
				Level:              types.LevelProject,
				ProjectPathPattern: "/test/project/*",
				Title:              "项目日志配置",
				Content:            "使用 logback 配置日志",
				Source:             types.SourceManual,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := store.StoreMemory(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("StoreMemory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !resp.Success {
				t.Errorf("StoreMemory() success = %v, want true", resp.Success)
			}
		})
	}
}

func TestRecallMemories(t *testing.T) {
	store := getTestStore(t)

	// 准备专门用于 FTS5 测试的固定数据
	// 使用独特前缀确保搜索结果准确
	initFTS5TestData(t, store)

	tests := []struct {
		name         string
		query        string
		level        string // 显式指定层级
		wantCount    int
		checkTitle   string
	}{
		{
			name:      "search language level: TEST_FTS5",
			query:     "TEST_FTS5",
			level:     "language",
			wantCount: 2, // function definition + variable declaration
			checkTitle: "",
		},
		{
			name:      "search library level: TEST_FTS5",
			query:     "TEST_FTS5",
			level:     "library",
			wantCount: 3, // tang framework + http client + middleware
			checkTitle: "",
		},
		{
			name:      "search project level: TEST_FTS5",
			query:     "TEST_FTS5",
			level:     "project",
			wantCount: 1, // project config
			checkTitle: "",
		},
		{
			name:      "search AND match: TEST_FTS5 tang",
			query:     "TEST_FTS5 tang",
			level:     "library", // 指定 library 层级
			wantCount: 2,         // tang framework + middleware
			checkTitle: "",
		},
		{
			name:      "search AND match: TEST_FTS5 function",
			query:     "TEST_FTS5 function",
			level:     "language", // 指定 language 层级
			wantCount: 1,
			checkTitle: "TEST_FTS5: function definition",
		},
		{
			name:      "search AND match: TEST_FTS5 http",
			query:     "TEST_FTS5 http",
			level:     "library", // 指定 library 层级
			wantCount: 1,
			checkTitle: "TEST_FTS5: http client",
		},
		{
			name:      "search AND match: TEST_FTS5 project",
			query:     "TEST_FTS5 project",
			level:     "project", // 指定 project 层级
			wantCount: 1,
			checkTitle: "TEST_FTS5: project config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := store.RecallMemories(types.RecallRequest{
				Query: tt.query,
				Level: tt.level,
			})
			if err != nil {
				t.Fatalf("RecallMemories() error = %v", err)
			}

			if resp.Total != tt.wantCount {
				t.Errorf("RecallMemories() total = %v, want %v", resp.Total, tt.wantCount)
			}

			// 验证搜索策略已设置
			if resp.SearchStrategy == "" {
				t.Error("RecallMemories() search strategy should be set")
			}

			if tt.checkTitle != "" && len(resp.Results) > 0 {
				// 验证第一条结果的标题
				if resp.Results[0].Title != tt.checkTitle {
					t.Errorf("RecallMemories() first result title = %q, want %q", resp.Results[0].Title, tt.checkTitle)
				}
			}
		})
	}
}

// initFTS5TestData 初始化专门用于 FTS5 全文搜索测试的数据
// 使用独特的前缀 "TEST_FTS5:" 确保搜索结果准确可靠
func initFTS5TestData(t *testing.T, store *Store) {
	ftsTestData := []types.StoreRequest{
		{
			Level:     types.LevelLanguage,
			Title:     "TEST_FTS5: function definition",
			Content:   "How to define functions in programming",
			Source:    types.SourceManual,
		},
		{
			Level:     types.LevelLanguage,
			Title:     "TEST_FTS5: variable declaration",
			Content:   "Variable declaration syntax and examples",
			Source:    types.SourceManual,
		},
		{
			Level:       types.LevelLibrary,
			LibraryName: "tang",
			Title:       "TEST_FTS5: tang framework",
			Content:     "Tang web framework introduction",
			Source:      types.SourceManual,
		},
		{
			Level:       types.LevelLibrary,
			LibraryName: "http-client",
			Title:       "TEST_FTS5: http client",
			Content:     "HTTP client library usage",
			Source:      types.SourceManual,
		},
		{
			Level:       types.LevelLibrary,
			LibraryName: "tang",
			Title:       "TEST_FTS5: middleware",
			Content:     "Middleware pattern in tang",
			Source:      types.SourceManual,
		},
		{
			Level:              types.LevelProject,
			ProjectPathPattern: "/test/*",
			Title:              "TEST_FTS5: project config",
			Content:            "Project configuration files",
			Source:             types.SourceManual,
		},
	}

	// 存储测试数据
	for _, mem := range ftsTestData {
		_, err := store.StoreMemory(mem)
		if err != nil {
			t.Fatalf("failed to init FTS5 test data: %v", err)
		}
	}
}

func TestAutoLevelDetermination(t *testing.T) {
	store := getTestStore(t)

	// 预存测试数据
	memories := []types.StoreRequest{
		{
			Level:     types.LevelLanguage,
			Title:     "Type Definition",
			Content:   "How to define types",
			Source:    types.SourceManual,
		},
		{
			Level:              types.LevelProject,
			ProjectPathPattern: "/test/*",
			Title:              "Project Config",
			Content:            "Our project configuration",
			Source:             types.SourceManual,
		},
	}

	for _, mem := range memories {
		_, err := store.StoreMemory(mem)
		if err != nil {
			t.Fatalf("failed to store test memory: %v", err)
		}
	}

	// 测试语言级关键词判断（包含"定义"关键词）
	resp, err := store.RecallMemories(types.RecallRequest{
		Query: "如何定义类型",
	})
	if err != nil {
		t.Fatalf("RecallMemories() error = %v", err)
	}

	// 验证策略识别为语言级
	if resp.SearchStrategy != "auto_determined_language" {
		t.Errorf("RecallMemories() strategy = %v, want auto_determined_language", resp.SearchStrategy)
	}

	// 测试项目级关键词判断（包含"我项目"关键词 + 项目上下文）
	resp2, err := store.RecallMemories(types.RecallRequest{
		Query:          "我项目的配置",
		ProjectContext: "/test/*",
	})
	if err != nil {
		t.Fatalf("RecallMemories() error = %v", err)
	}

	// 验证策略识别为项目级
	if resp2.SearchStrategy != "auto_determined_project" {
		t.Errorf("RecallMemories() strategy = %v, want auto_determined_project", resp2.SearchStrategy)
	}
}

func TestDefaultValues(t *testing.T) {
	store := getTestStore(t)

	// 测试默认值设置
	tests := []struct {
		name  string
		req   types.ListRequest
		check func(types.ListRequest)
	}{
		{
			name: "默认 language_tag",
			req:  types.ListRequest{},
			check: func(r types.ListRequest) {
				// 内部应该设置默认值为 "cangjie"
			},
		},
		{
			name: "默认 limit",
			req:  types.ListRequest{},
		},
		{
			name: "默认 order_by",
			req:  types.ListRequest{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 测试不会因为缺少默认值而出错
			_, err := store.ListMemories(tt.req)
			if err != nil {
				t.Errorf("ListMemories() with defaults error = %v", err)
			}
		})
	}

	// 测试 RecallMemories 默认值
	_, err := store.RecallMemories(types.RecallRequest{
		Query: "测试",
	})
	if err != nil {
		t.Errorf("RecallMemories() with defaults error = %v", err)
	}
}
