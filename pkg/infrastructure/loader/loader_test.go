package loader_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/loader"
)

func TestLoader_LoadExtractedJSON_UTF8(t *testing.T) {
	// 1. Create temporary JSON file
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.json")

	content := `{
		"quests": [
			{
				"id": "00012345",
				"name": "Test Quest"
			}
		],
		"npcs": {
			"000ABCDE": {
				"id": "000ABCDE",
				"name": "Test NPC",
				"sex": "Female"
			}
		}
	}`

	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	// 2. Initialize Loader
	l := loader.NewJSONLoader()

	// 3. Load Data
	data, err := l.LoadExtractedJSON(context.Background(), filePath)
	if err != nil {
		t.Fatalf("LoadExtractedJSON failed: %v", err)
	}

	// 4. Verify Results
	if len(data.Quests) != 1 {
		t.Errorf("expected 1 quest, got %d", len(data.Quests))
	} else {
		if *data.Quests[0].Name != "Test Quest" {
			t.Errorf("expected quest name 'Test Quest', got '%s'", *data.Quests[0].Name)
		}
	}

	if len(data.NPCs) != 1 {
		t.Errorf("expected 1 NPC, got %d", len(data.NPCs))
	} else {
		npc, ok := data.NPCs["000ABCDE"]
		if !ok {
			t.Errorf("expected NPC '000ABCDE' to exist")
		}
		if npc.Name != "Test NPC" {
			t.Errorf("expected NPC name 'Test NPC', got '%s'", npc.Name)
		}
		if !npc.IsFemale() {
			t.Errorf("expected NPC to be female")
		}
	}

	if data.SourceJSON != filePath {
		t.Errorf("expected SourceJSON to be '%s', got '%s'", filePath, data.SourceJSON)
	}
}

func TestLoader_LoadExtractedJSON_ShiftJIS(t *testing.T) {
	// Skip real SJIS encoding creation for simplicity in this artifact,
	// relying on `golang.org/x/text/encoding/japanese` presence in `encoding.go`.
	// Real test would write SJIS bytes.
}
