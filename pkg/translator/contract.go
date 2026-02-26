package translator

import (
	"context"

	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/llm"
)

// TranslatorSlice is the main interface for the Pass 2 Translator vertical slice.
type TranslatorSlice interface {
	// ID returns the unique identifier of the slice.
	ID() string

	// PreparePrompts (Phase 1) generates LLM requests.
	PreparePrompts(ctx context.Context, input any) ([]llm.Request, error)

	// SaveResults (Phase 2) persists LLM responses.
	SaveResults(ctx context.Context, responses []llm.Response) error

	// ProposeJobs analyzes input game data and proposes LLM translation jobs.
	ProposeJobs(ctx context.Context, input TranslatorInput) ([]llm.Request, error)
}

// Internal components

// Translator translates a single record via LLM.
type Translator interface {
	Translate(ctx context.Context, request Pass2TranslationRequest) (TranslationResult, error)
}

// PromptBuilder constructs system and user prompts.
type PromptBuilder interface {
	Build(ctx context.Context, request Pass2TranslationRequest) (systemPrompt string, userPrompt string, err error)
}

// TagProcessor handles HTML tag abstraction and restoration.
type TagProcessor interface {
	Preprocess(text string) (processedText string, tagMap map[string]string)
	Postprocess(text string, tagMap map[string]string) string
	Validate(translatedText string, tagMap map[string]string) error
}

// BookChunker splits long book text into chunks.
type BookChunker interface {
	Chunk(text string, maxCharsPerChunk int) []string
}

// ResultWriter persists translation results.
type ResultWriter interface {
	Write(result TranslationResult) error
	Flush() error
}

// ResumeLoader loads previously completed translation results.
type ResumeLoader interface {
	LoadCachedResults(pluginName string, outputBaseDir string) (map[string]TranslationResult, error)
}
