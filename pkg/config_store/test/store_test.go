package test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/ishibata91/ai-translation-engine-2/pkg/config_store"
	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) (*sql.DB, *config_store.SQLiteStore) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}

	ctx := context.Background()
	store, err := config_store.NewSQLiteStore(ctx, db)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	return db, store
}

func TestConfigStore_Basic(t *testing.T) {
	db, store := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	tests := []struct {
		name      string
		namespace string
		key       string
		value     string
	}{
		{"set string", "llm", "provider", "gemini"},
		{"update string", "llm", "provider", "openai"},
		{"numeric string", "llm", "temp", "0.7"},
		{"boolean string", "ui", "dark_mode", "true"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := store.Set(ctx, tt.namespace, tt.key, tt.value); err != nil {
				t.Errorf("Set() failed: %v", err)
			}
			val, err := store.Get(ctx, tt.namespace, tt.key)
			if err != nil {
				t.Errorf("Get() failed: %v", err)
			}
			if val != tt.value {
				t.Errorf("Expected %s, got %s", tt.value, val)
			}
		})
	}
}

func TestTypedAccessor(t *testing.T) {
	db, store := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	accessor := config_store.NewTypedAccessor(store)

	// Set some values
	store.Set(ctx, "test", "int", "123")
	store.Set(ctx, "test", "float", "3.14")
	store.Set(ctx, "test", "bool", "true")

	t.Run("GetInt", func(t *testing.T) {
		if val := accessor.GetInt(ctx, "test", "int", 0); val != 123 {
			t.Errorf("Expected 123, got %d", val)
		}
		if val := accessor.GetInt(ctx, "test", "missing", 456); val != 456 {
			t.Errorf("Expected default 456, got %d", val)
		}
	})

	t.Run("GetFloat", func(t *testing.T) {
		if val := accessor.GetFloat(ctx, "test", "float", 0.0); val != 3.14 {
			t.Errorf("Expected 3.14, got %f", val)
		}
	})

	t.Run("GetBool", func(t *testing.T) {
		if val := accessor.GetBool(ctx, "test", "bool", false); val != true {
			t.Errorf("Expected true, got %v", val)
		}
	})
}

func TestWatch(t *testing.T) {
	db, store := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()
	changed := make(chan bool, 1)

	store.Watch("test", "key", func(event config_store.ChangeEvent) {
		if event.NewValue == "new" {
			changed <- true
		}
	})

	store.Set(ctx, "test", "key", "new")

	select {
	case <-changed:
		// Success
	case <-time.After(time.Second):
		t.Error("Watch callback not triggered")
	}
}

func TestUIStateStore_JSON(t *testing.T) {
	db, store := setupTestDB(t)
	defer db.Close()

	ctx := context.Background()

	type Config struct {
		Size    int      `json:"size"`
		Options []string `json:"options"`
	}

	input := Config{Size: 100, Options: []string{"a", "b"}}
	if err := store.SetJSON(ctx, "ui", "layout", input); err != nil {
		t.Fatalf("SetJSON failed: %v", err)
	}

	var output Config
	if err := store.GetJSON(ctx, "ui", "layout", &output); err != nil {
		t.Fatalf("GetJSON failed: %v", err)
	}

	if output.Size != input.Size || len(output.Options) != 2 {
		t.Errorf("Expected %+v, got %+v", input, output)
	}
}
