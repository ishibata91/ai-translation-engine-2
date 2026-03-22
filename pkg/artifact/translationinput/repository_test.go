package translationinput

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/ishibata91/ai-translation-engine-2/pkg/format/parser/skyrim"
	_ "modernc.org/sqlite"
)

func TestRepository_SaveAndPreview_TableDriven(t *testing.T) {
	testCases := []struct {
		name                       string
		buildOutput                func() *skyrim.ParserOutput
		expectedPreviewRows        int
		expectedTerminologyEntries int
		verify                     func(t *testing.T, db *sql.DB, repo Repository, file InputFile)
	}{
		{
			name:                       "all sections are saved and projected into preview",
			buildOutput:                buildAllSectionOutput,
			expectedPreviewRows:        19,
			expectedTerminologyEntries: 9,
			verify: func(t *testing.T, db *sql.DB, repo Repository, file InputFile) {
				t.Helper()
				assertQueryCount(t, db, `SELECT COUNT(1) FROM translation_input_dialogue_groups WHERE file_id = ?`, file.ID, 1)
				assertQueryCount(t, db, `SELECT COUNT(1) FROM translation_input_dialogue_responses r JOIN translation_input_dialogue_groups g ON g.id = r.dialogue_group_id WHERE g.file_id = ?`, file.ID, 1)
				assertQueryCount(t, db, `SELECT COUNT(1) FROM translation_input_quests WHERE file_id = ?`, file.ID, 1)
				assertQueryCount(t, db, `SELECT COUNT(1) FROM translation_input_quest_stages qs JOIN translation_input_quests q ON q.id = qs.quest_id WHERE q.file_id = ?`, file.ID, 1)
				assertQueryCount(t, db, `SELECT COUNT(1) FROM translation_input_quest_objectives qo JOIN translation_input_quests q ON q.id = qo.quest_id WHERE q.file_id = ?`, file.ID, 1)

				preview, err := repo.ListPreviewRows(context.Background(), file.ID, 1, 50)
				if err != nil {
					t.Fatalf("ListPreviewRows failed: %v", err)
				}
				requiredSections := map[string]bool{
					"dialogue_response":  false,
					"quest_stage":        false,
					"quest_objective":    false,
					"item_name":          false,
					"item_description":   false,
					"item_text":          false,
					"magic_name":         false,
					"magic_description":  false,
					"location_name":      false,
					"cell_name":          false,
					"system_name":        false,
					"system_description": false,
					"message_text":       false,
					"message_title":      false,
					"load_screen_text":   false,
					"npc_name":           false,
				}
				for _, row := range preview.Rows {
					if _, ok := requiredSections[row.Section]; ok {
						requiredSections[row.Section] = true
					}
				}
				for section, seen := range requiredSections {
					if !seen {
						t.Fatalf("section %s was not projected in preview", section)
					}
				}
				terminologyInput, err := repo.LoadTerminologyInput(context.Background(), "task-translation-1")
				if err != nil {
					t.Fatalf("LoadTerminologyInput failed: %v", err)
				}
				if len(terminologyInput.Entries) != 9 {
					t.Fatalf("unexpected terminology entry count: got=%d want=9", len(terminologyInput.Entries))
				}
				if terminologyInput.Entries[0].SourceFile != file.SourceFileName {
					t.Fatalf("unexpected source file: got=%q want=%q", terminologyInput.Entries[0].SourceFile, file.SourceFileName)
				}
				foundWEAP := false
				foundNPCFull := false
				for _, entry := range terminologyInput.Entries {
					if entry.RecordType == "WEAP:FULL" && entry.SourceText == "Weapon Name" {
						foundWEAP = true
					}
					if entry.RecordType == "NPC_:FULL" && entry.Variant == "full" {
						foundNPCFull = true
					}
				}
				if !foundWEAP {
					t.Fatalf("expected WEAP:FULL terminology entry to be preserved")
				}
				if !foundNPCFull {
					t.Fatalf("expected npc variant rows to preserve variant values: full=%t", foundNPCFull)
				}
			},
		},
		{
			name:                       "editor_id is projected from source_record_id fallback when empty",
			buildOutput:                buildOutputWithoutEditorID,
			expectedPreviewRows:        1,
			expectedTerminologyEntries: 1,
			verify: func(t *testing.T, _ *sql.DB, repo Repository, file InputFile) {
				t.Helper()
				preview, err := repo.ListPreviewRows(context.Background(), file.ID, 1, 50)
				if err != nil {
					t.Fatalf("ListPreviewRows failed: %v", err)
				}
				if len(preview.Rows) != 1 {
					t.Fatalf("unexpected preview row count: got=%d want=1", len(preview.Rows))
				}
				if preview.Rows[0].EditorID != "item-no-edid" {
					t.Fatalf("unexpected editor_id fallback: got=%q want=%q", preview.Rows[0].EditorID, "item-no-edid")
				}
				terminologyInput, err := repo.LoadTerminologyInput(context.Background(), "task-translation-1")
				if err != nil {
					t.Fatalf("LoadTerminologyInput failed: %v", err)
				}
				if len(terminologyInput.Entries) != 1 {
					t.Fatalf("unexpected terminology entry count: got=%d want=1", len(terminologyInput.Entries))
				}
			},
		},
		{
			name:                       "preview paging uses 50-row boundaries",
			buildOutput:                func() *skyrim.ParserOutput { return buildDialogueOnlyOutput(55) },
			expectedPreviewRows:        55,
			expectedTerminologyEntries: 0,
			verify: func(t *testing.T, _ *sql.DB, repo Repository, file InputFile) {
				t.Helper()
				firstPage, err := repo.ListPreviewRows(context.Background(), file.ID, 1, 50)
				if err != nil {
					t.Fatalf("ListPreviewRows page1 failed: %v", err)
				}
				if firstPage.TotalRows != 55 {
					t.Fatalf("unexpected total rows on page1: got=%d want=55", firstPage.TotalRows)
				}
				if len(firstPage.Rows) != 50 {
					t.Fatalf("unexpected row count on page1: got=%d want=50", len(firstPage.Rows))
				}

				secondPage, err := repo.ListPreviewRows(context.Background(), file.ID, 2, 50)
				if err != nil {
					t.Fatalf("ListPreviewRows page2 failed: %v", err)
				}
				if secondPage.TotalRows != 55 {
					t.Fatalf("unexpected total rows on page2: got=%d want=55", secondPage.TotalRows)
				}
				if len(secondPage.Rows) != 5 {
					t.Fatalf("unexpected row count on page2: got=%d want=5", len(secondPage.Rows))
				}
				if len(firstPage.Rows) == 0 || len(secondPage.Rows) == 0 {
					t.Fatalf("expected non-empty paging rows")
				}
				if firstPage.Rows[0].ID == secondPage.Rows[0].ID {
					t.Fatalf("page1 and page2 must not start from the same row")
				}
				terminologyInput, err := repo.LoadTerminologyInput(context.Background(), "task-translation-1")
				if err != nil {
					t.Fatalf("LoadTerminologyInput failed: %v", err)
				}
				if len(terminologyInput.Entries) != 0 {
					t.Fatalf("unexpected terminology entry count: got=%d want=0", len(terminologyInput.Entries))
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			db, cleanup := setupRepositoryTestDB(t)
			defer cleanup()

			repo := NewRepository(db)
			input := tc.buildOutput()
			savedFile, err := repo.SaveParsedOutput(context.Background(), "task-translation-1", `C:\mods\input.json`, input)
			if err != nil {
				t.Fatalf("SaveParsedOutput failed: %v", err)
			}

			if savedFile.PreviewRowCount != tc.expectedPreviewRows {
				t.Fatalf("unexpected preview row count: got=%d want=%d", savedFile.PreviewRowCount, tc.expectedPreviewRows)
			}

			files, err := repo.ListFiles(context.Background(), "task-translation-1")
			if err != nil {
				t.Fatalf("ListFiles failed: %v", err)
			}
			if len(files) != 1 {
				t.Fatalf("unexpected file count: got=%d want=1", len(files))
			}
			if files[0].PreviewRowCount != tc.expectedPreviewRows {
				t.Fatalf("unexpected listed preview row count: got=%d want=%d", files[0].PreviewRowCount, tc.expectedPreviewRows)
			}

			terminologyInput, err := repo.LoadTerminologyInput(context.Background(), "task-translation-1")
			if err != nil {
				t.Fatalf("LoadTerminologyInput failed: %v", err)
			}
			if len(terminologyInput.Entries) != tc.expectedTerminologyEntries {
				t.Fatalf("unexpected terminology entry count: got=%d want=%d", len(terminologyInput.Entries), tc.expectedTerminologyEntries)
			}

			if tc.verify != nil {
				tc.verify(t, db, repo, savedFile)
			}
		})
	}
}

func TestRepository_LoadTerminologyInput_NormalizesLegacyRecordTypes(t *testing.T) {
	db, cleanup := setupRepositoryTestDB(t)
	defer cleanup()

	repo := NewRepository(db)
	editorID := "EditorLegacy"
	name := "Legacy Weapon"

	if _, err := repo.SaveParsedOutput(context.Background(), "task-legacy", `C:\mods\legacy.json`, &skyrim.ParserOutput{
		Items: []skyrim.Item{{
			BaseExtractedRecord: skyrim.BaseExtractedRecord{
				ID:         "legacy-item-1",
				EditorID:   &editorID,
				Type:       "WEAP FULL",
				SourceJSON: `C:\mods\legacy.json`,
			},
			Name: &name,
		}},
	}); err != nil {
		t.Fatalf("SaveParsedOutput failed: %v", err)
	}

	input, err := repo.LoadTerminologyInput(context.Background(), "task-legacy")
	if err != nil {
		t.Fatalf("LoadTerminologyInput failed: %v", err)
	}
	if len(input.Entries) != 1 {
		t.Fatalf("unexpected terminology entry count: got=%d want=1", len(input.Entries))
	}
	if input.Entries[0].RecordType != "WEAP:FULL" {
		t.Fatalf("unexpected normalized record type: got=%q want=%q", input.Entries[0].RecordType, "WEAP:FULL")
	}
}

func TestRepository_LoadTerminologyInput_ExcludesNonDictionaryREC(t *testing.T) {
	db, cleanup := setupRepositoryTestDB(t)
	defer cleanup()

	repo := NewRepository(db)
	editorID := "EditorBook"
	bookName := "Client Details"
	bookText := "<font face='$HandwrittenFont'>Deliver the enchanted sword</font>"

	if _, err := repo.SaveParsedOutput(context.Background(), "task-book", `C:\mods\book.json`, &skyrim.ParserOutput{
		Items: []skyrim.Item{
			{
				BaseExtractedRecord: skyrim.BaseExtractedRecord{
					ID:         "book-full-1",
					EditorID:   &editorID,
					Type:       "BOOK FULL",
					SourceJSON: `C:\mods\book.json`,
				},
				Name: &bookName,
			},
			{
				BaseExtractedRecord: skyrim.BaseExtractedRecord{
					ID:         "book-desc-1",
					EditorID:   &editorID,
					Type:       "BOOK DESC",
					SourceJSON: `C:\mods\book.json`,
				},
				Text: &bookText,
			},
		},
	}); err != nil {
		t.Fatalf("SaveParsedOutput failed: %v", err)
	}

	input, err := repo.LoadTerminologyInput(context.Background(), "task-book")
	if err != nil {
		t.Fatalf("LoadTerminologyInput failed: %v", err)
	}
	if len(input.Entries) != 1 {
		t.Fatalf("unexpected terminology entry count: got=%d want=1", len(input.Entries))
	}
	if input.Entries[0].RecordType != "BOOK:FULL" {
		t.Fatalf("unexpected record type: got=%q want=%q", input.Entries[0].RecordType, "BOOK:FULL")
	}
	if input.Entries[0].SourceText != bookName {
		t.Fatalf("unexpected source text: got=%q want=%q", input.Entries[0].SourceText, bookName)
	}
}

func TestRepository_LoadPersonaInput_ProjectsNPCsAndDialogues(t *testing.T) {
	db, cleanup := setupRepositoryTestDB(t)
	defer cleanup()

	repo := NewRepository(db)
	if _, err := repo.SaveParsedOutput(context.Background(), "task-persona", `C:\mods\Skyrim.esm`, buildAllSectionOutput()); err != nil {
		t.Fatalf("SaveParsedOutput failed: %v", err)
	}

	input, err := repo.LoadPersonaInput(context.Background(), "task-persona")
	if err != nil {
		t.Fatalf("LoadPersonaInput failed: %v", err)
	}

	if input.TaskID != "task-persona" {
		t.Fatalf("unexpected task id: got=%q want=%q", input.TaskID, "task-persona")
	}
	if len(input.NPCs) != 1 {
		t.Fatalf("unexpected npc count: got=%d want=1", len(input.NPCs))
	}
	npc, ok := input.NPCs["npc-1"]
	if !ok {
		t.Fatalf("expected npc keyed by source_record_id")
	}
	if npc.SourcePlugin != "Skyrim.esm" {
		t.Fatalf("unexpected npc source plugin: got=%q want=%q", npc.SourcePlugin, "Skyrim.esm")
	}
	if npc.SourceHint != "Skyrim.esm" {
		t.Fatalf("unexpected npc source hint: got=%q want=%q", npc.SourceHint, "Skyrim.esm")
	}
	if npc.NPCName != "NPC Name" {
		t.Fatalf("unexpected npc name: got=%q want=%q", npc.NPCName, "NPC Name")
	}

	if len(input.Dialogues) != 1 {
		t.Fatalf("unexpected dialogue count: got=%d want=1", len(input.Dialogues))
	}
	dialogue := input.Dialogues[0]
	if dialogue.SpeakerID != "npc_1" {
		t.Fatalf("unexpected dialogue speaker id: got=%q want=%q", dialogue.SpeakerID, "npc_1")
	}
	if dialogue.SourcePlugin != "Skyrim.esm" {
		t.Fatalf("unexpected dialogue source plugin: got=%q want=%q", dialogue.SourcePlugin, "Skyrim.esm")
	}
	if dialogue.SourceHint != "Skyrim.esm" {
		t.Fatalf("unexpected dialogue source hint: got=%q want=%q", dialogue.SourceHint, "Skyrim.esm")
	}
	if dialogue.QuestID != "QuestID01" {
		t.Fatalf("unexpected dialogue quest id: got=%q want=%q", dialogue.QuestID, "QuestID01")
	}
	if !dialogue.IsServicesBranch {
		t.Fatalf("expected services branch dialogue to be preserved")
	}
}

func TestRepository_LoadPersonaInput_PreservesHintWhenPluginCannotBeResolved(t *testing.T) {
	db, cleanup := setupRepositoryTestDB(t)
	defer cleanup()

	repo := NewRepository(db)
	if _, err := repo.SaveParsedOutput(context.Background(), "task-persona-hint", `C:\mods\persona-input.json`, buildAllSectionOutput()); err != nil {
		t.Fatalf("SaveParsedOutput failed: %v", err)
	}

	input, err := repo.LoadPersonaInput(context.Background(), "task-persona-hint")
	if err != nil {
		t.Fatalf("LoadPersonaInput failed: %v", err)
	}

	npc, ok := input.NPCs["npc-1"]
	if !ok {
		t.Fatalf("expected npc keyed by source_record_id")
	}
	if npc.SourcePlugin != "" {
		t.Fatalf("unexpected npc source plugin: got=%q want empty", npc.SourcePlugin)
	}
	if npc.SourceHint != "persona-input.json" {
		t.Fatalf("unexpected npc source hint: got=%q want=%q", npc.SourceHint, "persona-input.json")
	}

	if len(input.Dialogues) != 1 {
		t.Fatalf("unexpected dialogue count: got=%d want=1", len(input.Dialogues))
	}
	dialogue := input.Dialogues[0]
	if dialogue.SourcePlugin != "" {
		t.Fatalf("unexpected dialogue source plugin: got=%q want empty", dialogue.SourcePlugin)
	}
	if dialogue.SourceHint != "persona-input.json" {
		t.Fatalf("unexpected dialogue source hint: got=%q want=%q", dialogue.SourceHint, "persona-input.json")
	}
}

func setupRepositoryTestDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()

	db, err := sql.Open("sqlite", "file:translation_input_repo_test?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("open sqlite memory db: %v", err)
	}

	if err := Migrate(context.Background(), db); err != nil {
		_ = db.Close()
		t.Fatalf("Migrate failed: %v", err)
	}

	cleanup := func() {
		_ = db.Close()
	}
	return db, cleanup
}

