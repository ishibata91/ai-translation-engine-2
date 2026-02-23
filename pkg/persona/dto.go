package persona

// NPCDialogueData holds collected dialogue data for a single NPC.
type NPCDialogueData struct {
	SpeakerID string          `json:"speaker_id"`
	EditorID  string          `json:"editor_id"`
	NPCName   string          `json:"npc_name"`
	Race      string          `json:"race"`
	Sex       string          `json:"sex"`
	VoiceType string          `json:"voice_type"`
	Dialogues []DialogueEntry `json:"dialogues"`
}

// DialogueEntry represents a single dialogue line used for persona generation.
type DialogueEntry struct {
	Text              string `json:"text"`
	EnglishText       string `json:"english_text"`
	QuestID           string `json:"quest_id,omitempty"`
	IsServicesBranch  bool   `json:"is_services_branch"`
	Order             int    `json:"order"`
}

// ScoredDialogueEntry pairs a DialogueEntry with its computed importance score.
type ScoredDialogueEntry struct {
	Entry             DialogueEntry `json:"entry"`
	ImportanceScore   int           `json:"importance_score"`
	ProperNounHits    int           `json:"proper_noun_hits"`
	EmotionIndicators int           `json:"emotion_indicators"`
	BasePriority      int           `json:"base_priority"`
}

// PersonaResult represents the result of persona generation for a single NPC.
type PersonaResult struct {
	SpeakerID       string `json:"speaker_id"`
	EditorID        string `json:"editor_id"`
	NPCName         string `json:"npc_name"`
	Race            string `json:"race"`
	Sex             string `json:"sex"`
	VoiceType       string `json:"voice_type"`
	PersonaText     string `json:"persona_text"`
	DialogueCount   int    `json:"dialogue_count"`
	EstimatedTokens int    `json:"estimated_tokens"`
	SourcePlugin    string `json:"source_plugin"`
	Status          string `json:"status"`
	ErrorMessage    string `json:"error_message,omitempty"`
}

// TokenEstimation holds the result of token usage estimation.
type TokenEstimation struct {
	InputTokens  int  `json:"input_tokens"`
	OutputTokens int  `json:"output_tokens"`
	TotalTokens  int  `json:"total_tokens"`
	ExceedsLimit bool `json:"exceeds_limit"`
}

// ScoringConfig holds configuration for importance scoring weights.
type ScoringConfig struct {
	NounWeight      int      `json:"noun_weight"`
	EmotionWeight   int      `json:"emotion_weight"`
	StrongWordsList []string `json:"strong_words_list"`
}

// PersonaConfig holds configuration for persona generation.
type PersonaConfig struct {
	MaxDialoguesPerNPC   int `json:"max_dialogues_per_npc"`
	ContextWindowLimit   int `json:"context_window_limit"`
	SystemPromptOverhead int `json:"system_prompt_overhead"`
	MaxOutputTokens      int `json:"max_output_tokens"`
	MinDialogueThreshold int `json:"min_dialogue_threshold"`
	MaxWorkers           int `json:"max_workers"`
}
