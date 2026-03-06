package llm

import "errors"

var (
	// ErrStructuredOutputNotSupported is returned when provider does not implement structured output.
	ErrStructuredOutputNotSupported = errors.New("llm: structured output not supported by provider")
	// ErrModelRequired is returned when model is omitted in configuration.
	ErrModelRequired = errors.New("llm: model must be specified")
)
