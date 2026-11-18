// Package slug builds URL-friendly slugs from arbitrary text.
package slug

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// Make converts text into a lowercase, ASCII, hyphen-separated slug. Accented
// letters are folded to their base form (e.g. "Anúncios" -> "anuncios").
func Make(text string) string {
	decomposed := norm.NFD.String(text)

	var b strings.Builder
	lastHyphen := false
	for _, r := range decomposed {
		if unicode.Is(unicode.Mn, r) {
			continue // drop combining marks (accents)
		}
		lr := unicode.ToLower(r)
		switch {
		case (lr >= 'a' && lr <= 'z') || (lr >= '0' && lr <= '9'):
			b.WriteRune(lr)
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
