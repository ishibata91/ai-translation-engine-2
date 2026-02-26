package translator

import (
	"context"
	"testing"

	"github.com/ishibata91/ai-translation-engine-2/pkg/infrastructure/llm"
)

func TestTranslatorSlice_SaveResults(t *testing.T) {
	writer := &mockResultWriter{}
	s := NewTranslatorSlice(
		&mockContextEngine{},
		&mockPromptBuilder{},
		&mockResumeLoader{},
		writer,
		&mockTagProcessor{},
		&mockBookChunker{},
	)

	responses := []llm.Response{
		{
			Content: "こんにちは",
			Metadata: map[string]interface{}{
				"id":            "dial_1",
				"record_type":   "INFO",
				"source_plugin": "TestPlugin",
				"tags":          map[string]string{},
				"chunk_index":   0,
				"is_chunked":    false,
			},
		},
	}

	err := s.SaveResults(context.Background(), responses)
	if err != nil {
		t.Fatalf("SaveResults failed: %v", err)
	}

	if len(writer.writtenRecords) != 1 {
		t.Errorf("expected 1 written record, got %d", len(writer.writtenRecords))
	}
}
