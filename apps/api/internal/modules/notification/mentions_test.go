package notification

import (
	"reflect"
	"testing"
)

func TestExtractMentions(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want []string
	}{
		{"single", "oi @maria tudo bem?", []string{"maria"}},
		{"start of text", "@joao veja isso", []string{"joao"}},
		{"dedup case-insensitive", "@Ana e @ana e @ANA", []string{"ana"}},
		{"order preserved", "@b_1 depois @a_2", []string{"b_1", "a_2"}},
		{"punctuation after", "valeu @pedro, @lia!", []string{"pedro", "lia"}},
		{"ignores email", "mande para joao@site.com", nil},
		{"ignores too short", "@ab não conta", nil},
		{"ignores double at", "@@duplo", nil},
		{"newline", "linha\n@zeca", []string{"zeca"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ExtractMentions(tc.in)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("ExtractMentions(%q) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestExtractMentionsCap(t *testing.T) {
	content := ""
	for i := 0; i < 30; i++ {
		content += " @user" + string(rune('a'+i%26)) + string(rune('a'+i/26))
	}
	if got := ExtractMentions(content); len(got) != maxMentions {
		t.Fatalf("len = %d, want %d", len(got), maxMentions)
	}
}
