package llmio

// Request represents an LLM request payload shared between workflow and slices.
type Request struct {
	SystemPrompt   string                 `json:"system_prompt"`
	UserPrompt     string                 `json:"user_prompt"`
	Temperature    float32                `json:"temperature"`
	ResponseSchema map[string]interface{} `json:"response_schema,omitempty"`
	StopSequences  []string               `json:"stop_sequences,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// Response represents an LLM response payload shared between workflow and slices.
type Response struct {
	Content  string                 `json:"content"`
	Success  bool                   `json:"success"`
	Error    string                 `json:"error,omitempty"`
	Usage    TokenUsage             `json:"usage"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// TokenUsage represents token consumption data for one request/response pair.
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}
