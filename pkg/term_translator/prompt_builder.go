package term_translator

import (
	"bytes"
	"context"
	"fmt"
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
