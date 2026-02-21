package main_translator

// TranslationResult represents the result of translating a single record.
type TranslationResult struct {
	ID             string  `json:"id"`
	RecordType     string  `json:"record_type"`
	SourceText     string  `json:"source_text"`
	TranslatedText *string `json:"translated_text,omitempty"`
	Index          *int    `json:"index,omitempty"`
	Status         string  `json:"status"`
	ErrorMessage   *string `json:"error_message,omitempty"`
	SourcePlugin   string  `json:"source_plugin"`
	SourceFile     string  `json:"source_file"`
	EditorID       *string `json:"editor_id,omitempty"`
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
