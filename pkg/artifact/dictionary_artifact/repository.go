package dictionaryartifact

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type sqliteRepository struct {
	db *sql.DB
}

// NewRepository creates a SQLite-backed dictionary artifact repository.
func NewRepository(db *sql.DB) Repository {
	return &sqliteRepository{db: db}
}

func (r *sqliteRepository) GetSources(ctx context.Context) ([]Source, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, file_name, format, file_path, file_size, entry_count,
		       status, IFNULL(error_message, ''), imported_at, created_at
		FROM artifact_dictionary_sources
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("query dictionary sources: %w", err)
	}
	defer rows.Close()

	sources := make([]Source, 0)
	for rows.Next() {
		var source Source
		var importedAt sql.NullTime
		if err := rows.Scan(
			&source.ID, &source.FileName, &source.Format, &source.FilePath,
			&source.FileSize, &source.EntryCount, &source.Status,
			&source.ErrorMessage, &importedAt, &source.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan dictionary source row: %w", err)
		}
		if importedAt.Valid {
			source.ImportedAt = &importedAt.Time
		}
		sources = append(sources, source)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate dictionary source rows: %w", err)
	}
	return sources, nil
}

func (r *sqliteRepository) CreateSource(ctx context.Context, source *Source) (int64, error) {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO artifact_dictionary_sources (file_name, format, file_path, file_size, entry_count, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, source.FileName, source.Format, source.FilePath, source.FileSize, source.EntryCount, source.Status, time.Now().UTC())
	if err != nil {
		return 0, fmt.Errorf("create dictionary source: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("resolve dictionary source insert id: %w", err)
	}
	return id, nil
}

func (r *sqliteRepository) UpdateSourceStatus(ctx context.Context, id int64, status string, count int, errMsg string) error {
	var importedAt any
	if status == "COMPLETED" {
		importedAt = time.Now().UTC()
	}
	if _, err := r.db.ExecContext(ctx, `
		UPDATE artifact_dictionary_sources
		SET status = ?, entry_count = ?, error_message = ?, imported_at = ?
		WHERE id = ?
	`, status, count, errMsg, importedAt, id); err != nil {
		return fmt.Errorf("update dictionary source status id=%d: %w", id, err)
	}
	return nil
}

func (r *sqliteRepository) DeleteSource(ctx context.Context, id int64) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM artifact_dictionary_sources WHERE id = ?`, id); err != nil {
		return fmt.Errorf("delete dictionary source id=%d: %w", id, err)
	}
	return nil
}

func (r *sqliteRepository) FindExactBySourceText(ctx context.Context, text string) ([]Entry, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, source_id, edid, record_type, source_text, dest_text
		FROM artifact_dictionary_entries
		WHERE source_text = ?
	`, text)
	if err != nil {
		return nil, fmt.Errorf("find exact dictionary entry source_text=%q: %w", text, err)
	}
	defer rows.Close()

	entries := make([]Entry, 0)
	for rows.Next() {
		var entry Entry
		if err := rows.Scan(&entry.ID, &entry.SourceID, &entry.EDID, &entry.RecordType, &entry.SourceText, &entry.DestText); err != nil {
			return nil, fmt.Errorf("scan exact dictionary entry row source_text=%q: %w", text, err)
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate exact dictionary entry rows source_text=%q: %w", text, err)
	}
	return entries, nil
}

