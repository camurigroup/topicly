package httpx

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeJSONRejectsInvalidUTF8(t *testing.T) {
	var dst struct {
		Name string `json:"name"`
	}
	// "Anúncios" encoded as Latin-1 (0xFA) — what a misconfigured client sends.
	bad := httptest.NewRequest("POST", "/", strings.NewReader("{\"name\":\"An\xfancios\"}"))
	if err := DecodeJSON(bad, &dst, 1024); err != ErrInvalidUTF8 {
		t.Fatalf("expected ErrInvalidUTF8, got %v (name=%q)", err, dst.Name)
	}
	good := httptest.NewRequest("POST", "/", strings.NewReader(`{"name":"Anúncios à vista"}`))
	if err := DecodeJSON(good, &dst, 1024); err != nil || dst.Name != "Anúncios à vista" {
		t.Fatalf("valid UTF-8 must decode: %v %q", err, dst.Name)
	}
}
