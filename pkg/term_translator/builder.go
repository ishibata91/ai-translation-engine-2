package term_translator

import (
	"context"
	"log/slog"

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
	slog.DebugContext(ctx, "ENTER TermRequestBuilderImpl.BuildRequests")
	defer slog.DebugContext(ctx, "EXIT TermRequestBuilderImpl.BuildRequests")

	var requests []TermTranslationRequest

	npcRequests := b.buildNPCPairedRequests(ctx, data)
	requests = append(requests, npcRequests...)

	itemRequests := b.buildItemRequests(ctx, data)
	requests = append(requests, itemRequests...)

	magicRequests := b.buildMagicRequests(ctx, data)
	requests = append(requests, magicRequests...)

	locationRequests := b.buildLocationRequests(ctx, data)
	requests = append(requests, locationRequests...)

	return requests, nil
}

// buildNPCPairedRequests creates paired NPC requests (FULL + SHRT) and orphan SHRT requests.
func (b *TermRequestBuilderImpl) buildNPCPairedRequests(ctx context.Context, data models.ExtractedData) []TermTranslationRequest {
	slog.DebugContext(ctx, "ENTER TermRequestBuilderImpl.buildNPCPairedRequests")

	npcFulls, npcShorts := b.classifyNPCs(data)
	requests := b.pairNPCRequests(npcFulls, npcShorts)
	orphanRequests := b.buildOrphanNPCRequests(npcShorts)

	return append(requests, orphanRequests...)
}

// classifyNPCs separates NPCs into FULL and SHRT maps keyed by EditorID.
func (b *TermRequestBuilderImpl) classifyNPCs(data models.ExtractedData) (map[string]*models.NPC, map[string]*models.NPC) {
	slog.Debug("ENTER TermRequestBuilderImpl.classifyNPCs")

	npcFulls := make(map[string]*models.NPC)
	npcShorts := make(map[string]*models.NPC)

	for _, npc := range data.NPCs {
		if !b.config.IsTarget(npc.Type) {
			continue
		}

		eid := getEditorID(npc.EditorID)
		if eid == "" {
			eid = npc.ID
		}

		if npc.Type == "NPC_:FULL" {
			n := npc
			npcFulls[eid] = &n
		} else if npc.Type == "NPC_:SHRT" {
			n := npc
			npcShorts[eid] = &n
		}
	}

	return npcFulls, npcShorts
}

// pairNPCRequests creates paired requests from FULL NPCs matched with their SHRT counterparts.
func (b *TermRequestBuilderImpl) pairNPCRequests(npcFulls map[string]*models.NPC, npcShorts map[string]*models.NPC) []TermTranslationRequest {
	slog.Debug("ENTER TermRequestBuilderImpl.pairNPCRequests")

	var requests []TermTranslationRequest

	for editorID, fullNpc := range npcFulls {
		shortNpc, hasShort := npcShorts[editorID]

		req := TermTranslationRequest{
			FormID:       fullNpc.ID,
			EditorID:     getEditorID(fullNpc.EditorID),
			RecordType:   "NPC_",
			SourceText:   fullNpc.Name,
			SourcePlugin: "Unknown",
			SourceFile:   "Unknown",
		}

		if hasShort {
			req.ShortName = shortNpc.Name
			delete(npcShorts, editorID)
		}

		requests = append(requests, req)
	}

	return requests
}

// buildOrphanNPCRequests creates requests for SHRT NPCs that had no FULL counterpart.
func (b *TermRequestBuilderImpl) buildOrphanNPCRequests(npcShorts map[string]*models.NPC) []TermTranslationRequest {
	slog.Debug("ENTER TermRequestBuilderImpl.buildOrphanNPCRequests")

	var requests []TermTranslationRequest

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

	return requests
}

// buildItemRequests creates translation requests for item records.
func (b *TermRequestBuilderImpl) buildItemRequests(ctx context.Context, data models.ExtractedData) []TermTranslationRequest {
	slog.DebugContext(ctx, "ENTER TermRequestBuilderImpl.buildItemRequests")

	var requests []TermTranslationRequest

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

	return requests
}

// buildMagicRequests creates translation requests for magic records.
func (b *TermRequestBuilderImpl) buildMagicRequests(ctx context.Context, data models.ExtractedData) []TermTranslationRequest {
	slog.DebugContext(ctx, "ENTER TermRequestBuilderImpl.buildMagicRequests")

	var requests []TermTranslationRequest

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

	return requests
}

// buildLocationRequests creates translation requests for location records.
func (b *TermRequestBuilderImpl) buildLocationRequests(ctx context.Context, data models.ExtractedData) []TermTranslationRequest {
	slog.DebugContext(ctx, "ENTER TermRequestBuilderImpl.buildLocationRequests")

	var requests []TermTranslationRequest

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

	return requests
}
