//go:build sqlite_fts5
// +build sqlite_fts5

package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/ystyle/cangjie-mem/pkg/types"
)

// getTestDB 获取测试数据库实例
func getTestDB(t *testing.T) *Database {
	t.Helper()

	// 在项目目录下创建测试数据库
	testDir := "./test-data"
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("failed to create test directory: %v", err)
	}

	// 使用测试文件名（基于测试名称）
	tmpPath := filepath.Join(testDir, fmt.Sprintf("%s.db", t.Name()))

	// 清理旧数据库文件（如果存在）
	os.Remove(tmpPath)

	// 创建数据库实例
	db, err := New(Config{Path: tmpPath})
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}

	// 测试结束后清理
	t.Cleanup(func() {
		db.Close()
		os.Remove(tmpPath)
		// 如果目录为空，删除目录
		if entries, _ := os.ReadDir(testDir); len(entries) == 0 {
			os.Remove(testDir)
		}
	})

	return db
}

func TestNew(t *testing.T) {
	db := getTestDB(t)
	if db == nil {
		t.Fatal("failed to create database")
	}
}

func TestStore(t *testing.T) {
	db := getTestDB(t)

	tests := []struct {
		name    string
		req     types.StoreRequest
		wantErr bool
	}{
		{
			name: "存储语言级记忆",
			req: types.StoreRequest{
				Level:     types.LevelLanguage,
				Title:     "测试标题",
				Content:   "测试内容",
				Source:    types.SourceManual,
			},
			wantErr: false,
		},
		{
			name: "存储库级记忆（带库名）",
			req: types.StoreRequest{
				Level:       types.LevelLibrary,
				LibraryName: "tang",
				Title:       "Tang 路由配置",
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
				Title:              "项目配置",
				Content:            "项目特定配置",
				Source:             types.SourceManual,
			},
			wantErr: false,
		},
		{
			name: "无效层级",
			req: types.StoreRequest{
				Level:   "invalid",
				Title:   "测试",
				Content: "内容",
			},
			wantErr: true,
		},
		{
			name: "项目级缺少路径",
			req: types.StoreRequest{
				Level:   types.LevelProject,
				Title:   "测试",
				Content: "内容",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := db.Store(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Store() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !resp.Success {
				t.Errorf("Store() success = %v, want true", resp.Success)
			}
		})
	}
}

func TestGetByID(t *testing.T) {
	db := getTestDB(t)

	// 先存储一条记忆
	storeResp, err := db.Store(types.StoreRequest{
		Level:       types.LevelLibrary,
		LibraryName: "test-lib",
		Title:       "测试标题",
		Content:     "测试内容",
	})
	if err != nil {
		t.Fatalf("failed to store test memory: %v", err)
	}

	// 测试查询
	memory, err := db.GetByID(storeResp.ID)
	if err != nil {
		t.Fatalf("failed to get memory by ID: %v", err)
	}

	// 验证字段
	if memory.Title != "测试标题" {
		t.Errorf("Title = %v, want 测试标题", memory.Title)
	}
	if memory.Content != "测试内容" {
		t.Errorf("Content = %v, want 测试内容", memory.Content)
	}
	if memory.LibraryName != "test-lib" {
		t.Errorf("LibraryName = %v, want test-lib", memory.LibraryName)
	}
	if memory.Level != types.LevelLibrary {
		t.Errorf("Level = %v, want %v", memory.Level, types.LevelLibrary)
	}
}

func TestList(t *testing.T) {
	db := getTestDB(t)

	// 存储测试数据
	memories := []types.StoreRequest{
		{
			Level:       types.LevelLibrary,
			LibraryName: "tang",
			Title:       "Tang 路由",
			Content:     "路由配置",
		},
		{
			Level:       types.LevelLibrary,
			LibraryName: "tang",
			Title:       "Tang 中间件",
			Content:     "中间件使用",
		},
		{
			Level:       types.LevelLibrary,
			LibraryName: "http-client",
			Title:       "HTTP 请求",
			Content:     "发送请求",
		},
		{
			Level:              types.LevelProject,
			ProjectPathPattern: "/test/*",
			Title:              "项目配置",
			Content:            "配置文件",
		},
	}

	for _, mem := range memories {
		_, err := db.Store(mem)
		if err != nil {
			t.Fatalf("failed to store test memory: %v", err)
		}
	}

	tests := []struct {
		name         string
		req          types.ListRequest
		wantCount    int
		wantTotal    int
		checkResults bool
	}{
		{
			name: "列出所有记忆",
			req:  types.ListRequest{},
			wantTotal: 4,
			wantCount: 4,
		},
		{
			name: "按层级筛选（library）",
			req: types.ListRequest{
				Level: "library",
			},
			wantTotal: 3,
			wantCount: 3,
		},
		{
			name: "按库名筛选（tang）",
			req: types.ListRequest{
				LibraryName: "tang",
			},
			wantTotal: 2,
			wantCount: 2,
			checkResults: true,
		},
		{
			name: "按项目路径筛选",
			req: types.ListRequest{
				ProjectPathPattern: "/test/*",
			},
			wantTotal: 1,
			wantCount: 1,
		},
		{
			name: "组合筛选（library + tang）",
			req: types.ListRequest{
				Level:       "library",
				LibraryName: "tang",
			},
			wantTotal: 2,
			wantCount: 2,
		},
		{
			name: "分页测试",
			req: types.ListRequest{
				Limit: 2,
			},
			wantTotal: 4,
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := db.List(tt.req)
			if err != nil {
				t.Errorf("List() error = %v", err)
				return
			}

			if resp.Total != tt.wantTotal {
				t.Errorf("List() total = %v, want %v", resp.Total, tt.wantTotal)
			}

			if len(resp.Results) != tt.wantCount {
				t.Errorf("List() results count = %v, want %v", len(resp.Results), tt.wantCount)
			}

			// 如果需要验证结果内容
			if tt.checkResults {
				for _, r := range resp.Results {
					if r.LibraryName != "tang" {
						t.Errorf("List() library_name = %v, want tang", r.LibraryName)
					}
				}
			}
		})
	}
}

