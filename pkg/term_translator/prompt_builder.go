package term_translator

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"text/template"
)

// TermPromptBuilder generates prompt strings for term translation
type TermPromptBuilder interface {
	BuildPrompt(ctx context.Context, request TermTranslationRequest) (string, error)
}

// TermPromptBuilderImpl implementation
type TermPromptBuilderImpl struct {
	systemPromptTemplate *template.Template
}

// NewTermPromptBuilder creates a new TermPromptBuilder
func NewTermPromptBuilder(templateString string) (*TermPromptBuilderImpl, error) {
	slog.Debug("ENTER NewTermPromptBuilder")

	if templateString == "" {
		templateString = defaultTermPromptSystem
	}
	tmpl, err := template.New("term_prompt").Parse(templateString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse prompt template: %w", err)
	}

	return &TermPromptBuilderImpl{
		systemPromptTemplate: tmpl,
	}, nil
}

// BuildPrompt creates the full prompt for the LLM
func (b *TermPromptBuilderImpl) BuildPrompt(ctx context.Context, request TermTranslationRequest) (string, error) {
	slog.DebugContext(ctx, "ENTER TermPromptBuilderImpl.BuildPrompt", slog.String("sourceText", request.SourceText))
	defer slog.DebugContext(ctx, "EXIT TermPromptBuilderImpl.BuildPrompt")

	return b.executeTemplate(request)
}

// executeTemplate renders the prompt template with the given request data.
func (b *TermPromptBuilderImpl) executeTemplate(request TermTranslationRequest) (string, error) {
	slog.Debug("ENTER TermPromptBuilderImpl.executeTemplate")

	var buf bytes.Buffer
	err := b.systemPromptTemplate.Execute(&buf, request)
	if err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}
	return buf.String(), nil
}

const defaultTermPromptSystem = `You are a translator for a Skyrim mod.
Record Type: {{.RecordType}}
Source File: {{.SourceFile}}
Editor ID: {{.EditorID}}

Please translate the following term into Japanese:
"{{.SourceText}}"

{{- if .ShortName }}
Short Name to also translate:
"{{.ShortName}}"
{{- end }}

Context/Reference Terms from Dictionary:
{{- range .ReferenceTerms }}
- {{.Source}}: {{.Translation}}
{{- else}}
None
{{- end }}

Requirements:
1. Translate the text idiomatically for Skyrim (e.g. Katakana for names, appropriate Kanji for titles).
2. Be consistent with the Reference Terms provided.
3. You MUST output the final translation in the following exact format and nothing else:
TL: |[translated_text]|

Example:
If translating "Iron Sword", you should output:
TL: |鉄の剣|
`
