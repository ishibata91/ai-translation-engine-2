package dictionary

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"
)

type sqliteDictionaryStore struct {
	db *sql.DB
}

// NewDictionaryStore は提供された SQL DB を使用して DictionaryStore の新しいインスタンスを生成する。
// テーブルが存在しない場合は自動的に作成する。
func NewDictionaryStore(db *sql.DB) (DictionaryStore, error) {
	s := &sqliteDictionaryStore{db: db}
	if err := s.ensureSchema(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ensure dictionary schema: %w", err)
	}
	return s, nil
}

// ensureSchema は必要なテーブルが存在しない場合に作成する。
func (s *sqliteDictionaryStore) ensureSchema(ctx context.Context) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS dlc_sources (
			id           INTEGER PRIMARY KEY AUTOINCREMENT,
			file_name    TEXT NOT NULL,
			format       TEXT NOT NULL DEFAULT 'xml',
			file_path    TEXT NOT NULL,
			file_size    INTEGER NOT NULL DEFAULT 0,
			entry_count  INTEGER NOT NULL DEFAULT 0,
			status       TEXT NOT NULL DEFAULT 'PENDING',
			error_message TEXT,
			imported_at  DATETIME,
			created_at   DATETIME NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS dlc_dictionary_entries (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			source_id   INTEGER NOT NULL REFERENCES dlc_sources(id) ON DELETE CASCADE,
			edid        TEXT NOT NULL,
			record_type TEXT NOT NULL,
			source_text TEXT NOT NULL,
			dest_text   TEXT NOT NULL
		);`,
		`CREATE INDEX IF NOT EXISTS idx_dlc_entries_source_id ON dlc_dictionary_entries(source_id);`,
	}
	for _, q := range queries {
		if _, err := s.db.ExecContext(ctx, q); err != nil {
			return fmt.Errorf("schema error: %w", err)
		}
	}
	return nil
}

// ─── 辞書ソース管理 ───────────────────────────────────────────────────────────

// GetSources は dlc_sources の全レコードを返す。
func (s *sqliteDictionaryStore) GetSources(ctx context.Context) ([]DictSource, error) {
	slog.DebugContext(ctx, "ENTER DictionaryStore.GetSources")
	defer slog.DebugContext(ctx, "EXIT DictionaryStore.GetSources")

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, file_name, format, file_path, file_size, entry_count,
		       status, IFNULL(error_message, ''), imported_at, created_at
		FROM dlc_sources
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query sources: %w", err)
	}
	defer rows.Close()

	var sources []DictSource
	for rows.Next() {
		var src DictSource
		var importedAt sql.NullTime
		err := rows.Scan(
			&src.ID, &src.FileName, &src.Format, &src.FilePath,
			&src.FileSize, &src.EntryCount, &src.Status,
			&src.ErrorMessage, &importedAt, &src.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan source row: %w", err)
		}
		if importedAt.Valid {
			src.ImportedAt = &importedAt.Time
		}
		sources = append(sources, src)
	}
	return sources, rows.Err()
}

// CreateSource は新しい辞書ソースレコードを作成し、採番された ID を返す。
func (s *sqliteDictionaryStore) CreateSource(ctx context.Context, src *DictSource) (int64, error) {
	slog.DebugContext(ctx, "ENTER DictionaryStore.CreateSource", "fileName", src.FileName)
	defer slog.DebugContext(ctx, "EXIT DictionaryStore.CreateSource")

	result, err := s.db.ExecContext(ctx, `
		INSERT INTO dlc_sources (file_name, format, file_path, file_size, entry_count, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, src.FileName, src.Format, src.FilePath, src.FileSize, src.EntryCount, src.Status, time.Now().UTC())
	if err != nil {
		return 0, fmt.Errorf("failed to create source: %w", err)
	}
	return result.LastInsertId()
}

