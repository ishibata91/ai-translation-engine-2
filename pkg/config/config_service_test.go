package config

import (
	"context"
	"database/sql"
	"log/slog"
	"testing"

	_ "modernc.org/sqlite"
)

func setupConfigServiceTest(t *testing.T) (*sql.DB, *ConfigService) {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	store, err := NewSQLiteStore(context.Background(), db, slog.Default())
	if err != nil {
		t.Fatalf("failed to init sqlite store: %v", err)
	}
	return db, NewConfigService(store)
}

func TestConfigService_ConfigSetManyAndGetAll(t *testing.T) {
	db, service := setupConfigServiceTest(t)
	defer db.Close()

	values := map[string]string{
		"provider":    "lmstudio",
		"model":       "llama-3",
		"endpoint":    "http://localhost:1234",
		"api_key":     "plain-text-key",
		"temperature": "0.3",
		"max_tokens":  "500",
	}
	if err := service.ConfigSetMany("master_persona.llm", values); err != nil {
		t.Fatalf("ConfigSetMany failed: %v", err)
	}

	got, err := service.ConfigGetAll("master_persona.llm")
	if err != nil {
		t.Fatalf("ConfigGetAll failed: %v", err)
	}
	if len(got) != len(values) {
		t.Fatalf("unexpected key count: got=%d want=%d", len(got), len(values))
	}
	for key, want := range values {
		if got[key] != want {
			t.Fatalf("unexpected value for %s: got=%s want=%s", key, got[key], want)
		}
	}
}

func TestConfigService_ConfigGet_MissingReturnsEmpty(t *testing.T) {
	db, service := setupConfigServiceTest(t)
	defer db.Close()

	got, err := service.ConfigGet("master_persona.llm", "provider")
	if err != nil {
		t.Fatalf("ConfigGet failed: %v", err)
	}
	if got != "" {
		t.Fatalf("expected empty for missing key, got=%s", got)
	}
}
