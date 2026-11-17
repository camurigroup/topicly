package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/forum-framework/api/internal/config"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Open opens a connection pool for the configured driver (postgres or mariadb).
func Open(cfg config.DatabaseConfig) (*sql.DB, error) {
	driverName, dataSource, err := dsn(cfg)
	if err != nil {
		return nil, err
	}
	db, err := sql.Open(driverName, dataSource)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)
	return db, nil
}

// Dialect maps a driver name to its migration subdirectory and canonical name.
func Dialect(driver string) (string, error) {
	switch driver {
	case "postgres", "postgresql", "pgx":
		return "postgres", nil
	case "mariadb", "mysql":
		return "mariadb", nil
	default:
		return "", fmt.Errorf("unsupported DB_DRIVER %q (use postgres or mariadb)", driver)
	}
}

func dsn(cfg config.DatabaseConfig) (string, string, error) {
	switch cfg.Driver {
	case "postgres", "postgresql", "pgx":
		ssl := cfg.SSLMode
		if ssl == "" {
			ssl = "disable"
		}
		ds := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, ssl)
		return "pgx", ds, nil
	case "mariadb", "mysql":
		ds := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&multiStatements=true&charset=utf8mb4&collation=utf8mb4_unicode_ci",
			cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)
		return "mysql", ds, nil
	default:
		return "", "", fmt.Errorf("unsupported DB_DRIVER %q (use postgres or mariadb)", cfg.Driver)
	}
}
