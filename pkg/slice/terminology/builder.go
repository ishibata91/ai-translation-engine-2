package terminology

import (
	"context"
	"strings"
)

// TermRequestBuilderImpl implements TermRequestBuilder.
type TermRequestBuilderImpl struct {
	config *TermRecordConfig
}

// NewTermRequestBuilder creates a new TermRequestBuilderImpl.
func NewTermRequestBuilder(config *TermRecordConfig) *TermRequestBuilderImpl {
	return &TermRequestBuilderImpl{
		config: config,
	}
}

// BuildRequests constructs translation requests from normalized terminology entries.
func (b *TermRequestBuilderImpl) BuildRequests(ctx context.Context, data TerminologyInput) ([]TermTranslationRequest, error) {
	_ = ctx

	grouped := make(map[string][]TerminologyEntry)
	orderedKeys := make([]string, 0, len(data.Entries))

	for _, entry := range data.Entries {
		if !b.config.IsTarget(entry.RecordType) {
			continue
		}
		if strings.TrimSpace(entry.SourceText) == "" {
			continue
		}

		key := requestGroupKey(entry)
		if _, exists := grouped[key]; !exists {
			orderedKeys = append(orderedKeys, key)
		}
		grouped[key] = append(grouped[key], entry)
	}

	requests := make([]TermTranslationRequest, 0, len(orderedKeys))
	for _, key := range orderedKeys {
		requests = append(requests, buildRequestsForGroup(grouped[key])...)
	}

	return requests, nil
}

func requestGroupKey(entry TerminologyEntry) string {
	if isPairedNPCEntry(entry) && strings.TrimSpace(entry.PairKey) != "" {
		return "npc:" + strings.TrimSpace(entry.PairKey)
	}
	return "term:" + strings.TrimSpace(entry.RecordType) + "\x00" + strings.TrimSpace(entry.SourceText)
}

func buildRequestsForGroup(entries []TerminologyEntry) []TermTranslationRequest {
	if len(entries) == 0 {
		return nil
	}

	if isNPCGroup(entries) {
		return []TermTranslationRequest{buildNPCRequest(entries)}
	}
	return []TermTranslationRequest{buildSingleRequest(entries[0])}
}

func isNPCGroup(entries []TerminologyEntry) bool {
	if len(entries) == 0 {
		return false
	}
	return strings.HasPrefix(entries[0].RecordType, "NPC")
}

func isPairedNPCEntry(entry TerminologyEntry) bool {
	if !strings.HasPrefix(entry.RecordType, "NPC") {
		return false
	}
	return strings.TrimSpace(entry.PairKey) != ""
}

func buildNPCRequest(entries []TerminologyEntry) TermTranslationRequest {
	fullEntry := selectNPCEntry(entries, "full")
	if fullEntry == nil {
		fullEntry = &entries[0]
	}
	shortEntry := selectNPCEntry(entries, "short")

	request := TermTranslationRequest{
		FormID:       fullEntry.ID,
		EditorID:     fullEntry.EditorID,
		RecordType:   fullEntry.RecordType,
		SourceText:   fullEntry.SourceText,
		SourcePlugin: "Unknown",
		SourceFile:   fullEntry.SourceFile,
	}

	if shortEntry != nil {
		request.RecordType = "NPC_"
		request.ShortName = shortEntry.SourceText
	}

	return request
}

func selectNPCEntry(entries []TerminologyEntry, variant string) *TerminologyEntry {
	for i := range entries {
		if strings.EqualFold(strings.TrimSpace(entries[i].Variant), variant) {
			return &entries[i]
		}
	}
	return nil
}

func buildSingleRequest(entry TerminologyEntry) TermTranslationRequest {
	return TermTranslationRequest{
		FormID:       entry.ID,
		EditorID:     entry.EditorID,
		RecordType:   entry.RecordType,
		SourceText:   entry.SourceText,
		SourcePlugin: "Unknown",
		SourceFile:   entry.SourceFile,
	}
}
