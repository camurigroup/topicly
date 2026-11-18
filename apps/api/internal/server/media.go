package server

import (
	"context"
	"database/sql"
	"strings"

	"github.com/forum-framework/api/internal/database"
	"github.com/forum-framework/api/internal/storage"
)

// mediaSettingKeys are site settings that may hold uploaded image URLs.
var mediaSettingKeys = []string{"brand.logo", "brand.favicon", "home.banner.image"}

// normalizeMediaURLs rewrites stored absolute local-file URLs
// ("http://host:port/files/…") to "/files/…", so images keep working when the
// API host or port changes. Idempotent; only runs for the local provider.
func (s *Server) normalizeMediaURLs(ctx context.Context) (int, error) {
	if p := strings.ToLower(s.cfg.Storage.Provider); p != "" && p != "local" {
		return 0, nil
	}
	fixed := 0
	for _, col := range []string{"avatar_url", "banner_url"} {
		n, err := s.rewriteColumn(ctx,
			"SELECT id, "+col+" FROM users WHERE "+col+" LIKE 'http%/files/%'",
			"UPDATE users SET "+col+" = ? WHERE id = ?")
		if err != nil {
			return fixed, err
		}
		fixed += n
	}
	keys := "'" + strings.Join(mediaSettingKeys, "', '") + "'"
	n, err := s.rewriteColumn(ctx,
		"SELECT setting_key, value FROM site_settings WHERE setting_key IN ("+keys+") AND value LIKE 'http%/files/%'",
		"UPDATE site_settings SET value = ? WHERE setting_key = ?")
	return fixed + n, err
}

// rewriteColumn applies storage.RelativeFileURL to the (id, value) rows of
// selectQ and saves the changed ones with updateQ(value, id).
func (s *Server) rewriteColumn(ctx context.Context, selectQ, updateQ string) (int, error) {
	rows, err := s.db.QueryContext(ctx, selectQ)
	if err != nil {
		return 0, err
	}
	type pair struct{ id, value string }
	var todo []pair
	for rows.Next() {
		var (
			id string
			v  sql.NullString
		)
		if err := rows.Scan(&id, &v); err != nil {
			rows.Close()
			return 0, err
		}
		if rel, changed := storage.RelativeFileURL(v.String); changed {
			todo = append(todo, pair{id, rel})
		}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return 0, err
	}
	upd := database.Rebind(s.dialect, updateQ)
	for _, p := range todo {
		if _, err := s.db.ExecContext(ctx, upd, p.value, p.id); err != nil {
			return 0, err
		}
	}
	return len(todo), nil
}
