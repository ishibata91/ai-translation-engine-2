package models

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
// Case-insensitive check for "Female".
func (n *NPC) IsFemale() bool {
	// Simple check, can be expanded to full case-insensitive if needed,
	// but strictly following spec: Female or female returns true.
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
