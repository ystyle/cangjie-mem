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
	schema := `
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

	_, err := d.db.Exec(schema)
	return err
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
			level, language_tag, project_path_pattern,
			title, content, summary, source, confidence
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, req.Level, req.LanguageTag, req.ProjectPathPattern,
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
			project_path_pattern, source,
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
		var pattern, summary sql.NullString

		err := rows.Scan(
			&r.ID, &r.Level, &r.Title, &r.Content, &summary,
			&pattern, &r.Source,
			&r.AccessCount, &r.Confidence,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		if summary.Valid {
			r.Summary = summary.String
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
	var pattern, summary sql.NullString
	var lastAccessed sql.NullTime

	err := d.db.QueryRow(`
		SELECT id, level, language_tag, project_path_pattern,
		       title, content, summary, source,
		       access_count, confidence, created_at, updated_at, last_accessed_at
		FROM knowledge_base WHERE id = ?
	`, id).Scan(
		&m.ID, &m.Level, &m.LanguageTag, &pattern,
		&m.Title, &m.Content, &summary, &m.Source,
		&m.AccessCount, &m.Confidence, &m.CreatedAt, &m.UpdatedAt, &lastAccessed,
	)

	if err != nil {
		return nil, err
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
