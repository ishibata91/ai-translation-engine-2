package dictionarycontroller

import (
	"context"
	"fmt"
	"testing"

	dictionary "github.com/ishibata91/ai-translation-engine-2/pkg/slice/dictionary"
	"github.com/ishibata91/ai-translation-engine-2/pkg/tests/api_tests/testenv"
)

// Env bundles dictionary controller test dependencies.
type Env struct {
	Service *FakeService
	TestEnv *testenv.Env
}

// FakeService is a minimal stub for DictionaryController API tests.
type FakeService struct {
	LastCtx context.Context

	Sources           []dictionary.DictSource
	SourcesErr        error
	DeleteSourceErr   error
	Entries           []dictionary.DictTerm
	EntriesErr        error
	EntriesPage       *dictionary.DictTermPage
	EntriesPageErr    error
	SearchPage        *dictionary.DictTermPage
	SearchPageErr     error
	UpdateEntryErr    error
	DeleteEntryErr    error
	StartImportTaskID int64
	StartImportErr    error

	LastDeleteSourceID int64
	LastEntriesSource  int64
	LastPageSourceID   int64
	LastPageQuery      string
	LastPageFilters    map[string]string
	LastPageNo         int
	LastPageSize       int
	LastSearchQuery    string
	LastSearchFilters  map[string]string
	LastSearchPageNo   int
	LastSearchPageSize int
	LastUpdatedTerm    dictionary.DictTerm
	LastDeleteEntryID  int64
	LastImportPath     string
}

func (f *FakeService) GetSources(ctx context.Context) ([]dictionary.DictSource, error) {
	f.LastCtx = ctx
	return f.Sources, f.SourcesErr
}

func (f *FakeService) DeleteSource(ctx context.Context, id int64) error {
	f.LastCtx = ctx
	f.LastDeleteSourceID = id
	return f.DeleteSourceErr
}

func (f *FakeService) GetEntries(ctx context.Context, sourceID int64) ([]dictionary.DictTerm, error) {
	f.LastCtx = ctx
	f.LastEntriesSource = sourceID
	return f.Entries, f.EntriesErr
}

func (f *FakeService) GetEntriesPaginated(ctx context.Context, sourceID int64, query string, filters map[string]string, page, pageSize int) (*dictionary.DictTermPage, error) {
	f.LastCtx = ctx
	f.LastPageSourceID = sourceID
	f.LastPageQuery = query
	f.LastPageFilters = filters
	f.LastPageNo = page
	f.LastPageSize = pageSize
	return f.EntriesPage, f.EntriesPageErr
}

func (f *FakeService) SearchAll(ctx context.Context, query string, filters map[string]string, page, pageSize int) (*dictionary.DictTermPage, error) {
	f.LastCtx = ctx
	f.LastSearchQuery = query
	f.LastSearchFilters = filters
	f.LastSearchPageNo = page
	f.LastSearchPageSize = pageSize
	return f.SearchPage, f.SearchPageErr
}

func (f *FakeService) UpdateEntry(ctx context.Context, term dictionary.DictTerm) error {
	f.LastCtx = ctx
	f.LastUpdatedTerm = term
	return f.UpdateEntryErr
}

func (f *FakeService) DeleteEntry(ctx context.Context, id int64) error {
	f.LastCtx = ctx
	f.LastDeleteEntryID = id
	return f.DeleteEntryErr
}

func (f *FakeService) StartImport(ctx context.Context, filePath string) (int64, error) {
	f.LastCtx = ctx
	f.LastImportPath = filePath
	if f.StartImportTaskID == 0 {
		f.StartImportTaskID = 1
	}
	return f.StartImportTaskID, f.StartImportErr
}

// Build creates dictionary controller dependencies on shared testenv.
func Build(t *testing.T, name string) *Env {
	t.Helper()

	base := testenv.NewFileSQLiteEnv(t, name)
	return &Env{
		Service: &FakeService{},
		TestEnv: base,
	}
}

// String returns a short summary useful in failures.
func (e *Env) String() string {
	if e == nil || e.TestEnv == nil {
		return "<nil dictionarycontroller env>"
	}
	return fmt.Sprintf("db=%s trace_id=%s", e.TestEnv.DBPath, testenv.TraceIDValue(e.TestEnv.Ctx))
}
