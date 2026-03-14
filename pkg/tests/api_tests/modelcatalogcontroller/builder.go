package modelcatalogcontroller

import (
	"context"
	"fmt"
	"testing"

	modelcatalog "github.com/ishibata91/ai-translation-engine-2/pkg/runtime/modelcatalog"
	"github.com/ishibata91/ai-translation-engine-2/pkg/tests/api_tests/testenv"
)

// Env bundles model catalog controller test dependencies.
type Env struct {
	Service *FakeService
	TestEnv *testenv.Env
}

// FakeService stubs model catalog service behavior.
type FakeService struct {
	LastCtx   context.Context
	LastInput modelcatalog.ListModelsInput
	Models    []modelcatalog.ModelOption
	Err       error
}

func (s *FakeService) ListModels(ctx context.Context, input modelcatalog.ListModelsInput) ([]modelcatalog.ModelOption, error) {
	s.LastCtx = ctx
	s.LastInput = input
	return s.Models, s.Err
}

// Build creates model catalog controller dependencies on shared testenv.
func Build(t *testing.T, name string) *Env {
	t.Helper()
	base := testenv.NewFileSQLiteEnv(t, name)
	return &Env{Service: &FakeService{}, TestEnv: base}
}

// String returns a short summary useful in failures.
func (e *Env) String() string {
	if e == nil || e.TestEnv == nil {
		return "<nil modelcatalogcontroller env>"
	}
	return fmt.Sprintf("db=%s trace_id=%s", e.TestEnv.DBPath, testenv.TraceIDValue(e.TestEnv.Ctx))
}
