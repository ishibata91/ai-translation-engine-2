package loader_slice

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/ishibata91/ai-translation-engine-2/pkg/domain/models"
)

// ParallelProcessor handles the parallel unmarshaling and normalization of ExtractedData.
type ParallelProcessor struct {
	rawMap map[string]json.RawMessage
}

// NewParallelProcessor creates a new processor with the raw JSON map.
func NewParallelProcessor(rawMap map[string]json.RawMessage) *ParallelProcessor {
	return &ParallelProcessor{rawMap: rawMap}
}

// Process executes the parallel unmarshaling and returns the constructed ExtractedData.
func (p *ParallelProcessor) Process(ctx context.Context) (*models.ExtractedData, error) {
	data := &models.ExtractedData{
		NPCs: make(map[string]models.NPC),
	}

	var g sync.WaitGroup
	errChan := make(chan error, 10) // buffer for errors

	// Helper to launch goroutine
	launch := func(fn func() error) {
		g.Add(1)
		go func() {
			defer g.Done()
			if err := fn(); err != nil {
				select {
				case errChan <- err:
				default:
					// Channel full, drop error (or log it if logger available)
				}
			}
		}()
	}

	// 1. Quests
	launch(func() error {
		if raw, ok := p.rawMap["quests"]; ok {
			var quests []models.Quest
			if err := json.Unmarshal(raw, &quests); err != nil {
				return fmt.Errorf("failed to unmarshal quests: %w", err)
			}
			data.Quests = quests
		}
		return nil
	})

	// 2. Dialogue Groups
	launch(func() error {
		if raw, ok := p.rawMap["dialogue_groups"]; ok {
			var dgs []models.DialogueGroup
			if err := json.Unmarshal(raw, &dgs); err != nil {
				return fmt.Errorf("failed to unmarshal dialogue_groups: %w", err)
			}
			data.DialogueGroups = dgs
		}
		return nil
	})

	// 3. Items
	launch(func() error {
		if raw, ok := p.rawMap["items"]; ok {
			var items []models.Item
			if err := json.Unmarshal(raw, &items); err != nil {
				return fmt.Errorf("failed to unmarshal items: %w", err)
			}
			data.Items = items
		}
		return nil
	})

	// 4. NPCs (Map Structure)
	launch(func() error {
		if raw, ok := p.rawMap["npcs"]; ok {
			var npcs map[string]models.NPC
			if err := json.Unmarshal(raw, &npcs); err != nil {
				return fmt.Errorf("failed to unmarshal npcs: %w", err)
			}
			// Normalization: Extract EditorID if needed, though strictly it should be done here.
			// For now, simple unmarshal is enough as per specs.
			data.NPCs = npcs
		}
		return nil
	})

	// 5. Locations
	launch(func() error {
		if raw, ok := p.rawMap["locations"]; ok {
			var locs []models.Location
			if err := json.Unmarshal(raw, &locs); err != nil {
				return fmt.Errorf("failed to unmarshal locations: %w", err)
			}
			data.Locations = locs
		}
		return nil
	})

	// 6. Cells
	launch(func() error {
		if raw, ok := p.rawMap["cells"]; ok {
			var cells []models.Location
			if err := json.Unmarshal(raw, &cells); err != nil {
				return fmt.Errorf("failed to unmarshal cells: %w", err)
			}
			data.Cells = cells
		}
		return nil
	})

	// 7. Magic
	launch(func() error {
		if raw, ok := p.rawMap["magic"]; ok {
			var magics []models.Magic
			if err := json.Unmarshal(raw, &magics); err != nil {
				return fmt.Errorf("failed to unmarshal magic: %w", err)
			}
			data.Magic = magics
		}
		return nil
	})

	// 8. System
	launch(func() error {
		if raw, ok := p.rawMap["system"]; ok {
			var sys []models.SystemRecord
			if err := json.Unmarshal(raw, &sys); err != nil {
				return fmt.Errorf("failed to unmarshal system: %w", err)
			}
			data.System = sys
		}
		return nil
	})

	// 9. Messages
	launch(func() error {
		if raw, ok := p.rawMap["messages"]; ok {
			var msgs []models.Message
			if err := json.Unmarshal(raw, &msgs); err != nil {
				return fmt.Errorf("failed to unmarshal messages: %w", err)
			}
			data.Messages = msgs
		}
		return nil
	})

	// 10. Load Screens
	launch(func() error {
		if raw, ok := p.rawMap["load_screens"]; ok {
			var ls []models.LoadScreen
			if err := json.Unmarshal(raw, &ls); err != nil {
				return fmt.Errorf("failed to unmarshal load_screens: %w", err)
			}
			data.LoadScreens = ls
		}
		return nil
	})

	// Wait for all goroutines
	done := make(chan struct{})
	go func() {
		g.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-errChan:
		// Return the first error encountered
		return nil, err
	case <-done:
		// Check if any error occurred but wasn't caught by select due to timing
		select {
		case err := <-errChan:
			return nil, err
		default:
			// Normalization: Post-process specific fields if required by spec.
			// Example: Extract EditorID from Name if Name format is "ProperName [EDID:123]"
			// This can be parallelized inside the specific loaders above, but keeping it simple here.
			normalizeData(data)
			return data, nil
		}
	}
}

func normalizeData(data *models.ExtractedData) {
	// Example normalization: Check ID consistency or other simple checks.
	// For Phase 1, strictly map the fields.
	// Future: Parallel loop over data.NPCs to methods like extractEditorID(n.Name)
	var wg sync.WaitGroup

	// Normalize NPCs efficiently
	if len(data.NPCs) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for k, npc := range data.NPCs {
				// Example: If Name contains brackets, it might need parsing.
				// Based on extractData.pas, EditorID is already a separate field.
				// But we might want to trim spaces.
				npc.Name = strings.TrimSpace(npc.Name)
				data.NPCs[k] = npc
			}
		}()
	}

	wg.Wait()
}
