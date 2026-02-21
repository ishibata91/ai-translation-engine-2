package term_translator

// TermTranslationRequest represents a single term translation request.
type TermTranslationRequest struct {
	FormID         string          `json:"form_id"`
	EditorID       string          `json:"editor_id"`
	RecordType     string          `json:"record_type"`
	SourceText     string          `json:"source_text"`
	ShortName      string          `json:"short_name,omitempty"`
	SourcePlugin   string          `json:"source_plugin"`
	SourceFile     string          `json:"source_file"`
	ReferenceTerms []ReferenceTerm `json:"reference_terms,omitempty"`
}

// ReferenceTerm represents a reference term (existing translation from dictionary).
type ReferenceTerm struct {
	Source      string `json:"source"`
	Translation string `json:"translation"`
}

// TermTranslationResult represents the result of a single term translation.
type TermTranslationResult struct {
	FormID         string `json:"form_id"`
	EditorID       string `json:"editor_id"`
	RecordType     string `json:"record_type"`
	SourceText     string `json:"source_text"`
	TranslatedText string `json:"translated_text"`
	SourcePlugin   string `json:"source_plugin"`
	SourceFile     string `json:"source_file"`
	Status         string `json:"status"`
	ErrorMessage   string `json:"error_message,omitempty"`
}

// NPCPair pairs a FULL name NPC record with its optional SHRT name record
// for simultaneous translation.
type NPCPair struct {
	Full  interface{} `json:"-"` // *models.NPC — typed at implementation
	Short interface{} `json:"-"` // *models.NPC — typed at implementation (may be nil)
}

// TermRecordConfig defines which record types are targets for term translation.
// Injected via DI from Config Store.
type TermRecordConfig struct {
	TargetRecordTypes []string `json:"target_record_types"`
}

// IsTarget returns true if the given record type is a term translation target.
func (c *TermRecordConfig) IsTarget(recordType string) bool {
	for _, t := range c.TargetRecordTypes {
		if t == recordType {
			return true
		}
	}
	return false
}
