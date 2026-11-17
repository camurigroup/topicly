// Package migrations embeds the SQL migration files for each supported dialect.
package migrations

import "embed"

//go:embed postgres mariadb
var FS embed.FS
