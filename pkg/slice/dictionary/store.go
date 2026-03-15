package dictionary

import (
	"context"
	"fmt"

	dictionary_artifact "github.com/ishibata91/ai-translation-engine-2/pkg/artifact/dictionary_artifact"
)

type artifactDictionaryStore struct {
	repo dictionary_artifact.Repository
}

// NewDictionaryStore creates a slice store backed by dictionary artifact repository.
func NewDictionaryStore(repo dictionary_artifact.Repository) DictionaryStore {
	return &artifactDictionaryStore{repo: repo}
}

func (s *artifactDictionaryStore) GetSources(ctx context.Context) ([]DictSource, error) {
	sources, err := s.repo.GetSources(ctx)
	if err != nil {
		return nil, fmt.Errorf("get dictionary sources from artifact: %w", err)
	}
	out := make([]DictSource, 0, len(sources))
	for _, source := range sources {
		out = append(out, DictSource{
			ID:           source.ID,
			FileName:     source.FileName,
			Format:       source.Format,
			FilePath:     source.FilePath,
			FileSize:     source.FileSize,
			EntryCount:   source.EntryCount,
			Status:       source.Status,
			ErrorMessage: source.ErrorMessage,
			ImportedAt:   source.ImportedAt,
			CreatedAt:    source.CreatedAt,
		})
	}
	return out, nil
}

func (s *artifactDictionaryStore) CreateSource(ctx context.Context, src *DictSource) (int64, error) {
	id, err := s.repo.CreateSource(ctx, &dictionary_artifact.Source{
		FileName:   src.FileName,
		Format:     src.Format,
		FilePath:   src.FilePath,
		FileSize:   src.FileSize,
		EntryCount: src.EntryCount,
		Status:     src.Status,
	})
	if err != nil {
		return 0, fmt.Errorf("create dictionary source in artifact: %w", err)
	}
	return id, nil
}

func (s *artifactDictionaryStore) UpdateSourceStatus(ctx context.Context, id int64, status string, count int, errMsg string) error {
	if err := s.repo.UpdateSourceStatus(ctx, id, status, count, errMsg); err != nil {
		return fmt.Errorf("update dictionary source status in artifact id=%d: %w", id, err)
	}
	return nil
}

func (s *artifactDictionaryStore) DeleteSource(ctx context.Context, id int64) error {
	if err := s.repo.DeleteSource(ctx, id); err != nil {
		return fmt.Errorf("delete dictionary source in artifact id=%d: %w", id, err)
	}
	return nil
}

func (s *artifactDictionaryStore) GetEntriesBySourceID(ctx context.Context, sourceID int64) ([]DictTerm, error) {
	entries, err := s.repo.GetEntriesBySourceID(ctx, sourceID)
	if err != nil {
		return nil, fmt.Errorf("get dictionary entries from artifact source_id=%d: %w", sourceID, err)
	}
	return toSliceTerms(entries), nil
}

func (s *artifactDictionaryStore) GetEntriesBySourceIDPaginated(ctx context.Context, sourceID int64, query string, filters map[string]string, limit int, offset int) (*DictTermPage, error) {
	page, err := s.repo.GetEntriesBySourceIDPaginated(ctx, sourceID, query, filters, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("get dictionary entries paginated from artifact source_id=%d: %w", sourceID, err)
	}
	return &DictTermPage{
		Entries:    toSliceTerms(page.Entries),
		TotalCount: page.TotalCount,
	}, nil
}

func (s *artifactDictionaryStore) SearchAllEntriesPaginated(ctx context.Context, query string, filters map[string]string, limit int, offset int) (*DictTermPage, error) {
	page, err := s.repo.SearchAllEntriesPaginated(ctx, query, filters, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("search dictionary entries from artifact: %w", err)
	}
	return &DictTermPage{
		Entries:    toSliceTerms(page.Entries),
		TotalCount: page.TotalCount,
	}, nil
}

func (s *artifactDictionaryStore) SaveTerms(ctx context.Context, terms []DictTerm) error {
	entries := make([]dictionary_artifact.Entry, 0, len(terms))
	for _, term := range terms {
		entries = append(entries, dictionary_artifact.Entry{
			SourceID:   term.SourceID,
			EDID:       term.EDID,
			RecordType: term.RecordType,
			SourceText: term.Source,
			DestText:   term.Dest,
		})
	}
	if err := s.repo.SaveEntries(ctx, entries); err != nil {
		return fmt.Errorf("save dictionary entries in artifact: %w", err)
	}
	return nil
}

func (s *artifactDictionaryStore) UpdateEntry(ctx context.Context, term DictTerm) error {
	if err := s.repo.UpdateEntry(ctx, dictionary_artifact.Entry{
		ID:         term.ID,
		SourceText: term.Source,
		DestText:   term.Dest,
	}); err != nil {
		return fmt.Errorf("update dictionary entry in artifact id=%d: %w", term.ID, err)
	}
	return nil
}

func (s *artifactDictionaryStore) DeleteEntry(ctx context.Context, id int64) error {
	if err := s.repo.DeleteEntry(ctx, id); err != nil {
		return fmt.Errorf("delete dictionary entry in artifact id=%d: %w", id, err)
	}
	return nil
}

func toSliceTerms(entries []dictionary_artifact.Entry) []DictTerm {
	out := make([]DictTerm, 0, len(entries))
	for _, entry := range entries {
		out = append(out, DictTerm{
			ID:         entry.ID,
			SourceID:   entry.SourceID,
			SourceName: entry.SourceName,
			EDID:       entry.EDID,
			RecordType: entry.RecordType,
			Source:     entry.SourceText,
			Dest:       entry.DestText,
		})
	}
	return out
}