func assertQueryCount(t *testing.T, db *sql.DB, query string, fileID int64, expected int) {
	t.Helper()

	var got int
	if err := db.QueryRowContext(context.Background(), query, fileID).Scan(&got); err != nil {
		t.Fatalf("count query failed: %v", err)
	}
	if got != expected {
		t.Fatalf("unexpected count for query %q: got=%d want=%d", query, got, expected)
	}
}

func buildAllSectionOutput() *skyrim.ParserOutput {
	editorID := "EditorA"
	sourcePath := `C:\mods\input.json`
	playerText := "Player"
	questID := "QuestID01"
	serviceType := "Service"
	nam1 := "NAM1"
	dialogueSource := "dialogue-source"
	prompt := "prompt"
	topic := "topic"
	menu := "menu"
	speakerID := "npc_1"
	voiceType := "FemaleNord"
	previousID := "prev"
	responseIndex := 1

	questName := "Quest Name"
	questSource := "quest-source"
	stageSource := "stage-source"
	objectiveSource := "objective-source"

	itemName := "Item Name"
	itemDescription := "Item Description"
	itemText := "Item Text"
	itemTypeHint := "Book"
	itemSource := "item-source"
	weaponName := "Weapon Name"
	weaponDescription := "Weapon Description"
	weaponText := "Weapon Text"
	weaponSource := "weapon-source"

	magicName := "Magic Name"
	magicDescription := "Magic Description"
	magicSource := "magic-source"

	locationName := "Location Name"
	locationParentID := "LocationParent"
	locationSource := "location-source"

	cellName := "Cell Name"
	cellParentID := "CellParent"
	cellSource := "cell-source"

	systemName := "System Name"
	systemDescription := "System Description"
	systemSource := "system-source"

	messageTitle := "Message Title"
	messageQuestID := "MessageQuest"
	messageSource := "message-source"

	loadScreenSource := "load-screen-source"
	loadScreenText := "Load Screen Text"

	npcName := "NPC Name"
	npcRace := "Nord"
	npcVoice := "FemaleNord"
	npcSex := "Female"
	npcClass := "Warrior"
	npcSource := "npc-source"

	return &skyrim.ParserOutput{
		DialogueGroups: []skyrim.DialogueGroup{
			{
				BaseExtractedRecord: skyrim.BaseExtractedRecord{ID: "dg-1", EditorID: &editorID, Type: "DIAL", SourceJSON: sourcePath},
				PlayerText:          &playerText,
				QuestID:             &questID,
				IsServicesBranch:    true,
				ServicesType:        &serviceType,
				NAM1:                &nam1,
				Source:              &dialogueSource,
				Responses: []skyrim.DialogueResponse{
					{
						BaseExtractedRecord: skyrim.BaseExtractedRecord{ID: "dr-1", EditorID: &editorID, Type: "INFO", SourceJSON: sourcePath},
						Text:                "Dialogue Text",
						Prompt:              &prompt,
						TopicText:           &topic,
						MenuDisplayText:     &menu,
						SpeakerID:           &speakerID,
						VoiceType:           &voiceType,
						Order:               1,
						PreviousID:          &previousID,
						Source:              &dialogueSource,
						Index:               &responseIndex,
					},
				},
			},
		},
		Quests: []skyrim.Quest{
			{
				BaseExtractedRecord: skyrim.BaseExtractedRecord{ID: "q-1", EditorID: &editorID, Type: "QUST", SourceJSON: sourcePath},
				Name:                &questName,
				Source:              &questSource,
				Stages: []skyrim.QuestStage{{
					StageIndex:     10,
					LogIndex:       1,
					Type:           "log",
					Text:           "Quest Stage Text",
					ParentID:       "q-1",
					ParentEditorID: editorID,
					Source:         &stageSource,
				}},
				Objectives: []skyrim.QuestObjective{{
					Index:          "20",
					Type:           "objective",
					Text:           "Quest Objective Text",
					ParentID:       "q-1",
					ParentEditorID: editorID,
					Source:         &objectiveSource,
				}},
			},
		},
		Items: []skyrim.Item{{
			BaseExtractedRecord: skyrim.BaseExtractedRecord{ID: "item-1", EditorID: &editorID, Type: "BOOK", SourceJSON: sourcePath},
			Name:                &itemName,
			Description:         &itemDescription,
			Text:                &itemText,
			TypeHint:            &itemTypeHint,
			Source:              &itemSource,
		}, {
			BaseExtractedRecord: skyrim.BaseExtractedRecord{ID: "item-2", EditorID: &editorID, Type: "WEAP", SourceJSON: sourcePath},
			Name:                &weaponName,
			Description:         &weaponDescription,
			Text:                &weaponText,
			Source:              &weaponSource,
		}},
		Magic: []skyrim.Magic{{
			BaseExtractedRecord: skyrim.BaseExtractedRecord{ID: "magic-1", EditorID: &editorID, Type: "SPEL", SourceJSON: sourcePath},
			Name:                &magicName,
			Description:         &magicDescription,
			Source:              &magicSource,
		}},
		Locations: []skyrim.Location{{
			BaseExtractedRecord: skyrim.BaseExtractedRecord{ID: "location-1", EditorID: &editorID, Type: "LCTN", SourceJSON: sourcePath},
			Name:                &locationName,
			ParentID:            &locationParentID,
			Source:              &locationSource,
		}},
		Cells: []skyrim.Location{{
			BaseExtractedRecord: skyrim.BaseExtractedRecord{ID: "cell-1", EditorID: &editorID, Type: "CELL", SourceJSON: sourcePath},
			Name:                &cellName,
			ParentID:            &cellParentID,
			Source:              &cellSource,
		}},
		System: []skyrim.SystemRecord{{
			BaseExtractedRecord: skyrim.BaseExtractedRecord{ID: "system-1", EditorID: &editorID, Type: "GMST", SourceJSON: sourcePath},
			Name:                &systemName,
			Description:         &systemDescription,
			Source:              &systemSource,
		}},
		Messages: []skyrim.Message{{
			BaseExtractedRecord: skyrim.BaseExtractedRecord{ID: "message-1", EditorID: &editorID, Type: "MESG", SourceJSON: sourcePath},
			Text:                "Message Text",
			Title:               &messageTitle,
			QuestID:             &messageQuestID,
			Source:              &messageSource,
		}},
		LoadScreens: []skyrim.LoadScreen{{
			BaseExtractedRecord: skyrim.BaseExtractedRecord{ID: "ls-1", EditorID: &editorID, Type: "LSCR", SourceJSON: sourcePath},
			Text:                loadScreenText,
			Source:              &loadScreenSource,
		}},
		NPCs: map[string]skyrim.NPC{
			"npc-key-1": {
				BaseExtractedRecord: skyrim.BaseExtractedRecord{ID: "npc-1", EditorID: &editorID, Type: "NPC_", SourceJSON: sourcePath},
				Name:                npcName,
				Race:                npcRace,
				Voice:               npcVoice,
				Sex:                 npcSex,
				ClassName:           &npcClass,
				Source:              &npcSource,
			},
		},
	}
}

