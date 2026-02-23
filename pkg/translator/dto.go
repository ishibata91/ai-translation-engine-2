package translator

// TranslationResult represents the result of translating a single record.
type TranslationResult struct {
	ID             string  `json:"id"`
	RecordType     string  `json:"type"`
	SourceText     string  `json:"source_text"`
	TranslatedText *string `json:"translated_text,omitempty"`
	Index          *int    `json:"index,omitempty"`
	Status         string  `json:"status"`
	ErrorMessage   *string `json:"error_message,omitempty"`
	SourcePlugin   string  `json:"source_plugin"`
	SourceFile     string  `json:"source_file"`
	EditorID       *string `json:"editor_id,omitempty"`
	ParentID       *string `json:"parent_id,omitempty"`
	ParentEditorID *string `json:"parent_editor_id,omitempty"`
}

// Pass2TranslationRequest represents a single translation unit for Pass 2.
type Pass2TranslationRequest struct {
	ID                string               `json:"id"`
	RecordType        string               `json:"record_type"`
	SourceText        string               `json:"source_text"`
	Context           Pass2Context         `json:"context"`
	Index             *int                 `json:"index,omitempty"`
	ReferenceTerms    []Pass2ReferenceTerm `json:"reference_terms,omitempty"`
	EditorID          *string              `json:"editor_id,omitempty"`
	ParentID          *string              `json:"parent_id,omitempty"`
	ParentEditorID    *string              `json:"parent_editor_id,omitempty"`
	ForcedTranslation *string              `json:"forced_translation,omitempty"`
	SourcePlugin      string               `json:"source_plugin"`
	SourceFile        string               `json:"source_file"`
	MaxTokens         *int                 `json:"max_tokens,omitempty"`
}

// Pass2Context holds contextual information needed for high-quality translation.
type Pass2Context struct {
	PreviousLine    *string              `json:"previous_line,omitempty"`
	Speaker         *Pass2SpeakerProfile `json:"speaker,omitempty"`
	TopicName       *string              `json:"topic_name,omitempty"`
	QuestName       *string              `json:"quest_name,omitempty"`
	QuestSummary    *string              `json:"quest_summary,omitempty"`
	DialogueSummary *string              `json:"dialogue_summary,omitempty"`
	ItemTypeHint    *string              `json:"item_type_hint,omitempty"`
	ModDescription  *string              `json:"mod_description,omitempty"`
	PlayerTone      *string              `json:"player_tone,omitempty"`
}

// Pass2SpeakerProfile represents NPC speaker attributes for translation context.
type Pass2SpeakerProfile struct {
	Name            string  `json:"name"`
	Gender          string  `json:"gender"`
	Race            string  `json:"race"`
	VoiceType       string  `json:"voice_type"`
	ToneInstruction string  `json:"tone_instruction"`
	PersonaText     *string `json:"persona_text,omitempty"`
}

// Pass2ReferenceTerm represents a reference term (existing translation from dictionary).
type Pass2ReferenceTerm struct {
	OriginalEN string `json:"original_en"`
	OriginalJA string `json:"original_ja"`
}

// BatchConfig holds configuration for batch translation execution.
type BatchConfig struct {
	MaxWorkers     int     `json:"max_workers"`
	TimeoutSeconds float64 `json:"timeout_seconds"`
	MaxTokens      int     `json:"max_tokens"`
	OutputBaseDir  string  `json:"output_base_dir"`
	PluginName     string  `json:"plugin_name"`
}

// RetryConfig holds configuration for exponential backoff retry.
type RetryConfig struct {
	MaxRetries       int     `json:"max_retries"`
	BaseDelaySeconds float64 `json:"base_delay_seconds"`
	MaxDelaySeconds  float64 `json:"max_delay_seconds"`
	ExponentialBase  float64 `json:"exponential_base"`
}

// TagHallucinationError indicates that the LLM generated tags not present in the original text.
type TagHallucinationError struct {
	Message string
}

// Error implements the error interface.
func (e *TagHallucinationError) Error() string {
	return e.Message
}
