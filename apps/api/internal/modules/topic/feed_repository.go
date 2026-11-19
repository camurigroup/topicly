package topic

import (
	"context"
	"strings"
)

// ByAuthorRoles returns the newest topics written by users holding any of the
// given role slugs (e.g. a "from the team" home block).
func (r *Repository) ByAuthorRoles(ctx context.Context, roles []string, limit int) ([]Topic, error) {
	if len(roles) == 0 {
		return []Topic{}, nil
	}
	ph := strings.TrimSuffix(strings.Repeat("?, ", len(roles)), ", ")
	q := topicSelect + ` WHERE EXISTS (SELECT 1 FROM user_roles ur JOIN roles ro ON ro.id = ur.role_id
        WHERE ur.user_id = t.author_id AND ro.slug IN (` + ph + `))
        ORDER BY t.created_at DESC LIMIT ?`
	args := make([]any, 0, len(roles)+1)
	for _, s := range roles {
		args = append(args, s)
	}
	args = append(args, limit)
	rows, err := r.store.DB.QueryContext(ctx, r.store.Rebind(q), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Topic{}
	for rows.Next() {
		t, err := scanTopic(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *t)
	}
	return out, rows.Err()
}