func TestListCategories(t *testing.T) {
	db := getTestDB(t)

	// 存储测试数据
	memories := []types.StoreRequest{
		{
			Level:       types.LevelLibrary,
			LibraryName: "tang",
			Title:       "Tang 路由",
			Content:     "路由配置",
		},
		{
			Level:       types.LevelLibrary,
			LibraryName: "tang",
			Title:       "Tang 中间件",
			Content:     "中间件使用",
		},
		{
			Level:       types.LevelLibrary,
			LibraryName: "http-client",
			Title:       "HTTP 请求",
			Content:     "发送请求",
		},
		{
			Level:              types.LevelProject,
			ProjectPathPattern: "/test/*",
			Title:              "项目配置",
			Content:            "配置文件",
		},
		{
			Level:              types.LevelProject,
			ProjectPathPattern: "/test/*",
			Title:              "项目日志",
			Content:            "日志配置",
		},
	}

	for _, mem := range memories {
		_, err := db.Store(mem)
		if err != nil {
			t.Fatalf("failed to store test memory: %v", err)
		}
	}

	// 测试列出分类
	resp, err := db.ListCategories("cangjie")
	if err != nil {
		t.Fatalf("ListCategories() error = %v", err)
	}

	// 验证库分类
	if len(resp.Libraries) != 2 {
		t.Errorf("ListCategories() libraries count = %v, want 2", len(resp.Libraries))
	}

	// 验证项目分类
	if len(resp.Projects) != 1 {
		t.Errorf("ListCategories() projects count = %v, want 1", len(resp.Projects))
	}

	// 验证统计数量
	libCounts := make(map[string]int)
	for _, lib := range resp.Libraries {
		libCounts[lib.Name] = lib.Count
	}

	if libCounts["tang"] != 2 {
		t.Errorf("tang library count = %v, want 2", libCounts["tang"])
	}
	if libCounts["http-client"] != 1 {
		t.Errorf("http-client library count = %v, want 1", libCounts["http-client"])
	}

	if resp.Projects[0].Count != 2 {
		t.Errorf("project count = %v, want 2", resp.Projects[0].Count)
	}
}

