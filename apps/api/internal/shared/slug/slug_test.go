package slug

import "testing"

func TestMake(t *testing.T) {
	cases := map[string]string{
		"Anúncios Gerais":    "anuncios-gerais",
		"Discussão & Avisos": "discussao-avisos",
		"  Hello World!  ":   "hello-world",
		"Café com Código":    "cafe-com-codigo",
		"JÁ É":               "ja-e",
		"":                   "",
		"###":                "",
	}
	for in, want := range cases {
		if got := Make(in); got != want {
			t.Errorf("Make(%q) = %q, want %q", in, got, want)
		}
	}
}