// UpdateSourceStatus は指定ソースのステータス・エントリ数・エラーメッセージを更新する。
func (s *sqliteDictionaryStore) UpdateSourceStatus(ctx context.Context, id int64, status string, count int, errMsg string) error {
	slog.DebugContext(ctx, "ENTER DictionaryStore.UpdateSourceStatus", "id", id, "status", status)
	defer slog.DebugContext(ctx, "EXIT DictionaryStore.UpdateSourceStatus")

	var importedAt interface{}
	if status == "COMPLETED" {
		now := time.Now().UTC()
		importedAt = now
	}

	_, err := s.db.ExecContext(ctx, `
		UPDATE dlc_sources
		SET status = ?, entry_count = ?, error_message = ?, imported_at = ?
		WHERE id = ?
	`, status, count, errMsg, importedAt, id)
	if err != nil {
		return fmt.Errorf("failed to update source status: %w", err)
	}
	return nil
}

// DeleteSource は指定ソースを削除する（dlc_dictionary_entries はカスケード削除）。
func (s *sqliteDictionaryStore) DeleteSource(ctx context.Context, id int64) error {
	slog.DebugContext(ctx, "ENTER DictionaryStore.DeleteSource", "id", id)
	defer slog.DebugContext(ctx, "EXIT DictionaryStore.DeleteSource")

	_, err := s.db.ExecContext(ctx, `DELETE FROM dlc_sources WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete source: %w", err)
	}
	return nil
}

// ─── 辞書エントリ管理 ─────────────────────────────────────────────────────────

// GetEntriesBySourceID は指定ソースに紐付く全エントリを返す。
func (s *sqliteDictionaryStore) GetEntriesBySourceID(ctx context.Context, sourceID int64) ([]DictTerm, error) {
	slog.DebugContext(ctx, "ENTER DictionaryStore.GetEntriesBySourceID", "sourceID", sourceID)
	defer slog.DebugContext(ctx, "EXIT DictionaryStore.GetEntriesBySourceID")

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, source_id, edid, record_type, source_text, dest_text
		FROM dlc_dictionary_entries
		WHERE source_id = ?
		ORDER BY id
	`, sourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to query entries: %w", err)
	}
	defer rows.Close()

	var entries []DictTerm
	for rows.Next() {
		var e DictTerm
		if err := rows.Scan(&e.ID, &e.SourceID, &e.EDID, &e.RecordType, &e.Source, &e.Dest); err != nil {
			return nil, fmt.Errorf("failed to scan entry row: %w", err)
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

// SaveTerms は複数エントリをバッチで挿入する。
func (s *sqliteDictionaryStore) SaveTerms(ctx context.Context, terms []DictTerm) error {
	slog.DebugContext(ctx, "ENTER DictionaryStore.SaveTerms", "count", len(terms))
	defer slog.DebugContext(ctx, "EXIT DictionaryStore.SaveTerms")

	if len(terms) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO dlc_dictionary_entries (source_id, edid, record_type, source_text, dest_text)
		VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	defer stmt.Close()

	for _, term := range terms {
		if _, err := stmt.ExecContext(ctx, term.SourceID, term.EDID, term.RecordType, term.Source, term.Dest); err != nil {
			return fmt.Errorf("failed to insert term %s: %w", term.EDID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

// UpdateEntry は指定エントリの source_text / dest_text を更新する。
func (s *sqliteDictionaryStore) UpdateEntry(ctx context.Context, term DictTerm) error {
	slog.DebugContext(ctx, "ENTER DictionaryStore.UpdateEntry", "id", term.ID)
	defer slog.DebugContext(ctx, "EXIT DictionaryStore.UpdateEntry")

	_, err := s.db.ExecContext(ctx, `
		UPDATE dlc_dictionary_entries
		SET source_text = ?, dest_text = ?
		WHERE id = ?
	`, term.Source, term.Dest, term.ID)
	if err != nil {
		return fmt.Errorf("failed to update entry: %w", err)
	}
	return nil
}

// DeleteEntry は指定エントリを削除する。
func (s *sqliteDictionaryStore) DeleteEntry(ctx context.Context, id int64) error {
	slog.DebugContext(ctx, "ENTER DictionaryStore.DeleteEntry", "id", id)
	defer slog.DebugContext(ctx, "EXIT DictionaryStore.DeleteEntry")

	_, err := s.db.ExecContext(ctx, `DELETE FROM dlc_dictionary_entries WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete entry: %w", err)
	}
	return nil
}
