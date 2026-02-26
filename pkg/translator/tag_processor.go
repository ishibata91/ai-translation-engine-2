package translator

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var (
	// regex to match HTML-like tags: <...>
	tagRegex = regexp.MustCompile(`<[^>]+>`)
	// regex to match placeholders: [TAG_N]
	placeholderRegex = regexp.MustCompile(`\[TAG_(\d+)\]`)
)

type tagProcessor struct{}

// NewTagProcessor creates a new TagProcessor instance.
func NewTagProcessor() TagProcessor {
	return &tagProcessor{}
}

// Preprocess identifies tags and replaces them with placeholders [TAG_N].
func (p *tagProcessor) Preprocess(text string) (string, map[string]string) {
	tagMap := make(map[string]string)
	matches := tagRegex.FindAllString(text, -1)
	if len(matches) == 0 {
		return text, tagMap
	}

	// Use a unique placeholder for every tag encounter to maintain position tracking.
	count := 0
	result := tagRegex.ReplaceAllStringFunc(text, func(match string) string {
		placeholder := fmt.Sprintf("[TAG_%d]", count)
		tagMap[placeholder] = match
		count++
		return placeholder
	})

	return result, tagMap
}

// Postprocess restores tags from placeholders.
func (p *tagProcessor) Postprocess(text string, tagMap map[string]string) string {
	result := text
	// Replace placeholders in reverse order of index to avoid [TAG_10] matching [TAG_1]
	// First gather keys and sort them by index descending
	keys := make([]string, 0, len(tagMap))
	for k := range tagMap {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		// keys are [TAG_N], sort by N descending
		var ni, nj int
		fmt.Sscanf(keys[i], "[TAG_%d]", &ni)
		fmt.Sscanf(keys[j], "[TAG_%d]", &nj)
		return ni > nj
	})

	for _, k := range keys {
		result = strings.ReplaceAll(result, k, tagMap[k])
	}
	return result
}

// Validate checks if all placeholders in the original are present in the translated text.
func (p *tagProcessor) Validate(translatedText string, tagMap map[string]string) error {
	missing := []string{}
	for k := range tagMap {
		if !strings.Contains(translatedText, k) {
			missing = append(missing, k)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing placeholders in translation: %s", strings.Join(missing, ", "))
	}

	// Check for hallucinations (placeholders that were not in tagMap)
	matches := placeholderRegex.FindAllStringSubmatch(translatedText, -1)
	for _, m := range matches {
		placeholder := m[0]
		if _, ok := tagMap[placeholder]; !ok {
			return fmt.Errorf("hallucinated placeholder found: %s", placeholder)
		}
	}

	return nil
}
