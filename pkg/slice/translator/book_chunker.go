package translator

import (
	"strings"
)

type bookChunker struct{}

// NewBookChunker creates a new BookChunker instance.
func NewBookChunker() BookChunker {
	return &bookChunker{}
}

// Chunk splits a long text into multiple chunks, attempting to stay within maxChars limit.
// It tries to split at line breaks or sentence ends.
func (c *bookChunker) Chunk(text string, maxChars int) []string {
	if len(text) <= maxChars {
		return []string{text}
	}

	var chunks []string
	remaining := text

	for len(remaining) > 0 {
		if len(remaining) <= maxChars {
			chunks = append(chunks, remaining)
			break
		}

		// Find a good split point within maxChars
		splitPoint := maxChars

		// 1. Try to find the last newline within limits
		lastNewline := strings.LastIndex(remaining[:maxChars], "\n")
		if lastNewline > maxChars/2 {
			splitPoint = lastNewline + 1
		} else {
			// 2. Try to find the last sentence end (., !, ?)
			lastSentenceEnd := -1
			ends := []string{". ", "! ", "? ", "。 ", "！ ", "？ "}
			for _, end := range ends {
				idx := strings.LastIndex(remaining[:maxChars], end)
				if idx > lastSentenceEnd {
					lastSentenceEnd = idx + len(end)
				}
			}

			if lastSentenceEnd > maxChars/2 {
				splitPoint = lastSentenceEnd
			} else {
				// 3. Just find the last space
				lastSpace := strings.LastIndex(remaining[:maxChars], " ")
				if lastSpace > maxChars/2 {
					splitPoint = lastSpace + 1
				}
			}
		}

		// Ensure we don't split in the middle of an HTML tag or placeholder
		// Check if we are inside a tag at splitPoint
		if isInsideTag(remaining, splitPoint) {
			// Find the start of the tag/placeholder and split before it
			tagStart := findStartOfBoundary(remaining, splitPoint)
			if tagStart > 0 {
				splitPoint = tagStart
			}
		}

		chunks = append(chunks, strings.TrimSpace(remaining[:splitPoint]))
		remaining = strings.TrimSpace(remaining[splitPoint:])
	}

	return chunks
}

// isInsideTag checks if the position is inside an HTML tag <...> or a placeholder [TAG_N]
func isInsideTag(text string, pos int) bool {
	// Check for HTML tags
	lastOpenAngle := strings.LastIndex(text[:pos], "<")
	lastCloseAngle := strings.LastIndex(text[:pos], ">")
	if lastOpenAngle > lastCloseAngle {
		return true
	}

	// Check for placeholders [TAG_N]
	lastOpenBracket := strings.LastIndex(text[:pos], "[")
	lastCloseBracket := strings.LastIndex(text[:pos], "]")
	if lastOpenBracket > lastCloseBracket {
		// Verify if it looks like a TAG placeholder
		if strings.HasPrefix(text[lastOpenBracket:pos], "[TAG_") ||
			(pos < len(text) && strings.Contains(text[lastOpenBracket:], "]") && strings.HasPrefix(text[lastOpenBracket:], "[TAG_")) {
			return true
		}
	}

	return false
}

// findStartOfBoundary finds the start of the tag or placeholder at the current position
func findStartOfBoundary(text string, pos int) int {
	lastOpenAngle := strings.LastIndex(text[:pos], "<")
	lastOpenBracket := strings.LastIndex(text[:pos], "[")

	if lastOpenAngle > lastOpenBracket {
		return lastOpenAngle
	}
	return lastOpenBracket
}
