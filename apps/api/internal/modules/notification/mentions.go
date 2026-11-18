package notification

import (
	"regexp"
	"strings"
)

// maxMentions caps how many users a single piece of content can notify.
const maxMentions = 20

// mentionRe matches @handles (3–32 letters, digits or underscores) that are not
// part of a word or e-mail address.
var mentionRe = regexp.MustCompile(`(^|[^a-zA-Z0-9_@.])@([a-zA-Z0-9_]{3,32})\b`)

// ExtractMentions returns the unique, lowercased usernames @mentioned in content,
// in order of first appearance.
func ExtractMentions(content string) []string {
	seen := make(map[string]struct{})
	var out []string
	for _, m := range mentionRe.FindAllStringSubmatch(content, -1) {
		name := strings.ToLower(m[2])
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
		if len(out) == maxMentions {
			break
		}
	}
	return out
}
