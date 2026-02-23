package dictionary

import (
	"context"
	"io"
)

// DictionaryImporter orchestrates the XML parsing and dictionary persistence.
type DictionaryImporter interface {
	ImportXML(ctx context.Context, file io.Reader) (int, error)
}

// DictionaryStore persists dictionary terms to SQLite.
// This slice owns all table creation, INSERT/UPSERT operations.
type DictionaryStore interface {
	SaveTerms(ctx context.Context, terms []DictTerm) error
}
