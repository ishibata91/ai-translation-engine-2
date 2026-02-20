package models

// QuestStage represents a quest stage description.
type QuestStage struct {
	Index  int     `json:"index"`
	Type   string  `json:"type"`
	Text   string  `json:"text"`
	Source *string `json:"source,omitempty"`
}

// QuestObjective represents a quest objective text.
type QuestObjective struct {
	Index  int     `json:"index"`
	Type   string  `json:"type"`
	Text   string  `json:"text"`
	Source *string `json:"source,omitempty"`
}

// Quest represents a quest record with its stages and objectives.
type Quest struct {
	BaseExtractedRecord
	Name       *string          `json:"name,omitempty"`
	Stages     []QuestStage     `json:"stages"`
	Objectives []QuestObjective `json:"objectives"`
	Source     *string          `json:"source,omitempty"`
}
