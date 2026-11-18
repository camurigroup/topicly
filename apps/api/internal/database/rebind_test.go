package database

import "testing"

func TestRebind(t *testing.T) {
	cases := []struct {
		name    string
		dialect string
		in      string
		want    string
	}{
		{"mariadb keeps question marks", "mariadb", "SELECT * FROM x WHERE a = ? AND b = ?", "SELECT * FROM x WHERE a = ? AND b = ?"},
		{"postgres numbers placeholders", "postgres", "SELECT * FROM x WHERE a = ? AND b = ?", "SELECT * FROM x WHERE a = $1 AND b = $2"},
		{"postgres no placeholders", "postgres", "SELECT 1", "SELECT 1"},
		{"postgres ten placeholders", "postgres", "??????????", "$1$2$3$4$5$6$7$8$9$10"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Rebind(c.dialect, c.in); got != c.want {
				t.Fatalf("Rebind(%q, %q) = %q, want %q", c.dialect, c.in, got, c.want)
			}
		})
	}
}
