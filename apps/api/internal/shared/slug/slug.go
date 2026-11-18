// Package slug builds URL-friendly slugs from arbitrary text.
package slug

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var accentRemover = transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)

// Make converts text into a lowercase, ASCII, hyphen-separated slug.
func Make(text string) string {
	normalized, _, err := transform.String(accentRemover, text)
	if err != nil {
		normalized = text
	}
	normalized = strings.ToLower(normalized)

	var b strings.Builder
	lastHyphen := false
	for _, r := range normalized {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastHyphen = false
		default:
			if !lastHyphen && b.Len() > 0 {
				b.WriteByte('-')
				lastHyphen = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}
