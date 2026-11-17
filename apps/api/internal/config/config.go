package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all runtime configuration, loaded from environment variables.
type Config struct {
	AppName string
	AppEnv  string
	AppURL  string
	APIPort string
	APIURL  string

	JWTAccessSecret  string
	JWTRefreshSecret string
	JWTAccessTTL     time.Duration
	JWTRefreshTTL    time.Duration

	DB       DatabaseConfig
	RedisURL string
	Storage  StorageConfig

	CORSAllowedOrigins []string
}

// DatabaseConfig selects and configures the database driver.
type DatabaseConfig struct {
	Driver   string // postgres | mariadb
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

// StorageConfig selects and configures the file storage provider.
type StorageConfig struct {
	Provider       string // local | minio
	LocalPath      string
	PublicURL      string
	MinioEndpoint  string
	MinioAccessKey string
	MinioSecretKey string
	MinioBucket    string
	MinioUseSSL    bool
}

// Load reads configuration from the environment, applying sane defaults.
func Load() Config {
	return Config{
		AppName:          getEnv("APP_NAME", "Forum"),
		AppEnv:           getEnv("APP_ENV", "development"),
		AppURL:           getEnv("APP_URL", "http://localhost:3000"),
		APIPort:          getEnv("API_PORT", "8080"),
		APIURL:           getEnv("API_URL", "http://localhost:8080"),
		JWTAccessSecret:  getEnv("JWT_ACCESS_SECRET", "dev-access-secret"),
		JWTRefreshSecret: getEnv("JWT_REFRESH_SECRET", "dev-refresh-secret"),
		JWTAccessTTL:     getDuration("JWT_ACCESS_TTL", 15*time.Minute),
		JWTRefreshTTL:    getDuration("JWT_REFRESH_TTL", 720*time.Hour),
		DB: DatabaseConfig{
			Driver:   getEnv("DB_DRIVER", "postgres"),
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "forum"),
			Password: getEnv("DB_PASSWORD", "forum"),
			Name:     getEnv("DB_NAME", "forum"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		RedisURL: getEnv("REDIS_URL", ""),
		Storage: StorageConfig{
			Provider:       getEnv("STORAGE_PROVIDER", "local"),
			LocalPath:      getEnv("STORAGE_LOCAL_PATH", "./storage"),
			PublicURL:      getEnv("STORAGE_PUBLIC_URL", "http://localhost:8080/files"),
			MinioEndpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
			MinioAccessKey: getEnv("MINIO_ACCESS_KEY", ""),
			MinioSecretKey: getEnv("MINIO_SECRET_KEY", ""),
			MinioBucket:    getEnv("MINIO_BUCKET", "forum"),
			MinioUseSSL:    getBool("MINIO_USE_SSL", false),
		},
		CORSAllowedOrigins: getCSV("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000"}),
	}
}

// RedisEnabled reports whether a Redis URL was configured.
func (c Config) RedisEnabled() bool { return strings.TrimSpace(c.RedisURL) != "" }

func getEnv(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}

func getDuration(key string, def time.Duration) time.Duration {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

func getBool(key string, def bool) bool {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}

func getCSV(key string, def []string) []string {
	v, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(v) == "" {
		return def
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}
