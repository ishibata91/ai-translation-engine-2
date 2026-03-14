package controller

import (
	"errors"
	"testing"

	dictionary "github.com/ishibata91/ai-translation-engine-2/pkg/slice/dictionary"
	dictionarycontrollertest "github.com/ishibata91/ai-translation-engine-2/pkg/tests/api_tests/dictionarycontroller"
	apitestenv "github.com/ishibata91/ai-translation-engine-2/pkg/tests/api_tests/testenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDictionaryController_API_TableDriven(t *testing.T) {
	filters := map[string]string{"recordType": "INFO"}
	errDummy := errors.New("dummy")
	expectedTraceID := "dictionary-controller-trace"

	testCases := []struct {
		name string
		run  func(t *testing.T, controller *DictionaryController, fake *dictionarycontrollertest.FakeService)
	}{
		{
			name: "DictGetSources returns service result",
			run: func(t *testing.T, controller *DictionaryController, fake *dictionarycontrollertest.FakeService) {
				expected := []dictionary.DictSource{{ID: 1, FileName: "a.xml"}}
				fake.Sources = expected

				got, err := controller.DictGetSources()
				require.NoError(t, err)
				assert.Equal(t, expected, got)
				assert.NotNil(t, fake.LastCtx)
				assert.Equal(t, expectedTraceID, apitestenv.TraceIDValue(fake.LastCtx))
			},
		},
		{
			name: "DictGetSources returns service error",
			run: func(t *testing.T, controller *DictionaryController, fake *dictionarycontrollertest.FakeService) {
				fake.SourcesErr = errDummy
				_, err := controller.DictGetSources()
				require.Error(t, err)
				assert.ErrorIs(t, err, errDummy)
			},
		},
		{
			name: "DictDeleteSource delegates id",
			run: func(t *testing.T, controller *DictionaryController, fake *dictionarycontrollertest.FakeService) {
				err := controller.DictDeleteSource(42)
				require.NoError(t, err)
				assert.Equal(t, int64(42), fake.LastDeleteSourceID)
			},
		},
		{
			name: "DictGetEntries delegates source id",
			run: func(t *testing.T, controller *DictionaryController, fake *dictionarycontrollertest.FakeService) {
				expected := []dictionary.DictTerm{{ID: 10, SourceID: 9}}
				fake.Entries = expected
				got, err := controller.DictGetEntries(9)
				require.NoError(t, err)
				assert.Equal(t, expected, got)
				assert.Equal(t, int64(9), fake.LastEntriesSource)
			},
		},
		{
			name: "DictGetEntriesPaginated delegates params",
			run: func(t *testing.T, controller *DictionaryController, fake *dictionarycontrollertest.FakeService) {
				expected := &dictionary.DictTermPage{Entries: []dictionary.DictTerm{{ID: 1}}, TotalCount: 1}
				fake.EntriesPage = expected
				got, err := controller.DictGetEntriesPaginated(3, "hello", filters, 2, 50)
				require.NoError(t, err)
				assert.Equal(t, expected, got)
				assert.Equal(t, int64(3), fake.LastPageSourceID)
				assert.Equal(t, "hello", fake.LastPageQuery)
				assert.Equal(t, filters, fake.LastPageFilters)
				assert.Equal(t, 2, fake.LastPageNo)
				assert.Equal(t, 50, fake.LastPageSize)
			},
		},
		{
			name: "DictSearchAllEntriesPaginated delegates params",
			run: func(t *testing.T, controller *DictionaryController, fake *dictionarycontrollertest.FakeService) {
				expected := &dictionary.DictTermPage{Entries: []dictionary.DictTerm{{ID: 2}}, TotalCount: 1}
				fake.SearchPage = expected
				got, err := controller.DictSearchAllEntriesPaginated("q", filters, 1, 25)
				require.NoError(t, err)
				assert.Equal(t, expected, got)
				assert.Equal(t, "q", fake.LastSearchQuery)
				assert.Equal(t, filters, fake.LastSearchFilters)
				assert.Equal(t, 1, fake.LastSearchPageNo)
				assert.Equal(t, 25, fake.LastSearchPageSize)
			},
		},
		{
			name: "DictUpdateEntry delegates term and returns error",
			run: func(t *testing.T, controller *DictionaryController, fake *dictionarycontrollertest.FakeService) {
				fake.UpdateEntryErr = errDummy
				term := dictionary.DictTerm{ID: 7, Source: "src", Dest: "dst"}
				err := controller.DictUpdateEntry(term)
				require.Error(t, err)
				assert.ErrorIs(t, err, errDummy)
				assert.Equal(t, term, fake.LastUpdatedTerm)
			},
		},
		{
			name: "DictDeleteEntry delegates id and returns error",
			run: func(t *testing.T, controller *DictionaryController, fake *dictionarycontrollertest.FakeService) {
				fake.DeleteEntryErr = errDummy
				err := controller.DictDeleteEntry(8)
				require.Error(t, err)
				assert.ErrorIs(t, err, errDummy)
				assert.Equal(t, int64(8), fake.LastDeleteEntryID)
			},
		},
		{
			name: "DictStartImport returns task id and delegates path",
			run: func(t *testing.T, controller *DictionaryController, fake *dictionarycontrollertest.FakeService) {
				fake.StartImportTaskID = 99
				id, err := controller.DictStartImport("C:/tmp/test.xml")
				require.NoError(t, err)
				assert.Equal(t, int64(99), id)
				assert.Equal(t, "C:/tmp/test.xml", fake.LastImportPath)
			},
		},
		{
			name: "DictStartImport returns error",
			run: func(t *testing.T, controller *DictionaryController, fake *dictionarycontrollertest.FakeService) {
				fake.StartImportErr = errDummy
				_, err := controller.DictStartImport("bad.xml")
				require.Error(t, err)
				assert.ErrorIs(t, err, errDummy)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			env := dictionarycontrollertest.Build(t, tc.name)
			controller := NewDictionaryController(env.Service)
			controller.SetContext(apitestenv.NewTraceContext(expectedTraceID))
			tc.run(t, controller, env.Service)
		})
	}
}
