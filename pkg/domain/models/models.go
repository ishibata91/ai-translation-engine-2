package models

// BaseExtractedRecord is embedded in all domain models.
type BaseExtractedRecord struct {
	ID         string  `json:"id"`
	EditorID   *string `json:"editor_id,omitempty"`
	Type       string  `json:"type"`
	SourceJSON string  `json:"source_json"`
}

// ---------------------------------------------------------------------
// Dialogue Models
// ---------------------------------------------------------------------

type DialogueResponse struct {
	BaseExtractedRecord
	Text            string  `json:"text"`
	Prompt          *string `json:"prompt,omitempty"`
	TopicText       *string `json:"topic_text,omitempty"`
	MenuDisplayText *string `json:"menu_display_text,omitempty"`
	SpeakerID       *string `json:"speaker_id,omitempty"`
	VoiceType       *string `json:"voice_type,omitempty"`
	Order           int     `json:"order"`
	PreviousID      *string `json:"previous_id,omitempty"`
	Source          *string `json:"source,omitempty"`
	Index           *int    `json:"index,omitempty"`
}

type DialogueGroup struct {
	BaseExtractedRecord
	PlayerText       *string            `json:"player_text,omitempty"`
	QuestID          *string            `json:"quest_id,omitempty"`
	IsServicesBranch bool               `json:"is_services_branch"`
	ServicesType     *string            `json:"services_type,omitempty"`
	NAM1             *string            `json:"nam1,omitempty"`
	Responses        []DialogueResponse `json:"responses"`
	Source           *string            `json:"source,omitempty"`
}

// ---------------------------------------------------------------------
// Quest Models
// ---------------------------------------------------------------------

type QuestStage struct {
	Index  int     `json:"index"`
	Type   string  `json:"type"`
	Text   string  `json:"text"`
	Source *string `json:"source,omitempty"`
}

type QuestObjective struct {
	Index  int     `json:"index"`
	Type   string  `json:"type"`
	Text   string  `json:"text"`
	Source *string `json:"source,omitempty"`
}

type Quest struct {
	BaseExtractedRecord
	Name       *string          `json:"name,omitempty"`
	Stages     []QuestStage     `json:"stages"`
	Objectives []QuestObjective `json:"objectives"`
	Source     *string          `json:"source,omitempty"`
}

// ---------------------------------------------------------------------
// Entity Models
// ---------------------------------------------------------------------

type NPC struct {
	BaseExtractedRecord
	Name      string  `json:"name"`
	Race      string  `json:"race"`
	Voice     string  `json:"voice"`
	Sex       string  `json:"sex"`
	ClassName *string `json:"class_name,omitempty"`
	Source    *string `json:"source,omitempty"`
}

// IsFemale returns true if the NPC is female, false otherwise.
// Case-insensitive check for "Female".
func (n *NPC) IsFemale() bool {
	// Simple check, can be expanded to full case-insensitive if needed,
	// but strictly following spec: Female or female returns true.
	return n.Sex == "Female" || n.Sex == "female"
}

type Item struct {
	BaseExtractedRecord
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Text        *string `json:"text,omitempty"`
	TypeHint    *string `json:"type_hint,omitempty"`
	Source      *string `json:"source,omitempty"`
}

type Magic struct {
	BaseExtractedRecord
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Source      *string `json:"source,omitempty"`
}

// ---------------------------------------------------------------------
// Other Models
// ---------------------------------------------------------------------

type Location struct {
	BaseExtractedRecord
	Name     *string `json:"name,omitempty"`
	ParentID *string `json:"parent_id,omitempty"`
	Source   *string `json:"source,omitempty"`
}

type Message struct {
	BaseExtractedRecord
	Text    string  `json:"text"`
	Title   *string `json:"title,omitempty"`
	QuestID *string `json:"quest_id,omitempty"`
	Source  *string `json:"source,omitempty"`
}

type SystemRecord struct {
	BaseExtractedRecord
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Source      *string `json:"source,omitempty"`
}

type LoadScreen struct {
	BaseExtractedRecord
	Text   string  `json:"text"`
	Source *string `json:"source,omitempty"`
}

// ---------------------------------------------------------------------
// Root Container
// ---------------------------------------------------------------------

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
