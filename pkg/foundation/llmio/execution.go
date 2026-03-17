package llmio

// ExecutionConfig represents workflow-independent runtime settings for LLM execution.
type ExecutionConfig struct {
	Provider        string
	Model           string
	Endpoint        string
	APIKey          string
	Temperature     float32
	ContextLength   int
	SyncConcurrency int
	BulkStrategy    string
}
