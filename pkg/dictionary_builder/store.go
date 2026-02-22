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

	tx, err := s.beginUpsertTransaction(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback() // Safe to call even if committed

	if err := s.executeUpsertBatch(ctx, tx, terms); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// beginUpsertTransaction starts a new database transaction for batch upsert.
func (s *sqliteDictionaryStore) beginUpsertTransaction(ctx context.Context) (*sql.Tx, error) {
	slog.DebugContext(ctx, "ENTER DictionaryStore.beginUpsertTransaction")

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	return tx, nil
}

// executeUpsertBatch prepares and executes upsert statements for all terms within a transaction.
func (s *sqliteDictionaryStore) executeUpsertBatch(ctx context.Context, tx *sql.Tx, terms []DictTerm) error {
	slog.DebugContext(ctx, "ENTER DictionaryStore.executeUpsertBatch", slog.Int("count", len(terms)))

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

	return nil
}
