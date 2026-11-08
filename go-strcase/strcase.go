package gostrcase

import (
	"strings"
	"unicode"
)

// SplitWords breaks an arbitrary string into clean constituent words.
// Handles snake_case, kebab-case, camelCase, PascalCase, and uppercase acronyms (e.g. "HTTPServer" -> ["HTTP", "Server"]).
func SplitWords(s string) []string {
	var words []string
	var current strings.Builder

	runes := []rune(s)
	n := len(runes)

	for i := 0; i < n; i++ {
		r := runes[i]

		// Check delimiter
		if isDelimiter(r) {
			if current.Len() > 0 {
				words = append(words, current.String())
				current.Reset()
			}
			continue
		}

		if unicode.IsUpper(r) {
			// Case 1: Lowercase followed by uppercase (e.g. "fooBar" -> at 'B')
			if i > 0 && unicode.IsLower(runes[i-1]) && current.Len() > 0 {
				words = append(words, current.String())
				current.Reset()
			} else if i > 0 && unicode.IsUpper(runes[i-1]) && i+1 < n && unicode.IsLower(runes[i+1]) && current.Len() > 0 {
				// Case 2: Acronym ending into titlecase (e.g. "HTTPServer" -> at 'S', runes[i-1] is 'P', runes[i+1] is 'e')
				words = append(words, current.String())
				current.Reset()
			}
		}

		current.WriteRune(r)
	}

	if current.Len() > 0 {
		words = append(words, current.String())
	}

	return words
}

func isDelimiter(r rune) bool {
	return r == '_' || r == '-' || r == ' ' || r == '.' || r == '/' || r == '\\' || r == ':'
}

// ToSnake converts input to lowercase snake_case (e.g. "user_id").
func ToSnake(s string) string {
	words := SplitWords(s)
	for i, w := range words {
		words[i] = strings.ToLower(w)
	}
	return strings.Join(words, "_")
}

// ToKebab converts input to lowercase kebab-case (e.g. "user-id").
func ToKebab(s string) string {
	words := SplitWords(s)
	for i, w := range words {
		words[i] = strings.ToLower(w)
	}
	return strings.Join(words, "-")
}

// ToScreamingSnake converts input to uppercase SCREAMING_SNAKE_CASE (e.g. "USER_ID").
func ToScreamingSnake(s string) string {
	words := SplitWords(s)
	for i, w := range words {
		words[i] = strings.ToUpper(w)
	}
	return strings.Join(words, "_")
}

// ToPascal converts input to PascalCase (e.g. "UserId", "HttpServer").
func ToPascal(s string) string {
	words := SplitWords(s)
	var b strings.Builder
	for _, w := range words {
		if len(w) == 0 {
			continue
		}
		runes := []rune(w)
		b.WriteRune(unicode.ToUpper(runes[0]))
		for _, r := range runes[1:] {
			b.WriteRune(unicode.ToLower(r))
		}
	}
	return b.String()
}

// ToCamel converts input to camelCase (e.g. "userId", "httpServer").
func ToCamel(s string) string {
	words := SplitWords(s)
	if len(words) == 0 {
		return ""
	}

	var b strings.Builder
	// First word is lowercase
	b.WriteString(strings.ToLower(words[0]))

	// Subsequent words are TitleCased
	for _, w := range words[1:] {
		if len(w) == 0 {
			continue
		}
		runes := []rune(w)
		b.WriteRune(unicode.ToUpper(runes[0]))
		for _, r := range runes[1:] {
			b.WriteRune(unicode.ToLower(r))
		}
	}

	return b.String()
}

// ToCustom converts input with a custom rune delimiter and optional uppercasing.
func ToCustom(s string, delimiter rune, uppercase bool) string {
	words := SplitWords(s)
	for i, w := range words {
		if uppercase {
			words[i] = strings.ToUpper(w)
		} else {
			words[i] = strings.ToLower(w)
		}
	}
	return strings.Join(words, string(delimiter))
}
