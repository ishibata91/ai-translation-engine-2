package persona

import (
	"context"
	"testing"
)

func TestPersonaStore_UpsertAndDialogues(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name                 string
		overwriteExisting    bool
		secondPersonaText    string
		sourcePlugin         string
		sourceHint           string
		expectPersonaText    string
		expectDialogueSource string
		expectStoredPlugin   string
	}{
		{
			name:                 "上書きOFFでは既存personaを保持する",
			overwriteExisting:    false,
			secondPersonaText:    "updated persona",
			sourcePlugin:         "Skyrim.esm",
			expectPersonaText:    "base persona",
			expectDialogueSource: "first line",
			expectStoredPlugin:   "Skyrim.esm",
		},
		{
			name:                 "上書きONでは既存personaとdialogueを更新する",
			overwriteExisting:    true,
			secondPersonaText:    "updated persona",
			sourcePlugin:         "Skyrim.esm",
			expectPersonaText:    "updated persona",
			expectDialogueSource: "second line",
			expectStoredPlugin:   "Skyrim.esm",
		},
		{
			name:                 "source_plugin欠損時は入力名から補完する",
			overwriteExisting:    true,
			secondPersonaText:    "updated persona",
			sourceHint:           `C:\mods\Skyrim.esm.persona.json`,
			expectPersonaText:    "updated persona",
			expectDialogueSource: "second line",
			expectStoredPlugin:   "Skyrim.esm",
		},
		{
			name:                 "source_pluginの空白を含む名前を保持する",
			overwriteExisting:    true,
			secondPersonaText:    "updated persona",
			sourcePlugin:         "My Cool Mod.esp",
			expectPersonaText:    "updated persona",
			expectDialogueSource: "second line",
			expectStoredPlugin:   "My Cool Mod.esp",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			db, cleanup := setupTestDB(t)
			defer cleanup()

			store := NewPersonaStore(db)
			if err := store.InitSchema(ctx); err != nil {
				t.Fatalf("InitSchema failed: %v", err)
			}

			effectivePlugin := tc.sourcePlugin
			if effectivePlugin == "" {
				effectivePlugin = tc.expectStoredPlugin
			}

			saveState, err := store.SavePersonaBase(ctx, NPCDialogueData{
				SpeakerID:    "npc-001",
				EditorID:     "AelaEditor",
				NPCName:      "Aela",
				Race:         "Nord",
				Sex:          "Female",
				VoiceType:    "FemaleNord",
				SourcePlugin: tc.sourcePlugin,
				SourceHint:   tc.sourceHint,
				Dialogues: []DialogueEntry{
					{Text: "first line", EnglishText: "first line", Order: 1},
				},
			}, true)
			if err != nil {
				t.Fatalf("SavePersonaBase(first) failed: %v", err)
			}
			if err := store.ReplaceDialogues(ctx, saveState.PersonaID, tc.sourcePlugin, "npc-001", []DialogueEntry{
				{Text: "first line", EnglishText: "first line", Order: 1},
			}); err != nil {
				t.Fatalf("ReplaceDialogues(first) failed: %v", err)
			}
			if err := store.SaveGenerationRequest(ctx, effectivePlugin, "npc-001", "System Prompt:\nbase\n\nUser Prompt:\nfirst line"); err != nil {
				t.Fatalf("SaveGenerationRequest(first) failed: %v", err)
			}
			rows, err := store.ListNPCs(ctx)
			if err != nil {
				t.Fatalf("ListNPCs(after request) failed: %v", err)
			}
			if len(rows) != 1 {
				t.Fatalf("unexpected npc count after request: got=%d want=1", len(rows))
			}
			if rows[0].Status != "draft" {
				t.Fatalf("expected draft status after request generation, got=%q", rows[0].Status)
			}
			if err := store.SavePersona(ctx, PersonaResult{
				SpeakerID:    "npc-001",
				NPCName:      "Aela",
				PersonaText:  "base persona",
				SourcePlugin: effectivePlugin,
			}, true); err != nil {
				t.Fatalf("SavePersona(first) failed: %v", err)
			}

			saveState, err = store.SavePersonaBase(ctx, NPCDialogueData{
				SpeakerID:    "npc-001",
				EditorID:     "AelaEditor2",
				NPCName:      "Aela Updated",
				Race:         "Nord",
				Sex:          "Female",
				VoiceType:    "FemaleNord",
				SourcePlugin: tc.sourcePlugin,
				SourceHint:   tc.sourceHint,
				Dialogues: []DialogueEntry{
					{Text: "second line", EnglishText: "second line", Order: 1},
				},
			}, tc.overwriteExisting)
			if err != nil {
				t.Fatalf("SavePersonaBase(second) failed: %v", err)
			}
			if tc.overwriteExisting || saveState.PersonaText == "" {
				if err := store.ReplaceDialogues(ctx, saveState.PersonaID, tc.sourcePlugin, "npc-001", []DialogueEntry{
					{Text: "second line", EnglishText: "second line", Order: 1},
				}); err != nil {
					t.Fatalf("ReplaceDialogues(second) failed: %v", err)
				}
				if err := store.SaveGenerationRequest(ctx, effectivePlugin, "npc-001", "System Prompt:\nupdated\n\nUser Prompt:\nsecond line"); err != nil {
					t.Fatalf("SaveGenerationRequest(second) failed: %v", err)
				}
			}
			if err := store.SavePersona(ctx, PersonaResult{
				SpeakerID:    "npc-001",
				NPCName:      "Aela Updated",
				PersonaText:  tc.secondPersonaText,
				SourcePlugin: effectivePlugin,
			}, tc.overwriteExisting); err != nil {
				t.Fatalf("SavePersona(second) failed: %v", err)
			}

			personaText, err := store.GetPersona(ctx, effectivePlugin, "npc-001")
			if err != nil {
				t.Fatalf("GetPersona failed: %v", err)
			}
			if personaText != tc.expectPersonaText {
				t.Fatalf("unexpected persona text: got=%q want=%q", personaText, tc.expectPersonaText)
			}

			rows, err = store.ListNPCs(ctx)
			if err != nil {
				t.Fatalf("ListNPCs failed: %v", err)
			}
			if len(rows) != 1 {
				t.Fatalf("unexpected npc count: got=%d want=1", len(rows))
			}
			if rows[0].DialogueCount != 1 {
				t.Fatalf("expected aggregated dialogue count=1, got=%d", rows[0].DialogueCount)
			}
			if rows[0].Status != "generated" {
				t.Fatalf("expected generated status after persona save, got=%q", rows[0].Status)
			}
			if tc.overwriteExisting && rows[0].GenerationRequest != "System Prompt:\nupdated\n\nUser Prompt:\nsecond line" {
				t.Fatalf("unexpected generation request after overwrite: %q", rows[0].GenerationRequest)
			}
			if !tc.overwriteExisting && rows[0].GenerationRequest != "System Prompt:\nbase\n\nUser Prompt:\nfirst line" {
				t.Fatalf("unexpected generation request without overwrite: %q", rows[0].GenerationRequest)
			}

			dialogues, err := store.ListDialoguesByPersonaID(ctx, rows[0].PersonaID)
			if err != nil {
				t.Fatalf("ListDialoguesByPersonaID failed: %v", err)
			}
			if len(dialogues) != 1 {
				t.Fatalf("unexpected dialogue count: got=%d want=1", len(dialogues))
			}
			if dialogues[0].SourceText != tc.expectDialogueSource {
				t.Fatalf("unexpected dialogue source: got=%q want=%q", dialogues[0].SourceText, tc.expectDialogueSource)
			}
			if rows[0].SourcePlugin != tc.expectStoredPlugin {
				t.Fatalf("unexpected source_plugin on npc row: got=%q want=%q", rows[0].SourcePlugin, tc.expectStoredPlugin)
			}
			expectedDialoguePlugin := tc.expectStoredPlugin
			if tc.sourcePlugin == "" {
				expectedDialoguePlugin = "UNKNOWN"
			}
			if dialogues[0].SourcePlugin != expectedDialoguePlugin {
				t.Fatalf("unexpected source_plugin on dialogue row: got=%q want=%q", dialogues[0].SourcePlugin, expectedDialoguePlugin)
			}
		})
	}
}

