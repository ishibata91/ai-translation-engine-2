package term_translator

import (
	"context"

	"github.com/ishibata91/ai-translation-engine-2/pkg/domain/models"
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

// helper
func getEditorID(eid *string) string {
	if eid != nil {
		return *eid
	}
	return ""
}

// BuildRequests constructs translation requests from extracted data, applying NPC pairing.
func (b *TermRequestBuilderImpl) BuildRequests(ctx context.Context, data models.ExtractedData) ([]TermTranslationRequest, error) {
	var requests []TermTranslationRequest

	// Collect NPC Full names and Short names to pair them
	npcFulls := make(map[string]*models.NPC)
	npcShorts := make(map[string]*models.NPC)

	for _, npc := range data.NPCs {
		if !b.config.IsTarget(npc.Type) {
			continue
		}

		eid := getEditorID(npc.EditorID)
		if eid == "" {
			eid = npc.ID // fallback
		}

		if npc.Type == "NPC_:FULL" {
			n := npc
			npcFulls[eid] = &n
		} else if npc.Type == "NPC_:SHRT" {
			n := npc
			npcShorts[eid] = &n
		}
	}

	// Create pairs
	for editorID, fullNpc := range npcFulls {
		shortNpc, hasShort := npcShorts[editorID]

		req := TermTranslationRequest{
			FormID:     fullNpc.ID,
			EditorID:   getEditorID(fullNpc.EditorID),
			RecordType: "NPC_", // Use base prefix for paired request
			SourceText: fullNpc.Name,
			// SourcePlugin and SourceFile are typically part of a broader context,
			// omitting or passing defaults if not in model
			SourcePlugin: "Unknown",
			SourceFile:   "Unknown",
		}

		if hasShort {
			req.ShortName = shortNpc.Name
			// Remove from shorts map so we know it's processed
			delete(npcShorts, editorID)
		}

		requests = append(requests, req)
	}

	// Handle orphan short names (rare but possible)
	for _, shortNpc := range npcShorts {
		requests = append(requests, TermTranslationRequest{
			FormID:       shortNpc.ID,
			EditorID:     getEditorID(shortNpc.EditorID),
			RecordType:   shortNpc.Type,
			SourceText:   shortNpc.Name,
			SourcePlugin: "Unknown",
			SourceFile:   "Unknown",
		})
	}

	// Process Items
	for _, item := range data.Items {
		if !b.config.IsTarget(item.Type) {
			continue
		}
		name := ""
		if item.Name != nil {
			name = *item.Name
		} else if item.Text != nil {
			name = *item.Text
		}

		requests = append(requests, TermTranslationRequest{
			FormID:       item.ID,
			EditorID:     getEditorID(item.EditorID),
			RecordType:   item.Type,
			SourceText:   name,
			SourcePlugin: "Unknown",
			SourceFile:   "Unknown",
		})
	}

	// Process Magic
	for _, magic := range data.Magic {
		if !b.config.IsTarget(magic.Type) {
			continue
		}
		name := ""
		if magic.Name != nil {
			name = *magic.Name
		}
		requests = append(requests, TermTranslationRequest{
			FormID:       magic.ID,
			EditorID:     getEditorID(magic.EditorID),
			RecordType:   magic.Type,
			SourceText:   name,
			SourcePlugin: "Unknown",
			SourceFile:   "Unknown",
		})
	}

	// Process Locations
	for _, loc := range data.Locations {
		if !b.config.IsTarget(loc.Type) {
			continue
		}
		name := ""
		if loc.Name != nil {
			name = *loc.Name
		}
		requests = append(requests, TermTranslationRequest{
			FormID:       loc.ID,
			EditorID:     getEditorID(loc.EditorID),
			RecordType:   loc.Type,
			SourceText:   name,
			SourcePlugin: "Unknown",
			SourceFile:   "Unknown",
		})
	}

	return requests, nil
}
