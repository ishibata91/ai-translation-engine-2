package terminology

import (
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
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
	allMatches := m.findAllCandidateMatches(text, candidates)
	m.sortByLengthDescending(allMatches)
	selectedMatches := m.selectNonOverlappingMatches(allMatches)
	selected := make([]ReferenceTerm, 0, len(selectedMatches))
	seenSources := make(map[string]bool)
	for _, match := range selectedMatches {
		if seenSources[match.Term.Source] {
			continue
		}
		seenSources[match.Term.Source] = true
		selected = append(selected, match.Term)
	}
	return selected
}

// MatchSpans returns longest-first non-overlapping matches with strict boundaries.
func (m *GreedyLongestMatcher) MatchSpans(text string, candidates []ReferenceTerm) []PartialMatchResult {
	allMatches := m.findAllCandidateMatches(text, candidates)
	m.sortByLengthDescending(allMatches)
	selected := m.selectNonOverlappingMatches(allMatches)
	sort.Slice(selected, func(i, j int) bool {
		return selected[i].StartIndex < selected[j].StartIndex
	})
	return selected
}

// findAllCandidateMatches finds all occurrences of each candidate in the text (case-insensitive).
func (m *GreedyLongestMatcher) findAllCandidateMatches(text string, candidates []ReferenceTerm) []PartialMatchResult {
	lowerText := strings.ToLower(text)
	var allMatches []PartialMatchResult

	for _, cand := range candidates {
		lowerCand := strings.ToLower(cand.Source)
		if lowerCand == "" {
			continue
		}

		startIndex := 0
		for {
			idx := strings.Index(lowerText[startIndex:], lowerCand)
			if idx == -1 {
				break
			}
			absoluteIdx := startIndex + idx
			endIdx := absoluteIdx + len(lowerCand)
			if !isStrictKeywordBoundary(text, absoluteIdx, endIdx) {
				startIndex = absoluteIdx + len(lowerCand)
				continue
			}
			allMatches = append(allMatches, PartialMatchResult{
				Term:        cand,
				StartIndex:  absoluteIdx,
				EndIndex:    endIdx,
				MatchedText: text[absoluteIdx:endIdx],
			})
			startIndex = absoluteIdx + len(lowerCand)
		}
	}

	return allMatches
}

// sortByLengthDescending sorts matches by length descending, then by start index ascending.
func (m *GreedyLongestMatcher) sortByLengthDescending(allMatches []PartialMatchResult) {
	sort.Slice(allMatches, func(i, j int) bool {
		lenI := allMatches[i].EndIndex - allMatches[i].StartIndex
		lenJ := allMatches[j].EndIndex - allMatches[j].StartIndex
		if lenI != lenJ {
			return lenI > lenJ
		}
		return allMatches[i].StartIndex < allMatches[j].StartIndex
	})
}

// selectNonOverlapping applies greedy selection to prevent overlapping matches.
func (m *GreedyLongestMatcher) selectNonOverlappingMatches(allMatches []PartialMatchResult) []PartialMatchResult {
	var selected []PartialMatchResult
	consumedPositions := make(map[int]bool)

	for _, match := range allMatches {
		if m.isOverlapping(match, consumedPositions) {
			continue
		}

		m.markConsumed(match, consumedPositions)
		selected = append(selected, match)
	}

	return selected
}

// isOverlapping checks if a match overlaps with already consumed positions.
func (m *GreedyLongestMatcher) isOverlapping(match PartialMatchResult, consumedPositions map[int]bool) bool {
	for i := match.StartIndex; i < match.EndIndex; i++ {
		if consumedPositions[i] {
			return true
		}
	}
	return false
}

// markConsumed marks all positions of a match as consumed.
func (m *GreedyLongestMatcher) markConsumed(match PartialMatchResult, consumedPositions map[int]bool) {
	for i := match.StartIndex; i < match.EndIndex; i++ {
		consumedPositions[i] = true
	}
}

func isStrictKeywordBoundary(text string, start int, end int) bool {
	if start < 0 || end > len(text) || start >= end {
		return false
	}
	if start > 0 {
		before, _ := utf8DecodeLastRuneInString(text[:start])
		if isWordRune(before) {
			return false
		}
	}
	if end < len(text) {
		after, _ := utf8DecodeRuneInString(text[end:])
		if isWordRune(after) {
			return false
		}
	}
	return true
}

func utf8DecodeLastRuneInString(s string) (rune, int) {
	return utf8.DecodeLastRuneInString(s)
}

func utf8DecodeRuneInString(s string) (rune, int) {
	return utf8.DecodeRuneInString(s)
}

func isWordRune(r rune) bool {
	if r == '_' || r == '\'' {
		return true
	}
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}
