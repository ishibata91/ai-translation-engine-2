package models

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
