package terminology

import (
	"log/slog"
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
	slog.Debug("ENTER GreedyLongestMatcher.Match", slog.String("text", text), slog.Int("candidateCount", len(candidates)))

	allMatches := m.findAllCandidateMatches(text, candidates)
	m.sortByLengthDescending(allMatches)
	return m.selectNonOverlapping(allMatches)
}

// findAllCandidateMatches finds all occurrences of each candidate in the text (case-insensitive).
func (m *GreedyLongestMatcher) findAllCandidateMatches(text string, candidates []ReferenceTerm) []PartialMatchResult {
	slog.Debug("ENTER GreedyLongestMatcher.findAllCandidateMatches")

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
			allMatches = append(allMatches, PartialMatchResult{
				Term:        cand,
				StartIndex:  absoluteIdx,
				EndIndex:    absoluteIdx + len(lowerCand),
				MatchedText: text[absoluteIdx : absoluteIdx+len(lowerCand)],
			})
			startIndex = absoluteIdx + len(lowerCand)
		}
	}

	return allMatches
}

// sortByLengthDescending sorts matches by length descending, then by start index ascending.
func (m *GreedyLongestMatcher) sortByLengthDescending(allMatches []PartialMatchResult) {
	slog.Debug("ENTER GreedyLongestMatcher.sortByLengthDescending")

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
func (m *GreedyLongestMatcher) selectNonOverlapping(allMatches []PartialMatchResult) []ReferenceTerm {
	slog.Debug("ENTER GreedyLongestMatcher.selectNonOverlapping")

	var selected []ReferenceTerm
	consumedPositions := make(map[int]bool)
	seenSources := make(map[string]bool)

	for _, match := range allMatches {
		if m.isOverlapping(match, consumedPositions) {
			continue
		}

		m.markConsumed(match, consumedPositions)

		if !seenSources[match.Term.Source] {
			selected = append(selected, match.Term)
			seenSources[match.Term.Source] = true
		}
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