func buildOutputWithoutEditorID() *skyrim.ParserOutput {
	sourcePath := `C:\\mods\\no-editor-id.json`
	itemName := "Item without editor id"

	return &skyrim.ParserOutput{
		Items: []skyrim.Item{{
			BaseExtractedRecord: skyrim.BaseExtractedRecord{
				ID:         "item-no-edid",
				Type:       "BOOK",
				SourceJSON: sourcePath,
			},
			Name: &itemName,
		}},
	}
}

func buildDialogueOnlyOutput(count int) *skyrim.ParserOutput {
	editorID := "EditorB"
	sourcePath := `C:\mods\dialogue-only.json`
	responses := make([]skyrim.DialogueResponse, 0, count)
	for i := 0; i < count; i++ {
		responses = append(responses, skyrim.DialogueResponse{
			BaseExtractedRecord: skyrim.BaseExtractedRecord{
				ID:         fmt.Sprintf("response-%d", i+1),
				EditorID:   &editorID,
				Type:       "INFO",
				SourceJSON: sourcePath,
			},
			Text:  fmt.Sprintf("Dialogue %03d", i+1),
			Order: i + 1,
		})
	}

	return &skyrim.ParserOutput{
		DialogueGroups: []skyrim.DialogueGroup{
			{
				BaseExtractedRecord: skyrim.BaseExtractedRecord{
					ID:         "dialogue-group-main",
					EditorID:   &editorID,
					Type:       "DIAL",
					SourceJSON: sourcePath,
				},
				Responses: responses,
			},
		},
	}
}
