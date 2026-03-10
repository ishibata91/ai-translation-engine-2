package translator

import (
	"context"
	"testing"
)

type mockContextEngine struct {
	forced *string
}

func (m *mockContextEngine) BuildTranslationContext(ctx context.Context, record interface{}, input *ContextEngineInput) (*Pass2Context, []Pass2ReferenceTerm, *string, error) {
	prev := "Previous line"
	return &Pass2Context{
		PreviousLine: &prev,
	}, nil, m.forced, nil
}

type mockPromptBuilder struct{}

func (m *mockPromptBuilder) Build(ctx context.Context, request Pass2TranslationRequest) (string, string, error) {
	return "system", "user: " + *request.Context.PreviousLine, nil
}

type mockResumeLoader struct{}

func (m *mockResumeLoader) LoadCachedResults(pluginName string, outputBaseDir string) (map[string]TranslationResult, error) {
	return make(map[string]TranslationResult), nil
}

type mockResultWriter struct {
	writtenRecords []TranslationResult
}

func (m *mockResultWriter) Write(result TranslationResult) error {
	m.writtenRecords = append(m.writtenRecords, result)
	return nil
}
func (m *mockResultWriter) Flush() error { return nil }

type mockTagProcessor struct{}

func (m *mockTagProcessor) Preprocess(text string) (string, map[string]string) { return text, nil }
func (m *mockTagProcessor) Postprocess(text string, tagMap map[string]string) string {
	return text
}
func (m *mockTagProcessor) Validate(translatedText string, tagMap map[string]string) error {
	return nil
}

type mockBookChunker struct{}

func (m *mockBookChunker) Chunk(text string, maxCharsPerChunk int) []string { return []string{text} }

func TestTranslatorSlice_ProposeJobs(t *testing.T) {
	s := NewTranslatorSlice(
		&mockContextEngine{},
		&mockPromptBuilder{},
		&mockResumeLoader{},
		&mockResultWriter{},
		&mockTagProcessor{},
		&mockBookChunker{},
	)

	text := "Hello"
	input := TranslatorInput{
		GameData: ContextEngineInput{
			Dialogues: []ContextDialogue{
				{
					ID:   "dial_1",
					Text: &text,
					Type: "INFO",
				},
			},
		},
		OutputConfig: BatchConfig{
			PluginName: "TestPlugin",
			MaxTokens:  1000,
		},
	}

	reqs, err := s.ProposeJobs(context.Background(), input)
	if err != nil {
		t.Fatalf("ProposeJobs failed: %v", err)
	}

	if len(reqs) != 1 {
		t.Fatalf("expected 1 request, got %d", len(reqs))
	}

	expectedUserPrompt := "user: Previous line"
	if reqs[0].UserPrompt != expectedUserPrompt {
		t.Errorf("expected user prompt %q, got %q", expectedUserPrompt, reqs[0].UserPrompt)
	}
}

func TestTranslatorSlice_ProposeJobs_ForcedTranslation(t *testing.T) {
	forced := "こんにちは"
	writer := &mockResultWriter{}
	s := NewTranslatorSlice(
		&mockContextEngine{forced: &forced},
		&mockPromptBuilder{},
		&mockResumeLoader{},
		writer,
		&mockTagProcessor{},
		&mockBookChunker{},
	)

	text := "Hello"
	input := TranslatorInput{
		GameData: ContextEngineInput{
			Dialogues: []ContextDialogue{
				{
					ID:   "dial_1",
					Text: &text,
					Type: "INFO",
				},
			},
		},
		OutputConfig: BatchConfig{
			PluginName: "TestPlugin",
			MaxTokens:  1000,
		},
	}

	reqs, err := s.ProposeJobs(context.Background(), input)
	if err != nil {
		t.Fatalf("ProposeJobs failed: %v", err)
	}

	// Should not generate any LLM requests
	if len(reqs) != 0 {
		t.Errorf("expected 0 requests for forced translation, got %d", len(reqs))
	}

	// Should have written result to DB
	if len(writer.writtenRecords) != 1 {
		t.Errorf("expected 1 written record, got %d", len(writer.writtenRecords))
	} else {
		res := writer.writtenRecords[0]
		if res.ID != "dial_1" || *res.TranslatedText != forced || res.Status != "completed" {
			t.Errorf("unexpected written record: %+v", res)
		}
	}
}
