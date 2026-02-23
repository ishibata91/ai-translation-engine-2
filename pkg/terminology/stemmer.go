package terminology

import (
	"strings"

	"github.com/kljensen/snowball"
)

// KeywordStemmer provides stemming functionality for keywords.
type KeywordStemmer interface {
	Stem(word string) (string, error)
}

// SnowballStemmer is an implementation of KeywordStemmer using the snowball algorithm.
type SnowballStemmer struct {
	language string
}

// NewSnowballStemmer creates a new SnowballStemmer for the specified language.
func NewSnowballStemmer(language string) *SnowballStemmer {
	if language == "" {
		language = "english"
	}
	return &SnowballStemmer{
		language: language,
	}
}

// Stem returns the stemmed version of the input word.
func (s *SnowballStemmer) Stem(word string) (string, error) {
	// Snowball stemmer handles lowercase words best
	stemmed, err := snowball.Stem(strings.ToLower(word), s.language, true)
	if err != nil {
		return "", err
	}
	return stemmed, nil
}
