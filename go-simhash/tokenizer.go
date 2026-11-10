package simhash

import (
	"strings"
	"unicode"
)

// Feature represents a hashed token with its relative weight.
type Feature struct {
	Token  string
	Weight int
}

// TokenizeWords splits text into normalized lowercased word tokens.
func TokenizeWords(text string) []string {
	var words []string
	var cur strings.Builder

	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			cur.WriteRune(unicode.ToLower(r))
		} else if cur.Len() > 0 {
			words = append(words, cur.String())
			cur.Reset()
		}
	}
	if cur.Len() > 0 {
		words = append(words, cur.String())
	}
	return words
}

// ShingleWords generates word-level n-grams of size n.
func ShingleWords(words []string, n int) []string {
	if n <= 1 || len(words) < n {
		return words
	}
	shingles := make([]string, 0, len(words)-n+1)
	for i := 0; i <= len(words)-n; i++ {
		shingles = append(shingles, strings.Join(words[i:i+n], " "))
	}
	return shingles
}

// FeaturesFromText extracts frequency-weighted features from input text using unigrams and optional n-grams.
func FeaturesFromText(text string, ngramSize int) []Feature {
	words := TokenizeWords(text)
	if len(words) == 0 {
		return nil
	}

	counts := make(map[string]int)
	// Always include unigrams
	for _, w := range words {
		counts[w]++
	}

	// Include n-grams if requested
	if ngramSize > 1 {
		shingles := ShingleWords(words, ngramSize)
		for _, s := range shingles {
			counts[s]++
		}
	}

	features := make([]Feature, 0, len(counts))
	for tok, weight := range counts {
		features = append(features, Feature{
			Token:  tok,
			Weight: weight,
		})
	}
	return features
}
