package admin

import "strings"

// OrderBy builds an ORDER BY expression from an allow-list of sortable
// columns (key → SQL column). Unknown keys are rejected, so user input never
// reaches the SQL text. Order defaults to DESC.
func OrderBy(columns map[string]string, defaultKey, sort, order string) (string, error) {
	if sort == "" {
		sort = defaultKey
	}
	col, ok := columns[sort]
	if !ok {
		return "", seriesError("ordenação inválida: " + sort)
	}
	dir := "DESC"
	if strings.EqualFold(order, "asc") {
		dir = "ASC"
	}
	return col + " " + dir, nil
}

// OneOf validates an optional enum filter.
func OneOf(field, value string, allowed ...string) error {
	if value == "" {
		return nil
	}
	for _, a := range allowed {
		if a == value {
			return nil
		}
	}
	return seriesError(field + " inválido: " + value)
}