func TestDelete(t *testing.T) {
	db := getTestDB(t)

	// 存储一条记忆
	storeResp, err := db.Store(types.StoreRequest{
		Level:       types.LevelLibrary,
		LibraryName: "test-lib",
		Title:       "测试标题",
		Content:     "测试内容",
	})
	if err != nil {
		t.Fatalf("failed to store test memory: %v", err)
	}

	// 验证记忆存在
	_, err = db.GetByID(storeResp.ID)
	if err != nil {
		t.Fatalf("memory should exist before delete: %v", err)
	}

	// 删除记忆
	err = db.Delete(storeResp.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// 验证记忆已删除
	_, err = db.GetByID(storeResp.ID)
	if err == nil {
		t.Error("memory should not exist after delete")
	}

	// 删除不存在的记忆
	err = db.Delete(99999)
	if err == nil {
		t.Error("deleting non-existent memory should return error")
	}
}

func TestRecall(t *testing.T) {
	db := getTestDB(t)

	// 存储测试数据
	memories := []types.StoreRequest{
		{
			Level:       types.LevelLibrary,
			LibraryName: "tang",
			Title:       "Tang 路由配置",
			Content:     "使用 RouterGroup 配置路由",
		},
		{
			Level:       types.LevelLibrary,
			LibraryName: "http",
			Title:       "HTTP 客户端",
			Content:     "发送 HTTP 请求",
		},
		{
			Level:       types.LevelLibrary,
			LibraryName: "test",
			Title:       "RouterGroup",
			Content:     "Configuration for routing",
		},
	}

	for _, mem := range memories {
		_, err := db.Store(mem)
		if err != nil {
			t.Fatalf("failed to store test memory: %v", err)
		}
	}

	// 测试英文全文搜索
	results, err := db.Recall("RouterGroup", types.LevelLibrary, "cangjie", "", 10)
	if err != nil {
		t.Fatalf("Recall() error = %v", err)
	}

	if len(results) == 0 {
		t.Error("Recall() should return results for 'RouterGroup'")
	}

	// 验证搜索结果
	found := false
	for _, r := range results {
		if len(r.Content) > 0 && (r.Content == "Configuration for routing" || r.Title == "Tang 路由配置") {
			found = true
			break
		}
	}
	if !found {
		t.Logf("Recall() results: %+v", results)
		t.Error("Recall() results should contain matching memories")
	}
}

func TestUpdateAccessCount(t *testing.T) {
	db := getTestDB(t)

	// 存储一条记忆
	storeResp, err := db.Store(types.StoreRequest{
		Level:       types.LevelLibrary,
		LibraryName: "test-lib",
		Title:       "测试标题",
		Content:     "测试内容",
	})
	if err != nil {
		t.Fatalf("failed to store test memory: %v", err)
	}

	// 获取初始访问次数
	memory, err := db.GetByID(storeResp.ID)
	if err != nil {
		t.Fatalf("failed to get memory: %v", err)
	}
	initialCount := memory.AccessCount

	// 更新访问次数
	err = db.UpdateAccessCount(storeResp.ID)
	if err != nil {
		t.Fatalf("UpdateAccessCount() error = %v", err)
	}

	// 验证访问次数增加
	memory, err = db.GetByID(storeResp.ID)
	if err != nil {
		t.Fatalf("failed to get memory after update: %v", err)
	}

	if memory.AccessCount != initialCount+1 {
		t.Errorf("AccessCount = %v, want %v", memory.AccessCount, initialCount+1)
	}

	if memory.LastAccessedAt == nil {
		t.Error("LastAccessedAt should be set after UpdateAccessCount")
	}
}

func TestMigration(t *testing.T) {
	t.Run("新数据库创建", func(t *testing.T) {
		// 测试自动迁移功能
		db := getTestDB(t)

		// 验证 library_name 字段存在（通过插入带 library_name 的数据）
		_, err := db.Store(types.StoreRequest{
			Level:       types.LevelLibrary,
			LibraryName: "migration-test",
			Title:       "迁移测试",
			Content:     "测试自动迁移",
		})
		if err != nil {
			t.Errorf("Store() with library_name should work after migration: %v", err)
		}

		// 验证可以按库名查询
		listResp, err := db.List(types.ListRequest{
			LibraryName: "migration-test",
		})
		if err != nil {
			t.Fatalf("List() by library_name failed: %v", err)
		}

		if listResp.Total != 1 {
			t.Errorf("List() total = %v, want 1", listResp.Total)
		}
	})

	t.Run("旧数据库迁移", func(t *testing.T) {
		// 创建旧版本的数据库（没有 library_name 字段）
		testDir := "./test-data"
		oldDBPath := filepath.Join(testDir, fmt.Sprintf("%s_old.db", t.Name()))
		if err := os.MkdirAll(testDir, 0755); err != nil {
			t.Fatalf("failed to create test directory: %v", err)
		}

		// 清理旧数据库文件（如果存在）
		os.Remove(oldDBPath)

		// 创建旧版本数据库结构（v1.2.0）
		oldDB, err := sql.Open("sqlite3", oldDBPath+"?_fk=1")
		if err != nil {
			t.Fatalf("failed to create old database: %v", err)
		}

		// 设置单写连接（SQLite 限制）
		oldDB.SetMaxOpenConns(1)
		oldDB.SetMaxIdleConns(1)

		// 创建旧版本的表结构（不包含 library_name 字段）
		oldSchema := `
		CREATE TABLE IF NOT EXISTS knowledge_base (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			level TEXT NOT NULL CHECK (level IN ('language', 'project', 'library')),
			language_tag TEXT NOT NULL DEFAULT 'cangjie',
			project_path_pattern TEXT,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			summary TEXT,
			source TEXT CHECK (source IN ('manual', 'auto_captured')) DEFAULT 'manual',
			access_count INTEGER DEFAULT 0,
			confidence REAL DEFAULT 1.0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			last_accessed_at TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS idx_knowledge_level ON knowledge_base(level);
		CREATE INDEX IF NOT EXISTS idx_knowledge_language ON knowledge_base(language_tag);
		CREATE INDEX IF NOT EXISTS idx_knowledge_project_pattern ON knowledge_base(project_path_pattern);
		CREATE INDEX IF NOT EXISTS idx_knowledge_created_at ON knowledge_base(created_at DESC);

		CREATE VIRTUAL TABLE IF NOT EXISTS knowledge_base_fts USING fts5(
			title,
			content,
			summary,
			content=knowledge_base,
			content_rowid=rowid
		);

		CREATE TRIGGER IF NOT EXISTS knowledge_base_ai AFTER INSERT ON knowledge_base BEGIN
			INSERT INTO knowledge_base_fts(rowid, title, content, summary)
			VALUES (new.id, new.title, new.content, new.summary);
		END;

		CREATE TRIGGER IF NOT EXISTS knowledge_base_ad AFTER DELETE ON knowledge_base BEGIN
			INSERT INTO knowledge_base_fts(knowledge_base_fts, rowid, title, content, summary)
			VALUES ('delete', old.id, old.title, old.content, old.summary);
		END;

		CREATE TRIGGER IF NOT EXISTS knowledge_base_au AFTER UPDATE ON knowledge_base BEGIN
			INSERT INTO knowledge_base_fts(knowledge_base_fts, rowid, title, content, summary)
			VALUES ('delete', old.id, old.title, old.content, old.summary);
			INSERT INTO knowledge_base_fts(rowid, title, content, summary)
			VALUES (new.id, new.title, new.content, new.summary);
		END;
		`
		_, err = oldDB.Exec(oldSchema)
		if err != nil {
			oldDB.Close()
			t.Fatalf("failed to create old schema: %v", err)
		}

		// 插入一些旧数据
		_, err = oldDB.Exec(`
			INSERT INTO knowledge_base (level, title, content)
		 VALUES ('language', '仓颉基础语法', '变量声明的三种方式：var, let, const')
		`)
		if err != nil {
			oldDB.Close()
			t.Fatalf("failed to insert old data: %v", err)
		}

		// 关闭旧数据库
		oldDB.Close()

		// 现在使用新版本的 Database 打开旧数据库
		// 这会触发自动迁移
		newDB, err := New(Config{Path: oldDBPath})
		if err != nil {
			t.Fatalf("failed to open old database with new version: %v", err)
		}
		defer func() {
			newDB.Close()
			os.Remove(oldDBPath)
		}()

		// 验证 1: 可以查询旧数据
		listResp, err := newDB.List(types.ListRequest{})
		if err != nil {
			t.Fatalf("List() failed after migration: %v", err)
		}
		if listResp.Total != 1 {
			t.Errorf("List() total = %v, want 1", listResp.Total)
		}

		// 验证 2: 可以插入带 library_name 的数据
		storeResp, err := newDB.Store(types.StoreRequest{
			Level:       types.LevelLibrary,
			LibraryName: "tang",
			Title:       "Tang 路由",
			Content:     "路由配置",
		})
		if err != nil {
			t.Errorf("Store() with library_name should work after migration: %v", err)
		} else {
			t.Logf("✓ 成功插入带 library_name 的数据, ID=%d", storeResp.ID)
		}

		// 验证 3: 可以按 library_name 筛选
		listResp, err = newDB.List(types.ListRequest{
			LibraryName: "tang",
		})
		if err != nil {
			t.Fatalf("List() by library_name failed: %v", err)
		}
		if listResp.Total != 1 {
			t.Errorf("List() by library_name total = %v, want 1", listResp.Total)
		}

		// 验证 4: 直接查询数据库，确认 library_name 字段存在
		var hasColumn bool
		err = newDB.db.QueryRow(`
			SELECT COUNT(*) > 0 FROM pragma_table_info('knowledge_base') WHERE name = 'library_name'
		`).Scan(&hasColumn)
		if err != nil {
			t.Fatalf("failed to check library_name column: %v", err)
		}
		if !hasColumn {
			t.Error("library_name column should exist after migration")
		} else {
			t.Logf("✓ library_name 字段已成功添加")
		}

		// 验证 5: 确认索引存在
		var indexCount int
		err = newDB.db.QueryRow(`
			SELECT COUNT(*) FROM pragma_index_list('knowledge_base') WHERE name = 'idx_knowledge_library'
		`).Scan(&indexCount)
		if err != nil {
			t.Fatalf("failed to check index: %v", err)
		}
		if indexCount != 1 {
			t.Errorf("idx_knowledge_library index should exist, count=%d", indexCount)
		} else {
			t.Logf("✓ idx_knowledge_library 索引已成功创建")
		}
	})

	t.Run("迁移幂等性", func(t *testing.T) {
		// 测试重复迁移不会报错
		db := getTestDB(t)

		// 第一次迁移已经在 init() 中执行
		// 手动再次调用迁移函数
		err := db.migrateLibraryName()
		if err != nil {
			t.Errorf("migrateLibraryName() should be idempotent, got error: %v", err)
		}

		// 第三次调用
		err = db.migrateLibraryName()
		if err != nil {
			t.Errorf("migrateLibraryName() should be idempotent on repeated calls, got error: %v", err)
		}

		// 验证数据库功能正常
		_, err = db.Store(types.StoreRequest{
			Level:       types.LevelLibrary,
			LibraryName: "idempotent-test",
			Title:       "幂等性测试",
			Content:     "测试重复迁移",
		})
		if err != nil {
			t.Errorf("Store() should work after repeated migrations: %v", err)
		}
	})
}

