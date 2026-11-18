package storage

import "testing"

func TestRelativeFileURL(t *testing.T) {
	cases := []struct {
		in, want string
		changed  bool
	}{
		{"http://localhost:8080/files/avatars/a.png", "/files/avatars/a.png", true},
		{"https://forum.example.com/files/brand/x.webp", "/files/brand/x.webp", true},
		{"/files/avatars/a.png", "/files/avatars/a.png", false},
		{"https://cdn.example.com/bucket/avatars/a.png", "https://cdn.example.com/bucket/avatars/a.png", false},
		{"", "", false},
		{"http://evil.example/files/../../etc/passwd", "http://evil.example/files/../../etc/passwd", false},
	}
	for _, c := range cases {
		got, changed := RelativeFileURL(c.in)
		if got != c.want || changed != c.changed {
			t.Errorf("RelativeFileURL(%q) = %q, %v; want %q, %v", c.in, got, changed, c.want, c.changed)
		}
	}
}

func TestLocalPublicURLDefaultsToRelative(t *testing.T) {
	p := &localProvider{publicURL: ""}
	if got := p.PublicURL("avatars/a.png"); got != "/files/avatars/a.png" {
		t.Fatalf("PublicURL = %q", got)
	}
}
