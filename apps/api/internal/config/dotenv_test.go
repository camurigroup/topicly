package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseDotEnv(t *testing.T) {
	src := "# comentario\nDB_DRIVER=mariadb\n\nexport API_PORT=9000\nQUOTED=\"com espaco\"\nSINGLE='x=y'\nEMPTY=\nBROKEN LINE\n"
	got := parseDotEnv(src)
	want := map[string]string{"DB_DRIVER": "mariadb", "API_PORT": "9000", "QUOTED": "com espaco", "SINGLE": "x=y", "EMPTY": ""}
	if len(got) != len(want) {
		t.Fatalf("got %v", got)
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("%s = %q, want %q", k, got[k], v)
		}
	}
}

func TestLoadDotEnvDoesNotOverrideRealEnv(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(path, []byte("FORUM_T1=from-file\nFORUM_T2=from-file\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("FORUM_T1", "from-env")
	os.Unsetenv("FORUM_T2")
	if err := LoadDotEnv(path); err != nil {
		t.Fatal(err)
	}
	if os.Getenv("FORUM_T1") != "from-env" || os.Getenv("FORUM_T2") != "from-file" {
		t.Fatalf("T1=%q T2=%q", os.Getenv("FORUM_T1"), os.Getenv("FORUM_T2"))
	}
	os.Unsetenv("FORUM_T2")
	if err := LoadDotEnv(filepath.Join(t.TempDir(), "missing.env")); err != nil {
		t.Fatalf("missing file must be ignored: %v", err)
	}
}

func TestValidateRejectsWeakSecretsInProduction(t *testing.T) {
	strong := strings.Repeat("k", 40)
	ok := Config{AppEnv: "production", JWTAccessSecret: strong, JWTRefreshSecret: strong + "r"}
	if err := ok.Validate(); err != nil {
		t.Fatalf("strong secrets must pass: %v", err)
	}
	for _, c := range []Config{
		{AppEnv: "production", JWTAccessSecret: "dev-access-secret", JWTRefreshSecret: strong},
		{AppEnv: "production", JWTAccessSecret: "short", JWTRefreshSecret: strong},
		{AppEnv: "production", JWTAccessSecret: strong, JWTRefreshSecret: strong},
	} {
		if err := c.Validate(); err == nil {
			t.Errorf("expected error for %+v", c)
		}
	}
	dev := Config{AppEnv: "development", JWTAccessSecret: "dev-access-secret", JWTRefreshSecret: "dev-refresh-secret"}
	if err := dev.Validate(); err != nil {
		t.Fatalf("development keeps defaults: %v", err)
	}
}
