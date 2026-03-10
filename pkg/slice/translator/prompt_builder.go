package translator

import (
	"context"
	"fmt"
	"strings"
)

type defaultPromptBuilder struct{}

// NewDefaultPromptBuilder creates a new PromptBuilder instance.
func NewDefaultPromptBuilder() PromptBuilder {
	return &defaultPromptBuilder{}
}

func (b *defaultPromptBuilder) Build(ctx context.Context, req Pass2TranslationRequest) (string, string, error) {
	systemPrompt := "あなたはプロのゲーム翻訳者です。提供された文脈と用語集を参考に、自然な日本語に翻訳してください。"

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("原文: %s\n\n", req.SourceText))

	if req.Context.ModDescription != nil {
		sb.WriteString(fmt.Sprintf("Mod概要: %s\n", *req.Context.ModDescription))
	}

	if req.Context.Speaker != nil {
		sb.WriteString(fmt.Sprintf("話者: %s\n", req.Context.Speaker.Name))
		if req.Context.Speaker.PersonaText != nil {
			sb.WriteString(fmt.Sprintf("話者の性格: %s\n", *req.Context.Speaker.PersonaText))
		}
	}

	if len(req.ReferenceTerms) > 0 {
		sb.WriteString("用語解説:\n")
		for _, term := range req.ReferenceTerms {
			sb.WriteString(fmt.Sprintf("- %s: %s\n", term.OriginalEN, term.OriginalJA))
		}
	}

	return systemPrompt, sb.String(), nil
}
