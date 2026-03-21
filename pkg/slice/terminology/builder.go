package terminology

import (
	"context"
	"strings"
	"unicode"
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
		key := requestGroupKey(entry)
		if _, exists := grouped[key]; !exists {
			orderedKeys = append(orderedKeys, key)
		}
		grouped[key] = append(grouped[key], entry)
	}

	requests := make([]TermTranslationRequest, 0, len(orderedKeys))
	for _, key := range orderedKeys {
		request, ok := buildRequestForGroup(grouped[key])
		if !ok {
			continue
		}
		requests = append(requests, request)
	}

	return requests, nil
}

func requestGroupKey(entry TerminologyEntry) string {
	if isPairedNPCEntry(entry) && strings.TrimSpace(entry.PairKey) != "" {
		return "npc:" + strings.TrimSpace(entry.PairKey)
	}
	return "term:" + strings.TrimSpace(entry.RecordType) + "\x00" + strings.TrimSpace(entry.SourceText)
}

func buildRequestForGroup(entries []TerminologyEntry) (TermTranslationRequest, bool) {
	if len(entries) == 0 {
		return TermTranslationRequest{}, false
	}

	if isNPCGroup(entries) {
		return buildNPCRequest(entries)
	}

	entry := entries[0]
	if shouldExcludeEntry(entry) {
		return TermTranslationRequest{}, false
	}
	return buildSingleRequest(entry), true
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

func buildNPCRequest(entries []TerminologyEntry) (TermTranslationRequest, bool) {
	fullEntry := selectNPCEntry(entries, "full")
	if fullEntry == nil {
		fullEntry = &entries[0]
	}
	shortEntry := selectNPCEntry(entries, "short")
	if shouldExcludeEntry(*fullEntry) {
		return TermTranslationRequest{}, false
	}
	if shortEntry != nil && shouldExcludeEntry(*shortEntry) {
		return TermTranslationRequest{}, false
	}

	request := TermTranslationRequest{
		FormID:             fullEntry.ID,
		EditorID:           fullEntry.EditorID,
		RecordType:         fullEntry.RecordType,
		SourceText:         fullEntry.SourceText,
		OriginalSourceText: fullEntry.SourceText,
		SourcePlugin:       "Unknown",
		SourceFile:         fullEntry.SourceFile,
		Variant:            fullEntry.Variant,
	}

	if shortEntry != nil {
		request.RecordType = "NPC_"
		request.ShortName = shortEntry.SourceText
	}

	return request, true
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
		FormID:             entry.ID,
		EditorID:           entry.EditorID,
		RecordType:         entry.RecordType,
		SourceText:         entry.SourceText,
		OriginalSourceText: entry.SourceText,
		SourcePlugin:       "Unknown",
		SourceFile:         entry.SourceFile,
		Variant:            entry.Variant,
	}
}

func shouldExcludeEntry(entry TerminologyEntry) bool {
	sourceText := strings.TrimSpace(entry.SourceText)
	if sourceText == "" {
		return true
	}
	return containsJapaneseRune(sourceText)
}

func containsJapaneseRune(text string) bool {
	for _, r := range text {
		if isJapaneseRune(r) {
			return true
		}
	}
	return false
}

func isJapaneseRune(r rune) bool {
	if (r >= 0x3040 && r <= 0x309F) || (r >= 0x30A0 && r <= 0x30FF) || (r >= 0x4E00 && r <= 0x9FFF) {
		return true
	}
	return unicode.In(r, unicode.Han, unicode.Hiragana, unicode.Katakana)
}
