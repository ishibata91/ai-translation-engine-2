package context_engine

// TranslationRequest represents a single translation unit for Pass 2.
type TranslationRequest struct {
	ID                string             `json:"id"`
	RecordType        string             `json:"record_type"`
	SourceText        string             `json:"source_text"`
	Context           TranslationContext `json:"context"`
	Index             *int               `json:"index,omitempty"`
	ReferenceTerms    []ReferenceTerm    `json:"reference_terms,omitempty"`
	EditorID          *string            `json:"editor_id,omitempty"`
	ForcedTranslation *string            `json:"forced_translation,omitempty"`
	SourcePlugin      string             `json:"source_plugin"`
	SourceFile        string             `json:"source_file"`
	MaxTokens         *int               `json:"max_tokens,omitempty"`
}

// TranslationContext holds contextual information needed for high-quality translation.
type TranslationContext struct {
	PreviousLine    *string         `json:"previous_line,omitempty"`
	Speaker         *SpeakerProfile `json:"speaker,omitempty"`
	TopicName       *string         `json:"topic_name,omitempty"`
	QuestName       *string         `json:"quest_name,omitempty"`
	QuestSummary    *string         `json:"quest_summary,omitempty"`
	DialogueSummary *string         `json:"dialogue_summary,omitempty"`
	ItemTypeHint    *string         `json:"item_type_hint,omitempty"`
	ModDescription  *string         `json:"mod_description,omitempty"`
	PlayerTone      *string         `json:"player_tone,omitempty"`
}

// SpeakerProfile represents NPC speaker attributes for translation context.
type SpeakerProfile struct {
	Name            string  `json:"name"`
	Gender          string  `json:"gender"`
	Race            string  `json:"race"`
	VoiceType       string  `json:"voice_type"`
	ToneInstruction string  `json:"tone_instruction"`
	PersonaText     *string `json:"persona_text,omitempty"`
}

// ReferenceTerm represents a reference term (existing translation from dictionary).
type ReferenceTerm struct {
	OriginalEN string `json:"original_en"`
	OriginalJA string `json:"original_ja"`
}

// ContextEngineConfig holds runtime configuration for the Context Engine.
type ContextEngineConfig struct {
	ModDescription string `json:"mod_description"`
	PlayerTone     string `json:"player_tone"`
	SourceFile     string `json:"source_file"`
}
