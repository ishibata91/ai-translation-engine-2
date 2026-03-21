package terminology

import (
	"context"
	"database/sql"
	"io"
	"log/slog"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func TestGetPreviewTranslations_FallbackWithSingleConnection(t *testing.T) {
	t.Parallel()

	db, err := sql.Open("sqlite", "file:terminology_fallback_test?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	store := NewSQLiteModTermStore(db, slog.New(slog.NewTextHandler(io.Discard, nil)))

	if err := store.ensureModTable(ctx, modTableName("alpha.json")); err != nil {
		t.Fatalf("failed to create alpha table: %v", err)
	}
	if err := store.ensureModTable(ctx, modTableName("beta.json")); err != nil {
		t.Fatalf("failed to create beta table: %v", err)
	}
	if err := store.upsertTerms(ctx, modTableName("beta.json"), []TermTranslationResult{
		{
			SourceText:     "Hello",
			RecordType:     "BOOK:FULL",
			TranslatedText: "こんにちは",
			Status:         "completed",
			SourceFile:     "beta.json",
		},
	}); err != nil {
		t.Fatalf("failed to seed beta table translation: %v", err)
	}

	entries := []TerminologyEntry{
		{
			ID:         "row-1",
			RecordType: "BOOK:FULL",
			SourceText: "Hello",
			SourceFile: "alpha.json",
		},
	}

	translations, err := store.GetPreviewTranslations(ctx, entries)
	if err != nil {
		t.Fatalf("GetPreviewTranslations returned error: %v", err)
	}

	got, ok := translations["row-1"]
	if !ok {
		t.Fatalf("missing translation for row-1")
	}
	if got.TranslationState != "translated" {
		t.Fatalf("unexpected translation state: got=%q want=%q", got.TranslationState, "translated")
	}
	if got.TranslatedText != "こんにちは" {
		t.Fatalf("unexpected translated text: got=%q want=%q", got.TranslatedText, "こんにちは")
	}
}
