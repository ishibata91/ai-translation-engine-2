package llm

import "context"

// LLMClient defines the core interface for LLM request execution.
// All LLM providers (Gemini, OpenAI, Local/GGUF, etc.) implement this interface.
type LLMClient interface {
	Complete(ctx context.Context, req Request) (Response, error)
	StreamComplete(ctx context.Context, req Request) (StreamResponse, error)
	GetEmbedding(ctx context.Context, text string) ([]float32, error)
	HealthCheck(ctx context.Context) error
}

// LLMManager manages available LLM providers and creates client instances
// based on the current configuration (provider selection, API key, endpoint, etc.).
type LLMManager interface {
	GetClient(ctx context.Context, config LLMConfig) (LLMClient, error)
	GetBatchClient(ctx context.Context, config LLMConfig) (BatchClient, error)
	ResolveBulkStrategy(ctx context.Context, strategy BulkStrategy, provider string) BulkStrategy
}

// BatchClient abstracts asynchronous batch API job management and result retrieval.
type BatchClient interface {
	SubmitBatch(ctx context.Context, reqs []Request) (BatchJobID, error)
	GetBatchStatus(ctx context.Context, id BatchJobID) (BatchStatus, error)
	GetBatchResults(ctx context.Context, id BatchJobID) ([]Response, error)
}

// StreamResponse provides an iterator interface for streaming LLM responses.
type StreamResponse interface {
	Next() (Response, bool)
	Close() error
}
