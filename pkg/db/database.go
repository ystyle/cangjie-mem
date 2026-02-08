package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/ystyle/cangjie-mem/pkg/types"
)

// Database 数据库实例
type Database struct {
	db *sql.DB
}

// Config 数据库配置
type Config struct {
	Path string // 数据库文件路径
}

// New 创建新的数据库实例
func New(cfg Config) (*Database, error) {
	// 确保数据库文件目录存在
	dbPath := cfg.Path
	if dbPath == "" {
		homeDir, _ := os.UserHomeDir()
		dbPath = filepath.Join(homeDir, ".cangjie-mem", "memory.db")
	}

	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// 打开数据库连接
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 设置连接池参数
	db.SetMaxOpenConns(1) // SQLite 只允许单写
	db.SetMaxIdleConns(1)

	// 初始化数据库结构
	database := &Database{db: db}
	if err := database.init(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return database, nil
}

// init 初始化数据库表结构
func (d *Database) init() error {
	// 先创建基础表结构（不包含 library_name 索引）
	schema := `
	CREATE TABLE IF NOT EXISTS knowledge_base (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		level TEXT NOT NULL CHECK (level IN ('language', 'project', 'library')),
		language_tag TEXT NOT NULL DEFAULT 'cangjie',
		library_name TEXT,
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

	_, err := d.db.Exec(schema)
	if err != nil {
		return err
	}

	// 自动迁移：检查并添加 library_name 字段（兼容老数据库）
	// 迁移函数会负责创建 library_name 索引
	return d.migrateLibraryName()
}

// Store 存储记忆
func (d *Database) Store(req types.StoreRequest) (*types.StoreResponse, error) {
	// 验证层级
	if !req.Level.IsValid() {
		return nil, fmt.Errorf("invalid knowledge level: %s", req.Level)
	}

	// 项目级必须提供项目路径模式
	if req.Level == types.LevelProject && req.ProjectPathPattern == "" {
		return nil, fmt.Errorf("project_path_pattern is required for project level")
	}

	// 设置默认值
	if req.LanguageTag == "" {
		req.LanguageTag = "cangjie"
	}
	if req.Source == "" {
		req.Source = types.SourceManual
	}
	confidence := 1.0
	if req.Source == types.SourceAutoCaptured {
		confidence = 0.7
	}

	// 插入数据
	result, err := d.db.Exec(`
		INSERT INTO knowledge_base (
			level, language_tag, library_name, project_path_pattern,
			title, content, summary, source, confidence
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, req.Level, req.LanguageTag, req.LibraryName, req.ProjectPathPattern,
		req.Title, req.Content, req.Summary, req.Source, confidence)

	if err != nil {
		return nil, fmt.Errorf("failed to insert memory: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return &types.StoreResponse{
		Success: true,
		ID:      id,
		Message: "记忆已成功存储",
	}, nil
}

// Recall 查询记忆（基础查询，不包含智能逻辑）
func (d *Database) Recall(query string, level types.KnowledgeLevel, languageTag string, projectPath string, limit int) ([]types.RecallResult, error) {
	// 构建查询条件
	whereClause := "WHERE language_tag = ?"
	args := []interface{}{languageTag}

	if level.IsValid() {
		whereClause += " AND level = ?"
		args = append(args, level)
	}

	// 项目上下文匹配
	if projectPath != "" {
		whereClause += ` AND (
			project_path_pattern IS NULL
			OR project_path_pattern = ''
			OR project_path_pattern GLOB ?
		)`
		args = append(args, projectPath)
	}

	// 全文搜索
	queryClause := `
		AND id IN (
			SELECT rowid FROM knowledge_base_fts
			WHERE knowledge_base_fts MATCH ?
			ORDER BY bm25(knowledge_base_fts) LIMIT 100
		)
	`
	args = append(args, query)

	sqlQuery := `
		SELECT
			id, level, title, content, summary,
			library_name, project_path_pattern, source,
			access_count, confidence
		FROM knowledge_base
	` + whereClause + queryClause + `
		ORDER BY confidence DESC, access_count DESC
		LIMIT ?
	`
	args = append(args, limit)

	rows, err := d.db.Query(sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query memories: %w", err)
	}
	defer rows.Close()

	var results []types.RecallResult
	for rows.Next() {
		var r types.RecallResult
		var libraryName, pattern, summary sql.NullString

		err := rows.Scan(
			&r.ID, &r.Level, &r.Title, &r.Content, &summary,
			&libraryName, &pattern, &r.Source,
			&r.AccessCount, &r.Confidence,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		if summary.Valid {
			r.Summary = summary.String
		}
		if libraryName.Valid {
			r.LibraryName = libraryName.String
		}
		if pattern.Valid {
			r.ProjectPathPattern = pattern.String
		}

		results = append(results, r)
	}

	return results, nil
}

// UpdateAccessCount 更新访问次数和最后访问时间
func (d *Database) UpdateAccessCount(id int64) error {
	now := time.Now()
	_, err := d.db.Exec(`
		UPDATE knowledge_base
		SET access_count = access_count + 1,
		    last_accessed_at = ?
		WHERE id = ?
	`, now, id)
	return err
}

// GetByID 根据 ID 获取记忆
func (d *Database) GetByID(id int64) (*types.Memory, error) {
	var m types.Memory
	var libraryName, pattern, summary sql.NullString
	var lastAccessed sql.NullTime

	err := d.db.QueryRow(`
		SELECT id, level, language_tag, library_name, project_path_pattern,
		       title, content, summary, source,
		       access_count, confidence, created_at, updated_at, last_accessed_at
		FROM knowledge_base WHERE id = ?
	`, id).Scan(
		&m.ID, &m.Level, &m.LanguageTag, &libraryName, &pattern,
		&m.Title, &m.Content, &summary, &m.Source,
		&m.AccessCount, &m.Confidence, &m.CreatedAt, &m.UpdatedAt, &lastAccessed,
	)

	if err != nil {
		return nil, err
	}

	if libraryName.Valid {
		m.LibraryName = libraryName.String
	}
	if pattern.Valid {
		m.ProjectPathPattern = pattern.String
	}
	if summary.Valid {
		m.Summary = summary.String
	}
	if lastAccessed.Valid {
		m.LastAccessedAt = &lastAccessed.Time
	}

	return &m, nil
}

// Close 关闭数据库连接
func (d *Database) Close() error {
	return d.db.Close()
}

// migrateLibraryName 自动迁移：添加 library_name 字段
func (d *Database) migrateLibraryName() error {
	// 检查 library_name 字段是否存在
	var hasColumn bool
	err := d.db.QueryRow(`
		SELECT COUNT(*) > 0 FROM pragma_table_info('knowledge_base') WHERE name = 'library_name'
	`).Scan(&hasColumn)

	if err != nil {
		// pragma_table_info 查询失败（罕见），记录警告并强制执行迁移
		fmt.Printf("Warning: failed to check library_name column: %v, attempting migration\n", err)
		hasColumn = false // 强制执行迁移
	}

	if !hasColumn {
		// 执行 ALTER TABLE 添加字段
		_, err := d.db.Exec(`ALTER TABLE knowledge_base ADD COLUMN library_name TEXT`)
		if err != nil {
			return fmt.Errorf("failed to add library_name column: %v", err)
		}
		fmt.Println("✓ Migrated database: added library_name column")

		// 创建索引
		_, err = d.db.Exec(`CREATE INDEX IF NOT EXISTS idx_knowledge_library ON knowledge_base(library_name)`)
		if err != nil {
			return fmt.Errorf("failed to create library_name index: %v", err)
		}
		fmt.Println("✓ Migrated database: created library_name index")
	}

	return nil
}

// List 列出记忆（支持筛选和分页）
func (d *Database) List(req types.ListRequest) (*types.ListResponse, error) {
	// 构建动态 WHERE 条件
	whereClause := "WHERE 1=1"
	args := []interface{}{}

	if req.LanguageTag != "" {
		whereClause += " AND language_tag = ?"
		args = append(args, req.LanguageTag)
	} else {
		// 默认筛选 cangjie
		whereClause += " AND language_tag = ?"
		args = append(args, "cangjie")
	}

	if req.Level != "" {
		whereClause += " AND level = ?"
		args = append(args, req.Level)
	}

	if req.LibraryName != "" {
		whereClause += " AND library_name = ?"
		args = append(args, req.LibraryName)
	}

	if req.ProjectPathPattern != "" {
		whereClause += " AND project_path_pattern = ?"
		args = append(args, req.ProjectPathPattern)
	}

	// 查询总数
	var total int
	countQuery := "SELECT COUNT(*) FROM knowledge_base " + whereClause
	err := d.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to count memories: %w", err)
	}

	// 构建排序子句
	orderBy := "created_at DESC"
	if req.OrderBy == "access_count" {
		orderBy = "access_count DESC"
	} else if req.OrderBy == "updated_at" {
		orderBy = "updated_at DESC"
	}

	// 设置默认值
	limit := 20
	if req.Limit > 0 {
		limit = req.Limit
	}
	offset := 0
	if req.Offset > 0 {
		offset = req.Offset
	}

	// 构建查询字段（根据 brief 参数决定是否查询 content）
	var selectFields string
	if req.Brief {
		// 简洁模式：不查询 content 字段
		selectFields = "id, level, language_tag, title, '' as content, summary, library_name, project_path_pattern, source, access_count, confidence, created_at, updated_at"
	} else {
		// 详细模式：查询所有字段
		selectFields = "id, level, language_tag, title, content, summary, library_name, project_path_pattern, source, access_count, confidence, created_at, updated_at"
	}

	// 查询数据
	sqlQuery := `
		SELECT ` + selectFields + `
		FROM knowledge_base
	` + whereClause + `
		ORDER BY ` + orderBy + `
		LIMIT ? OFFSET ?
	`
	args = append(args, limit, offset)

	rows, err := d.db.Query(sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list memories: %w", err)
	}
	defer rows.Close()

	var results []types.Memory
	for rows.Next() {
		var m types.Memory
		var libraryName, pattern, summary sql.NullString
		var languageTag sql.NullString

		err := rows.Scan(
			&m.ID, &m.Level, &languageTag,
			&m.Title, &m.Content, &summary,
			&libraryName, &pattern,
			&m.Source,
			&m.AccessCount, &m.Confidence,
			&m.CreatedAt, &m.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// 处理可为空的字符串字段
		if languageTag.Valid {
			m.LanguageTag = languageTag.String
		} else {
			m.LanguageTag = "cangjie" // 默认值
		}
		if summary.Valid {
			m.Summary = summary.String
		}
		if libraryName.Valid {
			m.LibraryName = libraryName.String
		}
		if pattern.Valid {
			m.ProjectPathPattern = pattern.String
		}

		results = append(results, m)
	}

	return &types.ListResponse{
		Total:   total,
		Results: results,
	}, nil
}

// ListCategories 列出所有库和项目分类
func (d *Database) ListCategories(languageTag string) (*types.ListCategoriesResponse, error) {
	if languageTag == "" {
		languageTag = "cangjie"
	}

	// 查询所有库
	libRows, err := d.db.Query(`
		SELECT library_name, COUNT(*) as count
		FROM knowledge_base
		WHERE level = 'library' AND library_name IS NOT NULL AND library_name != '' AND language_tag = ?
		GROUP BY library_name
		ORDER BY count DESC
	`, languageTag)
	if err != nil {
		return nil, fmt.Errorf("failed to list libraries: %w", err)
	}
	defer libRows.Close()

	var libraries []types.CategoryInfo
	for libRows.Next() {
		var name sql.NullString
		var count int
		if err := libRows.Scan(&name, &count); err != nil {
			return nil, fmt.Errorf("failed to scan library row: %w", err)
		}
		if name.Valid {
			libraries = append(libraries, types.CategoryInfo{
				Name:  name.String,
				Count: count,
			})
		}
	}

	// 查询所有项目
	projectRows, err := d.db.Query(`
		SELECT project_path_pattern, COUNT(*) as count
		FROM knowledge_base
		WHERE level = 'project' AND project_path_pattern IS NOT NULL AND project_path_pattern != '' AND language_tag = ?
		GROUP BY project_path_pattern
		ORDER BY count DESC
	`, languageTag)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	defer projectRows.Close()

	var projects []types.CategoryInfo
	for projectRows.Next() {
		var pattern sql.NullString
		var count int
		if err := projectRows.Scan(&pattern, &count); err != nil {
			return nil, fmt.Errorf("failed to scan project row: %w", err)
		}
		if pattern.Valid {
			projects = append(projects, types.CategoryInfo{
				Name:  pattern.String,
				Count: count,
			})
		}
	}

	return &types.ListCategoriesResponse{
		Libraries: libraries,
		Projects:  projects,
	}, nil
}

// Delete 删除记忆
func (d *Database) Delete(id int64) error {
	result, err := d.db.Exec(`DELETE FROM knowledge_base WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete memory: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("memory not found: id=%d", id)
	}

	return nil
}

// Update 更新记忆
func (d *Database) Update(id int64, req types.StoreRequest) (*types.Memory, error) {
	// 验证层级
	if !req.Level.IsValid() {
		return nil, fmt.Errorf("invalid knowledge level: %s", req.Level)
	}

	// 项目级必须提供项目路径模式
	if req.Level == types.LevelProject && req.ProjectPathPattern == "" {
		return nil, fmt.Errorf("project_path_pattern is required for project level")
	}

	// 设置默认值
	if req.LanguageTag == "" {
		req.LanguageTag = "cangjie"
	}
	if req.Source == "" {
		req.Source = types.SourceManual
	}

	// 执行更新
	_, err := d.db.Exec(`
		UPDATE knowledge_base
		SET level = ?, language_tag = ?, library_name = ?, project_path_pattern = ?,
		    title = ?, content = ?, summary = ?, source = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, req.Level, req.LanguageTag, req.LibraryName, req.ProjectPathPattern,
		req.Title, req.Content, req.Summary, req.Source, id)

	if err != nil {
		return nil, fmt.Errorf("failed to update memory: %w", err)
	}

	// 获取更新后的记录
	return d.GetByID(id)
}

// ExportForImport 导出记忆用于导入（返回 StoreRequest 格式）
func (d *Database) ExportForImport(req types.ExportRequest) ([]types.StoreRequest, error) {
	// 构建查询条件
	whereClause := "WHERE 1=1"
	args := []interface{}{}

	if req.LanguageTag != "" {
		whereClause += " AND language_tag = ?"
		args = append(args, req.LanguageTag)
	} else {
		whereClause += " AND language_tag = ?"
		args = append(args, "cangjie")
	}

	if req.Level != "" {
		whereClause += " AND level = ?"
		args = append(args, req.Level)
	}

	if req.LibraryName != "" {
		whereClause += " AND library_name = ?"
		args = append(args, req.LibraryName)
	}

	if req.ProjectPathPattern != "" {
		whereClause += " AND project_path_pattern = ?"
		args = append(args, req.ProjectPathPattern)
	}

	// 查询数据
	sqlQuery := `
		SELECT level, language_tag, library_name, project_path_pattern,
		       title, content, summary, source
		FROM knowledge_base
	` + whereClause + `
		ORDER BY created_at DESC
	`

	rows, err := d.db.Query(sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to export memories: %w", err)
	}
	defer rows.Close()

	var results []types.StoreRequest
	for rows.Next() {
		var r types.StoreRequest
		var libraryName, pattern, summary sql.NullString
		var source sql.NullString

		err := rows.Scan(
			&r.Level, &r.LanguageTag, &libraryName, &pattern,
			&r.Title, &r.Content, &summary, &source,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		if libraryName.Valid {
			r.LibraryName = libraryName.String
		}
		if pattern.Valid {
			r.ProjectPathPattern = pattern.String
		}
		if summary.Valid {
			r.Summary = summary.String
		}
		if source.Valid {
			r.Source = types.KnowledgeSource(source.String)
		}

		results = append(results, r)
	}

	return results, nil
}

// FindConflicts 查找冲突（同库同标题）
func (d *Database) FindConflicts(memories []types.StoreRequest) ([]types.ConflictInfo, error) {
	var conflicts []types.ConflictInfo

	for _, mem := range memories {
		// 只对 library 层级检测冲突
		if mem.Level != types.LevelLibrary || mem.LibraryName == "" {
			continue
		}

		// 查找已存在的记录
		var id int64
		err := d.db.QueryRow(`
			SELECT id FROM knowledge_base
			WHERE level = ? AND library_name = ? AND title = ?
			LIMIT 1
		`, mem.Level, mem.LibraryName, mem.Title).Scan(&id)

		if err == nil {
			// 找到冲突
			conflicts = append(conflicts, types.ConflictInfo{
				ExistingID:  id,
				Title:       mem.Title,
				LibraryName: mem.LibraryName,
				Level:       mem.Level,
			})
		}
	}

	return conflicts, nil
}

// ImportMemories 导入记忆（支持覆盖）
func (d *Database) ImportMemories(memories []types.StoreRequest) (*types.ImportResult, error) {
	added := 0
	updated := 0

	for _, mem := range memories {
		// 设置默认值
		if mem.LanguageTag == "" {
			mem.LanguageTag = "cangjie"
		}
		if mem.Source == "" {
			mem.Source = types.SourceManual
		}

		// 查找是否已存在（同库同标题）
		var existingID int64
		err := d.db.QueryRow(`
			SELECT id FROM knowledge_base
			WHERE level = ? AND library_name = ? AND title = ?
			LIMIT 1
		`, mem.Level, mem.LibraryName, mem.Title).Scan(&existingID)

		if err == nil {
			// 已存在，更新
			_, err = d.db.Exec(`
				UPDATE knowledge_base
				SET language_tag = ?, project_path_pattern = ?,
				    content = ?, summary = ?, source = ?, updated_at = CURRENT_TIMESTAMP
				WHERE id = ?
			`, mem.LanguageTag, mem.ProjectPathPattern,
				mem.Content, mem.Summary, mem.Source, existingID)

			if err != nil {
				return nil, fmt.Errorf("failed to update memory %s: %w", mem.Title, err)
			}
			updated++
		} else {
			// 不存在，插入
			confidence := 1.0
			if mem.Source == types.SourceAutoCaptured {
				confidence = 0.7
			}

			_, err = d.db.Exec(`
				INSERT INTO knowledge_base (
					level, language_tag, library_name, project_path_pattern,
					title, content, summary, source, confidence
				) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
			`, mem.Level, mem.LanguageTag, mem.LibraryName, mem.ProjectPathPattern,
				mem.Title, mem.Content, mem.Summary, mem.Source, confidence)

			if err != nil {
				return nil, fmt.Errorf("failed to insert memory %s: %w", mem.Title, err)
			}
			added++
		}
	}

	return &types.ImportResult{
		Added:   added,
		Updated: updated,
		Total:   added + updated,
	}, nil
}
