package configstore

import (
	"context"
	"database/sql"
	"log/slog"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func newTestSQLiteStore(t *testing.T) (*SQLiteStore, *sql.DB) {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	store, err := NewSQLiteStore(context.Background(), db, slog.Default())
	if err != nil {
		_ = db.Close()
		t.Fatalf("new sqlite store: %v", err)
	}
	return store, db
}

func TestSQLiteStoreWatchUnsubscribeStopsNotification(t *testing.T) {
	store, db := newTestSQLiteStore(t)
	defer db.Close()

	ctx := context.Background()
	events := make(chan ChangeEvent, 4)
	unsubscribe := store.Watch("ns", "key", func(event ChangeEvent) {
		events <- event
	})

	if err := store.Set(ctx, "ns", "key", "v1"); err != nil {
		t.Fatalf("set v1: %v", err)
	}

	select {
	case <-events:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("watch callback was not called")
	}

	unsubscribe()

	if err := store.Set(ctx, "ns", "key", "v2"); err != nil {
		t.Fatalf("set v2: %v", err)
	}

	select {
	case event := <-events:
		t.Fatalf("unexpected event after unsubscribe: %+v", event)
	case <-time.After(200 * time.Millisecond):
	}
}