func (r *sqliteRepository) FindExactBySourceTextCI(ctx context.Context, text string) ([]Entry, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, source_id, edid, record_type, source_text, dest_text
		FROM artifact_dictionary_entries
		WHERE lower(source_text) = lower(?)
	`, text)
	if err != nil {
		return nil, fmt.Errorf("find case-insensitive exact dictionary entry source_text=%q: %w", text, err)
	}
	defer rows.Close()

	entries := make([]Entry, 0)
	for rows.Next() {
		var entry Entry
		if err := rows.Scan(&entry.ID, &entry.SourceID, &entry.EDID, &entry.RecordType, &entry.SourceText, &entry.DestText); err != nil {
			return nil, fmt.Errorf("scan case-insensitive exact dictionary entry row source_text=%q: %w", text, err)
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate case-insensitive exact dictionary entry rows source_text=%q: %w", text, err)
	}
	return entries, nil
}

func (r *sqliteRepository) FindExactBySourceTexts(ctx context.Context, texts []string) ([]Entry, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	placeholders := make([]string, 0, len(texts))
	args := make([]any, 0, len(texts))
	for _, text := range texts {
		placeholders = append(placeholders, "?")
		args = append(args, text)
	}

	query := `
		SELECT id, source_id, edid, record_type, source_text, dest_text
		FROM artifact_dictionary_entries
		WHERE source_text IN (` + strings.Join(placeholders, ", ") + `)
	`
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("find exact dictionary entries by texts: %w", err)
	}
	defer rows.Close()

	entries := make([]Entry, 0)
	for rows.Next() {
		var entry Entry
		if err := rows.Scan(&entry.ID, &entry.SourceID, &entry.EDID, &entry.RecordType, &entry.SourceText, &entry.DestText); err != nil {
			return nil, fmt.Errorf("scan exact dictionary entries by texts: %w", err)
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate exact dictionary entries by texts: %w", err)
	}
	return entries, nil
}

func (r *sqliteRepository) SearchBySourceTextLike(ctx context.Context, keyword string, limit int, npcOnly bool) ([]Entry, error) {
	trimmed := strings.TrimSpace(keyword)
	if trimmed == "" {
		return nil, nil
	}
	if limit <= 0 {
		limit = 1
	}

	query := `
		SELECT id, source_id, edid, record_type, source_text, dest_text
		FROM artifact_dictionary_entries
		WHERE source_text LIKE ?
		LIMIT ?
	`
	args := []any{"%" + trimmed + "%", limit}
	if npcOnly {
		query = `
			SELECT id, source_id, edid, record_type, source_text, dest_text
			FROM artifact_dictionary_entries
			WHERE source_text LIKE ? AND record_type LIKE 'NPC_%'
			LIMIT ?
		`
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("search dictionary entries by like keyword=%q: %w", trimmed, err)
	}
	defer rows.Close()

	entries := make([]Entry, 0)
	for rows.Next() {
		var entry Entry
		if err := rows.Scan(&entry.ID, &entry.SourceID, &entry.EDID, &entry.RecordType, &entry.SourceText, &entry.DestText); err != nil {
			return nil, fmt.Errorf("scan dictionary entries by like keyword=%q: %w", trimmed, err)
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate dictionary entries by like keyword=%q: %w", trimmed, err)
	}
	return entries, nil
}

func (r *sqliteRepository) GetEntriesBySourceID(ctx context.Context, sourceID int64) ([]Entry, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, source_id, edid, record_type, source_text, dest_text
		FROM artifact_dictionary_entries
		WHERE source_id = ?
		ORDER BY id
	`, sourceID)
	if err != nil {
		return nil, fmt.Errorf("query dictionary entries by source id=%d: %w", sourceID, err)
	}
	defer rows.Close()

	entries := make([]Entry, 0)
	for rows.Next() {
		var entry Entry
		if err := rows.Scan(&entry.ID, &entry.SourceID, &entry.EDID, &entry.RecordType, &entry.SourceText, &entry.DestText); err != nil {
			return nil, fmt.Errorf("scan dictionary entry row: %w", err)
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate dictionary entry rows: %w", err)
	}
	return entries, nil
}

func (r *sqliteRepository) GetEntriesBySourceIDPaginated(ctx context.Context, sourceID int64, query string, filters map[string]string, limit int, offset int) (*EntryPage, error) {
	whereClause, args := buildMapSearchWhere(query, filters, "")
	countQuery := `SELECT COUNT(*) FROM artifact_dictionary_entries WHERE source_id = ?` + whereClause
	countArgs := append([]any{sourceID}, args...)

	var totalCount int
	if err := r.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&totalCount); err != nil {
		return nil, fmt.Errorf("count dictionary entries source id=%d: %w", sourceID, err)
	}

	//nolint:gosec // query fragments are generated from fixed columns and placeholders only.
	queryStr := `
		SELECT id, source_id, edid, record_type, source_text, dest_text
		FROM artifact_dictionary_entries
		WHERE source_id = ?` + whereClause + `
		ORDER BY id
		LIMIT ? OFFSET ?
	`
	queryArgs := append(append([]any{}, countArgs...), limit, offset)
	rows, err := r.db.QueryContext(ctx, queryStr, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("query dictionary entries paginated source id=%d: %w", sourceID, err)
	}
	defer rows.Close()

	entries := make([]Entry, 0)
	for rows.Next() {
		var entry Entry
		if err := rows.Scan(&entry.ID, &entry.SourceID, &entry.EDID, &entry.RecordType, &entry.SourceText, &entry.DestText); err != nil {
			return nil, fmt.Errorf("scan dictionary entry row: %w", err)
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate dictionary entries paginated source id=%d: %w", sourceID, err)
	}
	return &EntryPage{Entries: entries, TotalCount: totalCount}, nil
}

