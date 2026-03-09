package parser

// ParserOutput is the root container for all extracted records.
// It replaces the domain/models.ExtractedData to isolate the loader slice.
type ParserOutput struct {
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

// BaseExtractedRecord is embedded in all domain models.
type BaseExtractedRecord struct {
	ID         string  `json:"id"`
	EditorID   *string `json:"editor_id,omitempty"`
	Type       string  `json:"type"`
	SourceJSON string  `json:"source_json"`
}

// DialogueResponse represents a single dialogue response line.
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

// DialogueGroup represents a group of dialogue responses.
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

// NPC represents a Non-Player Character record.
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
func (n *NPC) IsFemale() bool {
	return n.Sex == "Female" || n.Sex == "female"
}

// Item represents an item or book record.
type Item struct {
	BaseExtractedRecord
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Text        *string `json:"text,omitempty"`
	TypeHint    *string `json:"type_hint,omitempty"`
	Source      *string `json:"source,omitempty"`
}

// Magic represents a spell, enchantment, or potion effect.
type Magic struct {
	BaseExtractedRecord
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Source      *string `json:"source,omitempty"`
}

// QuestStage represents a quest stage description.
type QuestStage struct {
	StageIndex     int     `json:"stage_index"`
	LogIndex       int     `json:"log_index"`
	Type           string  `json:"type"`
	Text           string  `json:"text"`
	ParentID       string  `json:"parent_id"`
	ParentEditorID string  `json:"parent_editor_id"`
	Source         *string `json:"source,omitempty"`
}

// QuestObjective represents a quest objective text.
type QuestObjective struct {
	Index          string  `json:"index"`
	Type           string  `json:"type"`
	Text           string  `json:"text"`
	ParentID       string  `json:"parent_id"`
	ParentEditorID string  `json:"parent_editor_id"`
	Source         *string `json:"source,omitempty"`
}

// Quest represents a quest record with its stages and objectives.
type Quest struct {
	BaseExtractedRecord
	Name       *string          `json:"name,omitempty"`
	Stages     []QuestStage     `json:"stages"`
	Objectives []QuestObjective `json:"objectives"`
	Source     *string          `json:"source,omitempty"`
}

// Location represents a cell or world space.
type Location struct {
	BaseExtractedRecord
	Name     *string `json:"name,omitempty"`
	ParentID *string `json:"parent_id,omitempty"`
	Source   *string `json:"source,omitempty"`
}

// Message represents a message box text.
type Message struct {
	BaseExtractedRecord
	Text    string  `json:"text"`
	Title   *string `json:"title,omitempty"`
	QuestID *string `json:"quest_id,omitempty"`
	Source  *string `json:"source,omitempty"`
}

// SystemRecord represents a system setting like GMST.
type SystemRecord struct {
	BaseExtractedRecord
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	Source      *string `json:"source,omitempty"`
}

// LoadScreen represents a loading screen text.
type LoadScreen struct {
	BaseExtractedRecord
	Text   string  `json:"text"`
	Source *string `json:"source,omitempty"`
}
