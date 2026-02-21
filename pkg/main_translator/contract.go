package main_translator

import (
	"context"

	ce "github.com/ishibata91/ai-translation-engine-2/pkg/context_engine"
)

// BatchTranslator translates multiple TranslationRequests in parallel batches.
type BatchTranslator interface {
	TranslateBatch(ctx context.Context, requests []ce.TranslationRequest, config BatchConfig) ([]TranslationResult, error)
}

// Translator translates a single TranslationRequest via LLM.
type Translator interface {
	Translate(ctx context.Context, request ce.TranslationRequest) (TranslationResult, error)
}

// PromptBuilder constructs system and user prompts from a TranslationRequest.
type PromptBuilder interface {
	Build(request ce.TranslationRequest) (systemPrompt string, userPrompt string, err error)
}

// TagProcessor handles HTML tag abstraction before translation
// and restoration after translation.
type TagProcessor interface {
	Preprocess(text string) (processedText string, tagMap map[string]string)
	Postprocess(text string, tagMap map[string]string) string
	Validate(translatedText string, tagMap map[string]string) error
}

// TranslationVerifier validates the quality of a translation result.
type TranslationVerifier interface {
	Verify(sourceText string, translatedText string, tagMap map[string]string) error
}

// BookChunker splits long book text into chunks for individual translation.
type BookChunker interface {
	Chunk(text string, maxTokensPerChunk int) []string
}

// ResultWriter persists translation results incrementally to prevent data loss.
type ResultWriter interface {
	Write(result TranslationResult) error
	Flush() error
}

// ResumeLoader loads previously completed translation results for delta updates.
type ResumeLoader interface {
	LoadCachedResults(config BatchConfig) (map[string]TranslationResult, error)
}
