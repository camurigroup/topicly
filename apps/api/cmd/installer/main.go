// Command installer prepares a Forum Framework deployment: it asks for (or
// receives as flags) the database and storage settings, checks the database
// connection, applies the migrations and writes a .env with strong secrets.
// The Owner account and identity are then created by the web installer.
//
//	go run ./cmd/installer                 # interactive
//	go run ./cmd/installer -driver mariadb -db-host 127.0.0.1 -db-user forum \
//	    -db-password secret -db-name forum -yes
package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/forum-framework/api/internal/config"
	"github.com/forum-framework/api/internal/database"
	"github.com/forum-framework/api/migrations"
)

type answers struct {
	values map[string]string
	order  []string
}

func (a *answers) set(k, v string) {
	if _, ok := a.values[k]; !ok {
		a.order = append(a.order, k)
	}
	a.values[k] = v
}

func secret(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "erro:", err)
		os.Exit(1)
	}
}

func run() error {
	var (
		out      = flag.String("out", ".env", "arquivo .env a gerar")
		force    = flag.Bool("force", false, "sobrescrever um .env existente")
		yes      = flag.Bool("yes", false, "não perguntar: usar flags e padrões")
		driver   = flag.String("driver", "", "banco: postgres | mariadb")
		host     = flag.String("db-host", "", "host do banco")
		port     = flag.String("db-port", "", "porta do banco")
		user     = flag.String("db-user", "", "usuário do banco")
		pass     = flag.String("db-password", "", "senha do banco")
		name     = flag.String("db-name", "", "nome do banco")
		storage  = flag.String("storage", "", "armazenamento: local | minio")
		apiPort  = flag.String("api-port", "", "porta da API")
		appURL   = flag.String("app-url", "", "URL pública do site")
		skipMigr = flag.Bool("skip-migrations", false, "não aplicar migrations agora")
	)
	flag.Parse()

	if _, err := os.Stat(*out); err == nil && !*force {
		return fmt.Errorf("%s já existe — a instalação já foi feita. Use -force para reinstalar conscientemente", *out)
	}

	in := bufio.NewReader(os.Stdin)
	ask := func(label, given, def string) string {
		if given != "" || *yes {
			if given == "" {
				return def
			}
			return given
		}
		fmt.Printf("%s [%s]: ", label, def)
		line, _ := in.ReadString('\n')
		if v := strings.TrimSpace(line); v != "" {
			return v
		}
		return def
	}

	fmt.Println("== Forum Framework — instalador ==")
	a := &answers{values: map[string]string{}}
	drv := strings.ToLower(ask("Banco (postgres/mariadb)", *driver, "postgres"))
	if drv != "postgres" && drv != "mariadb" && drv != "mysql" {
		return errors.New("banco inválido: use postgres ou mariadb")
	}
	defPort := "5432"
	if drv != "postgres" {
		defPort = "3306"
	}
	a.set("APP_ENV", "production")
	a.set("APP_URL", ask("URL pública do site", *appURL, "http://localhost:3000"))
	a.set("API_PORT", ask("Porta da API", *apiPort, "8080"))
	a.set("DB_DRIVER", drv)
	a.set("DB_HOST", ask("Host do banco", *host, "127.0.0.1"))
	a.set("DB_PORT", ask("Porta do banco", *port, defPort))
	a.set("DB_USER", ask("Usuário do banco", *user, "forum"))
	a.set("DB_PASSWORD", ask("Senha do banco", *pass, ""))
	a.set("DB_NAME", ask("Nome do banco", *name, "forum"))
	if drv == "postgres" {
		a.set("DB_SSLMODE", "disable")
	}
	st := strings.ToLower(ask("Armazenamento (local/minio)", *storage, "local"))
	a.set("STORAGE_PROVIDER", st)
	if st == "minio" {
		a.set("MINIO_ENDPOINT", ask("MinIO endpoint", "", "localhost:9000"))
		a.set("MINIO_ACCESS_KEY", ask("MinIO access key", "", ""))
		a.set("MINIO_SECRET_KEY", ask("MinIO secret key", "", ""))
		a.set("MINIO_BUCKET", ask("MinIO bucket", "", "forum"))
	} else {
		a.set("STORAGE_LOCAL_PATH", ask("Pasta dos arquivos", "", "./storage"))
	}
	a.set("CORS_ALLOWED_ORIGINS", a.values["APP_URL"])
	a.set("JWT_ACCESS_SECRET", secret(32))
	a.set("JWT_REFRESH_SECRET", secret(32))
	a.set("INSTALL_TOKEN", secret(12))

	// Check the connection and prepare the schema with the chosen settings.
	for k, v := range a.values {
		os.Setenv(k, v)
	}
	cfg := config.Load()
	db, err := database.Open(cfg.DB)
	if err != nil {
		return fmt.Errorf("conexão com o banco: %w", err)
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("não foi possível conectar ao banco: %w", err)
	}
	fmt.Println("✓ conexão com o banco OK")
	if !*skipMigr {
		dialect, _ := database.Dialect(cfg.DB.Driver)
		n, err := database.Migrate(db, dialect, migrations.FS, dialect)
		if err != nil {
			return fmt.Errorf("migrations: %w", err)
		}
		fmt.Printf("✓ %d migration(s) aplicada(s)\n", n)
	}

	if err := writeEnv(*out, a); err != nil {
		return err
	}
	fmt.Printf("✓ %s gerado (permissão 600)\n\n", *out)
	fmt.Println("Próximos passos:")
	fmt.Println("  1. Inicie a API (ela lê o .env) e o site.")
	fmt.Printf("  2. Abra %s/install e conclua com o token: %s\n", a.values["APP_URL"], a.values["INSTALL_TOKEN"])
	return nil
}

func writeEnv(path string, a *answers) error {
	var b strings.Builder
	b.WriteString("# Gerado pelo instalador do Forum Framework em " + time.Now().Format(time.RFC3339) + "\n")
	b.WriteString("# Contém segredos: não versione este arquivo.\n")
	for _, k := range a.order {
		v := a.values[k]
		if strings.ContainsAny(v, " #\"'") {
			v = `"` + strings.ReplaceAll(v, `"`, `\"`) + `"`
		}
		b.WriteString(k + "=" + v + "\n")
	}
	return os.WriteFile(path, []byte(b.String()), 0o600)
}
