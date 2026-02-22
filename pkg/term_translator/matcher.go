package term_translator

import (
	"sort"
	"strings"
)

// PartialMatchResult represents a match found in a string
type PartialMatchResult struct {
	Term        ReferenceTerm
	StartIndex  int
	EndIndex    int
	MatchedText string
}

// GreedyLongestMatcher finds the longest non-overlapping matches from a dictionary
type GreedyLongestMatcher struct{}

// NewGreedyLongestMatcher creates a new GreedyLongestMatcher
func NewGreedyLongestMatcher() *GreedyLongestMatcher {
	return &GreedyLongestMatcher{}
}

// Match finds the longest non-overlapping occurrences of reference terms in the text
func (m *GreedyLongestMatcher) Match(text string, candidates []ReferenceTerm) []ReferenceTerm {
	// Case-insensitive matching logic
	lowerText := strings.ToLower(text)

	// Create all possible matches
	var allMatches []PartialMatchResult

	for _, cand := range candidates {
		lowerCand := strings.ToLower(cand.Source)
		if lowerCand == "" {
			continue
		}

		// Find all occurrences of the candidate in the text
		startIndex := 0
		for {
			idx := strings.Index(lowerText[startIndex:], lowerCand)
			if idx == -1 {
				break
			}
			absoluteIdx := startIndex + idx
			allMatches = append(allMatches, PartialMatchResult{
				Term:        cand,
				StartIndex:  absoluteIdx,
				EndIndex:    absoluteIdx + len(lowerCand),
				MatchedText: text[absoluteIdx : absoluteIdx+len(lowerCand)],
			})
			startIndex = absoluteIdx + len(lowerCand)
		}
	}

	// Sort by length (descending) first, then by StartIndex (ascending)
	sort.Slice(allMatches, func(i, j int) bool {
		lenI := allMatches[i].EndIndex - allMatches[i].StartIndex
		lenJ := allMatches[j].EndIndex - allMatches[j].StartIndex
		if lenI != lenJ {
			return lenI > lenJ
		}
		return allMatches[i].StartIndex < allMatches[j].StartIndex
	})

	// Apply greedy selection to prevent overlapping
	var selected []ReferenceTerm
	consumedPositions := make(map[int]bool)

	// Map to de-duplicate results
	seenSources := make(map[string]bool)

	for _, match := range allMatches {
		// check if overlapped
		overlapped := false
		for i := match.StartIndex; i < match.EndIndex; i++ {
			if consumedPositions[i] {
				overlapped = true
				break
			}
		}

		if !overlapped {
			// mark as consumed
			for i := match.StartIndex; i < match.EndIndex; i++ {
				consumedPositions[i] = true
			}

			if !seenSources[match.Term.Source] {
				selected = append(selected, match.Term)
				seenSources[match.Term.Source] = true
			}
		}
	}

	return selected
}
