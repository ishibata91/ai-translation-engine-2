package models

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
