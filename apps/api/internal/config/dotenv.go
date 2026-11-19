package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// LoadDotEnv reads KEY=VALUE lines from path into the process environment.
// Variables already set in the real environment win. A missing file is not
// an error (containers usually pass variables directly).
func LoadDotEnv(path string) error {
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	for k, v := range parseDotEnv(string(b)) {
		if _, set := os.LookupEnv(k); !set {
			if err := os.Setenv(k, v); err != nil {
				return err
			}
		}
	}
	return nil
}

// parseDotEnv supports comments, blank lines, "export KEY=v" and quoted values.
func parseDotEnv(src string) map[string]string {
	out := map[string]string{}
	for _, line := range strings.Split(strings.ReplaceAll(src, "\r\n", "\n"), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		key, val, ok := strings.Cut(line, "=")
		key = strings.TrimSpace(key)
		if !ok || key == "" || strings.ContainsAny(key, " \t") {
			continue
		}
		val = strings.TrimSpace(val)
		if len(val) >= 2 && (val[0] == '"' || val[0] == '\'') && val[len(val)-1] == val[0] {
			val = val[1 : len(val)-1]
		}
		out[key] = val
	}
	return out
}

// MinSecretLength is the minimum JWT secret size accepted in production.
const MinSecretLength = 32

var defaultSecrets = map[string]bool{"dev-access-secret": true, "dev-refresh-secret": true}

// Validate refuses unsafe settings in production (default, short or shared
// JWT secrets). Development keeps the convenient defaults.
func (c Config) Validate() error {
	if c.AppEnv != "production" {
		return nil
	}
	for name, s := range map[string]string{"JWT_ACCESS_SECRET": c.JWTAccessSecret, "JWT_REFRESH_SECRET": c.JWTRefreshSecret} {
		if defaultSecrets[s] || len(s) < MinSecretLength {
			return fmt.Errorf("%s deve ter ao menos %d caracteres e não pode ser o valor padrão em produção", name, MinSecretLength)
		}
	}
	if c.JWTAccessSecret == c.JWTRefreshSecret {
		return errors.New("JWT_ACCESS_SECRET e JWT_REFRESH_SECRET devem ser diferentes")
	}
	return nil
}
