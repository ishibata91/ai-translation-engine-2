package dictionary_builder

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
)

type sqliteDictionaryStore struct {
	db *sql.DB
}

// NewDictionaryStore creates a new instance of DictionaryStore using the provided SQL DB.
func NewDictionaryStore(db *sql.DB) DictionaryStore {
	return &sqliteDictionaryStore{db: db}
}

// SaveTerms inserts new terms or updates existing ones based on the EDID.
// It executes the operations within a single transaction for efficiency.
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
	defer tx.Rollback() // Safe to call even if committed

	query := `
	INSERT INTO dictionary_entries (edid, rec, source, dest, addon)
	VALUES (?, ?, ?, ?, ?)
	ON CONFLICT(edid) DO UPDATE SET
		rec = excluded.rec,
		source = excluded.source,
		dest = excluded.dest,
		addon = excluded.addon;
	`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare upsert statement: %w", err)
	}
	defer stmt.Close()

	for _, term := range terms {
		_, err := stmt.ExecContext(ctx, term.EDID, term.REC, term.Source, term.Dest, term.Addon)
		if err != nil {
			return fmt.Errorf("failed to upsert term %s: %w", term.EDID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