func TestPersonaStore_AllowsSameSpeakerAcrossPlugins(t *testing.T) {
	ctx := context.Background()
	db, cleanup := setupTestDB(t)
	defer cleanup()

	store := NewPersonaStore(db)
	if err := store.InitSchema(ctx); err != nil {
		t.Fatalf("InitSchema failed: %v", err)
	}

	plugins := []string{"Skyrim.esm", "Update.esm"}
	for _, plugin := range plugins {
		saveState, err := store.SavePersonaBase(ctx, NPCDialogueData{
			SpeakerID:    "shared-speaker",
			NPCName:      plugin,
			SourcePlugin: plugin,
			Dialogues: []DialogueEntry{
				{Text: plugin, EnglishText: plugin, Order: 1},
			},
		}, true)
		if err != nil {
			t.Fatalf("SavePersonaBase failed: %v", err)
		}
		if err := store.ReplaceDialogues(ctx, saveState.PersonaID, plugin, "shared-speaker", []DialogueEntry{
			{Text: plugin, EnglishText: plugin, Order: 1},
		}); err != nil {
			t.Fatalf("ReplaceDialogues failed: %v", err)
		}
	}

	rows, err := store.ListNPCs(ctx)
	if err != nil {
		t.Fatalf("ListNPCs failed: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 persona rows, got %d", len(rows))
	}
}
