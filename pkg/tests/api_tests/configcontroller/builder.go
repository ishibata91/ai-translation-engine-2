package configcontroller

import (
	"fmt"
	"testing"

	"github.com/ishibata91/ai-translation-engine-2/pkg/gateway/configstore"
	"github.com/ishibata91/ai-translation-engine-2/pkg/tests/api_tests/testenv"
)

// Env bundles config controller test dependencies built from shared testenv.
type Env struct {
	Store   *configstore.SQLiteStore
	TestEnv *testenv.Env
}

// Build creates config controller dependencies on top of shared testenv.
func Build(t *testing.T, name string) *Env {
	t.Helper()

	base := testenv.NewFileSQLiteEnv(t, name)
	store, err := configstore.NewSQLiteStore(base.Ctx, base.DB, base.Logger)
	if err != nil {
		t.Fatalf("create sqlite store: %v", err)
	}

	return &Env{
		Store:   store,
		TestEnv: base,
	}
}

// TraceID returns the trace id used in this test environment.
func (e *Env) TraceID() string {
	if e == nil || e.TestEnv == nil {
		return ""
	}
	return testenv.TraceIDValue(e.TestEnv.Ctx)
}

// String returns a debug summary for failures.
func (e *Env) String() string {
	if e == nil || e.TestEnv == nil {
		return "<nil configcontroller env>"
	}
	return fmt.Sprintf("db=%s trace_id=%s", e.TestEnv.DBPath, e.TraceID())
}
