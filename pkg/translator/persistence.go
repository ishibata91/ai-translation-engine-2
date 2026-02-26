package translator

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	_ "modernc.org/sqlite"
)

type sqlitePersistence struct {
	baseDir string
	dbs     map[string]*sql.DB
	mu      sync.Mutex
}

// NewSqlitePersistence creates a new instance that implements both ResultWriter and ResumeLoader.
// baseDir is the root directory where {PluginName}_translations.db files are stored.
func NewSqlitePersistence(baseDir string) *sqlitePersistence {
	return &sqlitePersistence{
		baseDir: baseDir,
		dbs:     make(map[string]*sql.DB),
	}
}

func (p *sqlitePersistence) getDB(pluginName string) (*sql.DB, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if db, ok := p.dbs[pluginName]; ok {
		return db, nil
	}

	if err := os.MkdirAll(p.baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	dbPath := filepath.Join(p.baseDir, fmt.Sprintf("%s_translations.db", pluginName))
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database %s: %w", dbPath, err)
	}

	// Initialize schema
	if err := p.initSchema(db); err != nil {
		db.Close()
		return nil, err
	}

	p.dbs[pluginName] = db
	return db, nil
}

func (p *sqlitePersistence) initSchema(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS main_translations (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		form_id TEXT,
		record_type TEXT,
		source_text TEXT,
		translated_text TEXT,
		stage_index INTEGER,
		status TEXT,
		error_message TEXT,
		source_plugin TEXT,
		editor_id TEXT,
		parent_form_id TEXT,
		parent_editor_id TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE UNIQUE INDEX IF NOT EXISTS idx_main_translations_form_id ON main_translations(form_id, record_type, stage_index);
	`
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}
	return nil
}

// LoadCachedResults implements ResumeLoader.
func (p *sqlitePersistence) LoadCachedResults(pluginName string, baseDir string) (map[string]TranslationResult, error) {
	// Override baseDir if provided (though we usually use the one from constructor)
	if baseDir != "" {
		p.baseDir = baseDir
	}

	db, err := p.getDB(pluginName)
	if err != nil {
		return nil, err
	}

	query := `SELECT form_id, record_type, source_text, translated_text, stage_index, status, error_message, source_plugin, editor_id, parent_form_id, parent_editor_id FROM main_translations`
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query translations: %w", err)
	}
	defer rows.Close()

	results := make(map[string]TranslationResult)
	for rows.Next() {
		var res TranslationResult
		var stageIndex sql.NullInt64
		err := rows.Scan(
			&res.ID,
			&res.RecordType,
			&res.SourceText,
			&res.TranslatedText,
			&stageIndex,
			&res.Status,
			&res.ErrorMessage,
			&res.SourcePlugin,
			&res.EditorID,
			&res.ParentID,
			&res.ParentEditorID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		if stageIndex.Valid {
			tmp := int(stageIndex.Int64)
			res.Index = &tmp
		}
		results[res.ID] = res
	}

	return results, nil
}

// Write implements ResultWriter.
func (p *sqlitePersistence) Write(result TranslationResult) error {
	db, err := p.getDB(result.SourcePlugin)
	if err != nil {
		return err
	}

	query := `
	INSERT INTO main_translations (
		form_id, record_type, source_text, translated_text, stage_index, status, error_message, source_plugin, editor_id, parent_form_id, parent_editor_id, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	ON CONFLICT(form_id, record_type, stage_index) DO UPDATE SET
		translated_text = excluded.translated_text,
		status = excluded.status,
		error_message = excluded.error_message,
		updated_at = CURRENT_TIMESTAMP;
	`

	var stageIndex any = nil
	if result.Index != nil {
		stageIndex = *result.Index
	}

	_, err = db.Exec(query,
		result.ID,
		result.RecordType,
		result.SourceText,
		result.TranslatedText,
		stageIndex,
		result.Status,
		result.ErrorMessage,
		result.SourcePlugin,
		result.EditorID,
		result.ParentID,
		result.ParentEditorID,
	)

	if err != nil {
		return fmt.Errorf("failed to upsert translation for %s: %w", result.ID, err)
	}

	return nil
}

// Flush implements ResultWriter.
func (p *sqlitePersistence) Flush() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// In this simple implementation, Write is already synchronous.
	// But we could implement batching here if needed.
	return nil
}

// Close closes all open database connections.
func (p *sqlitePersistence) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for name, db := range p.dbs {
		db.Close()
		delete(p.dbs, name)
	}
	return nil
}
