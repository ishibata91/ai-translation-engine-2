package controller

import (
	"context"

	"github.com/ishibata91/ai-translation-engine-2/pkg/runtime/telemetry"
	dictionary2 "github.com/ishibata91/ai-translation-engine-2/pkg/slice/dictionary"
)

// DictionaryController exposes Wails-facing dictionary operations.
type DictionaryController struct {
	ctx     context.Context
	service *dictionary2.DictionaryService
}

// NewDictionaryController constructs the dictionary controller adapter.
func NewDictionaryController(service *dictionary2.DictionaryService) *DictionaryController {
	return &DictionaryController{
		ctx:     context.Background(),
		service: service,
	}
}

// SetContext injects the Wails application context for downstream propagation.
func (c *DictionaryController) SetContext(ctx context.Context) {
	if ctx == nil {
		c.ctx = context.Background()
		return
	}
	c.ctx = ctx
}

// DictGetSources returns all registered dictionary sources.
func (c *DictionaryController) DictGetSources() ([]dictionary2.DictSource, error) {
	return c.service.GetSources(c.context())
}

// DictDeleteSource removes a dictionary source and its entries.
func (c *DictionaryController) DictDeleteSource(id int64) error {
	return c.service.DeleteSource(c.context(), id)
}

// DictGetEntries returns dictionary entries for the source.
func (c *DictionaryController) DictGetEntries(sourceID int64) ([]dictionary2.DictTerm, error) {
	return c.service.GetEntries(c.context(), sourceID)
}

// DictGetEntriesPaginated returns paginated dictionary entries for the source.
func (c *DictionaryController) DictGetEntriesPaginated(sourceID int64, query string, filters map[string]string, page, pageSize int) (*dictionary2.DictTermPage, error) {
	return c.service.GetEntriesPaginated(c.context(), sourceID, query, filters, page, pageSize)
}

// DictSearchAllEntriesPaginated searches dictionary entries across all sources.
func (c *DictionaryController) DictSearchAllEntriesPaginated(query string, filters map[string]string, page, pageSize int) (*dictionary2.DictTermPage, error) {
	return c.service.SearchAll(c.context(), query, filters, page, pageSize)
}

// DictUpdateEntry updates one dictionary entry.
func (c *DictionaryController) DictUpdateEntry(term dictionary2.DictTerm) error {
	return c.service.UpdateEntry(c.context(), term)
}

// DictDeleteEntry removes one dictionary entry.
func (c *DictionaryController) DictDeleteEntry(id int64) error {
	return c.service.DeleteEntry(c.context(), id)
}

// DictStartImport starts dictionary import for one file.
func (c *DictionaryController) DictStartImport(filePath string) (int64, error) {
	return c.service.StartImport(c.context(), filePath)
}

func (c *DictionaryController) context() context.Context {
	return telemetry.WithTraceID(c.ctx)
}
