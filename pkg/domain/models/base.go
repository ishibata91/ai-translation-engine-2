package models

// BaseExtractedRecord is embedded in all domain models.
type BaseExtractedRecord struct {
	ID         string  `json:"id"`
	EditorID   *string `json:"editor_id,omitempty"`
	Type       string  `json:"type"`
	SourceJSON string  `json:"source_json"`
}

// ExtractedData is the root container for all extracted records.
type ExtractedData struct {
	DialogueGroups []DialogueGroup `json:"dialogue_groups"`
	Quests         []Quest         `json:"quests"`
	Items          []Item          `json:"items"`
	Magic          []Magic         `json:"magic"`
	Locations      []Location      `json:"locations"`
	Cells          []Location      `json:"cells"`
	System         []SystemRecord  `json:"system"`
	Messages       []Message       `json:"messages"`
	LoadScreens    []LoadScreen    `json:"load_screens"`
	NPCs           map[string]NPC  `json:"npcs"`
	SourceJSON     string          `json:"-"` // Metadata, not in JSON
}
