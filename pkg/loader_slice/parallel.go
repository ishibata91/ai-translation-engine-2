package loader_slice

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
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
	slog.DebugContext(ctx, "ENTER ParallelProcessor.Process")
	defer slog.DebugContext(ctx, "EXIT ParallelProcessor.Process")

	data := &models.ExtractedData{
		NPCs: make(map[string]models.NPC),
	}

	errChan := p.launchUnmarshalWorkers(ctx, data)

	result, err := p.waitAndCollectErrors(ctx, errChan)
	if err != nil {
		return nil, err
	}
	_ = result

	p.postProcess(data)
	return data, nil
}

// launchUnmarshalWorkers starts goroutines for each data section and returns the error channel.
func (p *ParallelProcessor) launchUnmarshalWorkers(ctx context.Context, data *models.ExtractedData) chan error {
	slog.DebugContext(ctx, "ENTER ParallelProcessor.launchUnmarshalWorkers")

	var g sync.WaitGroup
	errChan := make(chan error, 10)

	launch := func(fn func() error) {
		g.Add(1)
		go func() {
			defer g.Done()
			if err := fn(); err != nil {
				select {
				case errChan <- err:
				default:
				}
			}
		}()
	}

	launch(func() error { return p.unmarshalQuests(data) })
	launch(func() error { return p.unmarshalDialogueGroups(data) })
	launch(func() error { return p.unmarshalItems(data) })
	launch(func() error { return p.unmarshalNPCs(data) })
	launch(func() error { return p.unmarshalLocations(data) })
	launch(func() error { return p.unmarshalCells(data) })
	launch(func() error { return p.unmarshalMagic(data) })
	launch(func() error { return p.unmarshalSystem(data) })
	launch(func() error { return p.unmarshalMessages(data) })
	launch(func() error { return p.unmarshalLoadScreens(data) })

	// Close errChan when all workers are done
	go func() {
		g.Wait()
		close(errChan)
	}()

	return errChan
}

// waitAndCollectErrors waits for workers to complete and returns the first error if any.
func (p *ParallelProcessor) waitAndCollectErrors(ctx context.Context, errChan chan error) (bool, error) {
	slog.DebugContext(ctx, "ENTER ParallelProcessor.waitAndCollectErrors")

	for err := range errChan {
		if err != nil {
			return false, err
		}
	}

	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
		return true, nil
	}
}

// postProcess applies normalization to the loaded data.
func (p *ParallelProcessor) postProcess(data *models.ExtractedData) {
	slog.Debug("ENTER ParallelProcessor.postProcess")
	normalizeData(data)
}

// --- Section Unmarshalers ---

func (p *ParallelProcessor) unmarshalQuests(data *models.ExtractedData) error {
	if raw, ok := p.rawMap["quests"]; ok {
		var quests []models.Quest
		if err := json.Unmarshal(raw, &quests); err != nil {
			return fmt.Errorf("failed to unmarshal quests: %w", err)
		}
		data.Quests = quests
	}
	return nil
}

func (p *ParallelProcessor) unmarshalDialogueGroups(data *models.ExtractedData) error {
	if raw, ok := p.rawMap["dialogue_groups"]; ok {
		var dgs []models.DialogueGroup
		if err := json.Unmarshal(raw, &dgs); err != nil {
			return fmt.Errorf("failed to unmarshal dialogue_groups: %w", err)
		}
		data.DialogueGroups = dgs
	}
	return nil
}

func (p *ParallelProcessor) unmarshalItems(data *models.ExtractedData) error {
	if raw, ok := p.rawMap["items"]; ok {
		var items []models.Item
		if err := json.Unmarshal(raw, &items); err != nil {
			return fmt.Errorf("failed to unmarshal items: %w", err)
		}
		data.Items = items
	}
	return nil
}

func (p *ParallelProcessor) unmarshalNPCs(data *models.ExtractedData) error {
	if raw, ok := p.rawMap["npcs"]; ok {
		var npcs map[string]models.NPC
		if err := json.Unmarshal(raw, &npcs); err != nil {
			return fmt.Errorf("failed to unmarshal npcs: %w", err)
		}
		data.NPCs = npcs
	}
	return nil
}

func (p *ParallelProcessor) unmarshalLocations(data *models.ExtractedData) error {
	if raw, ok := p.rawMap["locations"]; ok {
		var locs []models.Location
		if err := json.Unmarshal(raw, &locs); err != nil {
			return fmt.Errorf("failed to unmarshal locations: %w", err)
		}
		data.Locations = locs
	}
	return nil
}

func (p *ParallelProcessor) unmarshalCells(data *models.ExtractedData) error {
	if raw, ok := p.rawMap["cells"]; ok {
		var cells []models.Location
		if err := json.Unmarshal(raw, &cells); err != nil {
			return fmt.Errorf("failed to unmarshal cells: %w", err)
		}
		data.Cells = cells
	}
	return nil
}

func (p *ParallelProcessor) unmarshalMagic(data *models.ExtractedData) error {
	if raw, ok := p.rawMap["magic"]; ok {
		var magics []models.Magic
		if err := json.Unmarshal(raw, &magics); err != nil {
			return fmt.Errorf("failed to unmarshal magic: %w", err)
		}
		data.Magic = magics
	}
	return nil
}

func (p *ParallelProcessor) unmarshalSystem(data *models.ExtractedData) error {
	if raw, ok := p.rawMap["system"]; ok {
		var sys []models.SystemRecord
		if err := json.Unmarshal(raw, &sys); err != nil {
			return fmt.Errorf("failed to unmarshal system: %w", err)
		}
		data.System = sys
	}
	return nil
}

func (p *ParallelProcessor) unmarshalMessages(data *models.ExtractedData) error {
	if raw, ok := p.rawMap["messages"]; ok {
		var msgs []models.Message
		if err := json.Unmarshal(raw, &msgs); err != nil {
			return fmt.Errorf("failed to unmarshal messages: %w", err)
		}
		data.Messages = msgs
	}
	return nil
}

func (p *ParallelProcessor) unmarshalLoadScreens(data *models.ExtractedData) error {
	if raw, ok := p.rawMap["load_screens"]; ok {
		var ls []models.LoadScreen
		if err := json.Unmarshal(raw, &ls); err != nil {
			return fmt.Errorf("failed to unmarshal load_screens: %w", err)
		}
		data.LoadScreens = ls
	}
	return nil
}

// --- Normalization ---

func normalizeData(data *models.ExtractedData) {
	slog.Debug("ENTER normalizeData")

	var wg sync.WaitGroup

	if len(data.NPCs) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			normalizeNPCNames(data)
		}()
	}

	wg.Wait()
}

// normalizeNPCNames trims whitespace from NPC names.
func normalizeNPCNames(data *models.ExtractedData) {
	slog.Debug("ENTER normalizeNPCNames")

	for k, npc := range data.NPCs {
		npc.Name = strings.TrimSpace(npc.Name)
		data.NPCs[k] = npc
	}
}