func (r *sqliteRepository) SearchAllEntriesPaginated(ctx context.Context, query string, filters map[string]string, limit int, offset int) (*EntryPage, error) {
	whereClause, args := buildMapSearchWhere(query, filters, "")
	countQuery := `SELECT COUNT(*) FROM artifact_dictionary_entries`
	if whereClause != "" {
		countQuery += ` WHERE ` + strings.TrimPrefix(whereClause, " AND ")
	}

	var totalCount int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount); err != nil {
		return nil, fmt.Errorf("count dictionary entries all sources: %w", err)
	}

	whereClauseWithPrefix, argsWithPrefix := buildMapSearchWhere(query, filters, "e.")
	queryStr := `
		SELECT e.id, e.source_id, s.file_name, e.edid, e.record_type, e.source_text, e.dest_text
		FROM artifact_dictionary_entries e
		JOIN artifact_dictionary_sources s ON s.id = e.source_id
	`
	if whereClauseWithPrefix != "" {
		queryStr += ` WHERE ` + strings.TrimPrefix(whereClauseWithPrefix, " AND ")
	}
	queryStr += `
		ORDER BY e.source_id, e.id
		LIMIT ? OFFSET ?
	`

	queryArgs := append(append([]any{}, argsWithPrefix...), limit, offset)
	rows, err := r.db.QueryContext(ctx, queryStr, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("query dictionary entries all sources: %w", err)
	}
	defer rows.Close()

	entries := make([]Entry, 0)
	for rows.Next() {
		var entry Entry
		if err := rows.Scan(&entry.ID, &entry.SourceID, &entry.SourceName, &entry.EDID, &entry.RecordType, &entry.SourceText, &entry.DestText); err != nil {
			return nil, fmt.Errorf("scan dictionary entry row: %w", err)
		}
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate dictionary entries all sources: %w", err)
	}
	return &EntryPage{Entries: entries, TotalCount: totalCount}, nil
}

func (r *sqliteRepository) SaveEntries(ctx context.Context, entries []Entry) error {
	if len(entries) == 0 {
		return nil
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin dictionary entry transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO artifact_dictionary_entries (source_id, edid, record_type, source_text, dest_text)
		VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("prepare dictionary entry insert: %w", err)
	}
	defer stmt.Close()

	for _, entry := range entries {
		if _, err := stmt.ExecContext(ctx, entry.SourceID, entry.EDID, entry.RecordType, entry.SourceText, entry.DestText); err != nil {
			return fmt.Errorf("insert dictionary entry edid=%s: %w", entry.EDID, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit dictionary entry transaction: %w", err)
	}
	return nil
}

func (r *sqliteRepository) UpdateEntry(ctx context.Context, entry Entry) error {
	if _, err := r.db.ExecContext(ctx, `
		UPDATE artifact_dictionary_entries
		SET source_text = ?, dest_text = ?
		WHERE id = ?
	`, entry.SourceText, entry.DestText, entry.ID); err != nil {
		return fmt.Errorf("update dictionary entry id=%d: %w", entry.ID, err)
	}
	return nil
}

func (r *sqliteRepository) DeleteEntry(ctx context.Context, id int64) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM artifact_dictionary_entries WHERE id = ?`, id); err != nil {
		return fmt.Errorf("delete dictionary entry id=%d: %w", id, err)
	}
	return nil
}

func buildMapSearchWhere(query string, filters map[string]string, prefix string) (string, []any) {
	conditions := make([]string, 0)
	args := make([]any, 0)

	if query != "" {
		for _, keyword := range strings.Fields(query) {
			conditions = append(conditions, fmt.Sprintf("(%[1]ssource_text LIKE ? OR %[1]sdest_text LIKE ? OR %[1]sedid LIKE ? OR %[1]srecord_type LIKE ?)", prefix))
			pattern := "%" + keyword + "%"
			args = append(args, pattern, pattern, pattern, pattern)
		}
	}

	columnMap := map[string]string{
		"edid":       "edid",
		"recordType": "record_type",
		"sourceText": "source_text",
		"destText":   "dest_text",
	}
	for key, value := range filters {
		dbColumn, ok := columnMap[key]
		if !ok {
			continue
		}
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		for _, keyword := range strings.Fields(trimmed) {
			conditions = append(conditions, fmt.Sprintf("%[1]s%[2]s LIKE ?", prefix, dbColumn))
			args = append(args, "%"+keyword+"%")
		}
	}

	if len(conditions) == 0 {
		return "", nil
	}
	return " AND (" + strings.Join(conditions, " AND ") + ")", args
}
