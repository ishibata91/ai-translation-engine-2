package testenv

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

type contextKey string

const (
	dbDirName  = "tmp/api_test_db"
	driverName = "sqlite"
	traceIDKey = contextKey("trace_id")
)

// Env provides shared controller API test infrastructure.
type Env struct {
	DB     *sql.DB
	Logger *slog.Logger
	Ctx    context.Context
	DBPath string
}

// NewFileSQLiteEnv creates a file-based SQLite-backed API test environment.
func NewFileSQLiteEnv(t *testing.T, name string) *Env {
	t.Helper()

	dbPath := newDBPath(t, name)
	dsn := fmt.Sprintf("file:%s", dbPath)

	db, err := sql.Open(driverName, dsn)
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
		_ = os.Remove(dbPath)
	})

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ctx := NewTraceContext(fmt.Sprintf("%s-%d", sanitize(name), time.Now().UnixNano()))

	return &Env{
		DB:     db,
		Logger: logger,
		Ctx:    ctx,
		DBPath: dbPath,
	}
}

// NewTraceContext returns a context carrying a test trace id.
func NewTraceContext(traceID string) context.Context {
	return context.WithValue(context.Background(), traceIDKey, traceID)
}

// TraceIDValue extracts the trace id from context.
func TraceIDValue(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	value, _ := ctx.Value(traceIDKey).(string)
	return value
}

func newDBPath(t *testing.T, name string) string {
	t.Helper()

	if err := os.MkdirAll(dbDirName, 0o755); err != nil {
		t.Fatalf("create test db dir: %v", err)
	}

	fileName := fmt.Sprintf("%s_%d.db", sanitize(name), time.Now().UnixNano())
	return filepath.Join(dbDirName, fileName)
}

func sanitize(name string) string {
	replaced := strings.NewReplacer(" ", "_", "/", "_", "\\", "_").Replace(name)
	if replaced == "" {
		return "api_test"
	}
	return replaced
}
