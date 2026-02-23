package translator

import (
	"context"
	"testing"
)

type mockContextEngine struct{}

func (m *mockContextEngine) BuildTranslationContext(ctx context.Context, record interface{}, input *ContextEngineInput) (*Pass2Context, []Pass2ReferenceTerm, *string, error) {
	prev := "Previous line"
	return &Pass2Context{
		PreviousLine: &prev,
	}, nil, nil, nil
}

type mockPromptBuilder struct{}

func (m *mockPromptBuilder) Build(ctx context.Context, request Pass2TranslationRequest) (string, string, error) {
	return "system", "user: " + *request.Context.PreviousLine, nil
}

type mockResumeLoader struct{}

func (m *mockResumeLoader) LoadCachedResults(pluginName string, outputBaseDir string) (map[string]TranslationResult, error) {
	return make(map[string]TranslationResult), nil
}

type mockResultWriter struct{}

func (m *mockResultWriter) Write(result TranslationResult) error { return nil }
func (m *mockResultWriter) Flush() error                         { return nil }

type mockTagProcessor struct{}

func (m *mockTagProcessor) Preprocess(text string) (string, map[string]string) { return text, nil }
func (m *mockTagProcessor) Postprocess(text string, tagMap map[string]string) string {
	return text
}
func (m *mockTagProcessor) Validate(translatedText string, tagMap map[string]string) error {
	return nil
}

func TestTranslatorSlice_ProposeJobs(t *testing.T) {
	s := NewTranslatorSlice(
		&mockContextEngine{},
		&mockPromptBuilder{},
		&mockResumeLoader{},
		&mockResultWriter{},
		&mockTagProcessor{},
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
