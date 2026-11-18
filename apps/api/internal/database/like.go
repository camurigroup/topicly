package database

import "strings"

var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

// ContainsPattern builds a lowercase "%term%" LIKE pattern with LIKE wildcards
// escaped. Backslash is the default LIKE escape character in both PostgreSQL
// and MariaDB, so no ESCAPE clause is needed. Compare against LOWER(column).
func ContainsPattern(term string) string {
	return "%" + likeEscaper.Replace(strings.ToLower(term)) + "%"
}

// SearchTerms splits a free-text query into at most max distinct words.
func SearchTerms(q string, max int) []string {
	seen := make(map[string]struct{})
	var out []string
	for _, w := range strings.Fields(q) {
		w = strings.ToLower(w)
		if _, ok := seen[w]; ok {
			continue
		}
		seen[w] = struct{}{}
		out = append(out, w)
		if len(out) == max {
			break
		}
	}
	return out
}
