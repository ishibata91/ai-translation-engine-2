package llm_client

// Request represents a request to an LLM provider.
type Request struct {
	SystemPrompt   string                 `json:"system_prompt"`
	UserPrompt     string                 `json:"user_prompt"`
	MaxTokens      int                    `json:"max_tokens"`
	Temperature    float32                `json:"temperature"`
	ResponseSchema map[string]interface{} `json:"response_schema,omitempty"`
	StopSequences  []string               `json:"stop_sequences,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// Response represents a response from an LLM provider.
type Response struct {
	Content  string                 `json:"content"`
	Success  bool                   `json:"success"`
	Error    string                 `json:"error,omitempty"`
	Usage    TokenUsage             `json:"usage"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// TokenUsage represents token consumption for a request.
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// BulkStrategy represents the bulk processing strategy for LLM requests.
type BulkStrategy string

const (
	// BulkStrategyBatch uses provider-native async Batch API (e.g., Gemini, xAI).
	BulkStrategyBatch BulkStrategy = "batch"
	// BulkStrategySync uses ExecuteBulkSync for synchronous concurrent processing.
	// This is the only option for local LLM providers.
	BulkStrategySync BulkStrategy = "sync"
)

// ConfigStore keys for LLM bulk sync settings.
const (
	// LLMConfigNamespace is the ConfigStore namespace for LLM-related settings.
	LLMConfigNamespace = "llm"
	// LLMBulkStrategyKey is the ConfigStore key for the bulk processing strategy.
	// Values: "batch" or "sync".
	LLMBulkStrategyKey = "bulk_strategy"
	// LLMSyncConcurrencyKeySuffix is the suffix appended to the provider name
	// to build the ConfigStore key for sync concurrency (e.g., "sync_concurrency.gemini").
	LLMSyncConcurrencyKeySuffix = "sync_concurrency"
)

// DefaultConcurrency returns the default concurrency for a given provider.
// Local LLM providers default to 1; cloud providers default to 5.
func DefaultConcurrency(provider string) int {
	switch provider {
	case "local":
		return 1
	default:
		return 5
	}
}

// LLMConfig holds configuration for a specific LLM provider instance.
type LLMConfig struct {
	Provider   string                 `json:"provider"`
	APIKey     string                 `json:"api_key"`
	Endpoint   string                 `json:"endpoint"`
	Model      string                 `json:"model"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	// Concurrency controls how many parallel workers are used in ExecuteBulkSync.
	// Only relevant when BulkStrategy is BulkStrategySync.
	Concurrency int `json:"concurrency,omitempty"`
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
