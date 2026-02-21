package llm_client

// Request represents a request to an LLM provider.
type Request struct {
	SystemPrompt   string                 `json:"system_prompt"`
	UserPrompt     string                 `json:"user_prompt"`
	MaxTokens      int                    `json:"max_tokens"`
	Temperature    float32                `json:"temperature"`
	ResponseSchema map[string]interface{} `json:"response_schema,omitempty"`
	StopSequences  []string               `json:"stop_sequences,omitempty"`
}

// Response represents a response from an LLM provider.
type Response struct {
	Content string     `json:"content"`
	Success bool       `json:"success"`
	Error   string     `json:"error,omitempty"`
	Usage   TokenUsage `json:"usage"`
}

// TokenUsage represents token consumption for a request.
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// LLMConfig holds configuration for a specific LLM provider instance.
type LLMConfig struct {
	Provider   string                 `json:"provider"`
	APIKey     string                 `json:"api_key"`
	Endpoint   string                 `json:"endpoint"`
	Model      string                 `json:"model"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// BatchJobID identifies a batch processing job.
type BatchJobID struct {
	ID       string `json:"id"`
	Provider string `json:"provider"`
}

// BatchStatus represents the current status of a batch job.
type BatchStatus struct {
	ID       string  `json:"id"`
	State    string  `json:"state"`
	Progress float32 `json:"progress"`
}
