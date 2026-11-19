package site

import (
	"encoding/json"
	"regexp"
	"strings"
)

// Widget is one block of the home page layout. The list order is the render
// order; consecutive non-"full" widgets share a two-column row (main + side).
type Widget struct {
	ID       string   `json:"id"`
	Type     string   `json:"type"`
	Enabled  bool     `json:"enabled"`
	Column   string   `json:"column"`
	Title    string   `json:"title"`
	Variant  string   `json:"variant"`
	Limit    int      `json:"limit"`
	Source   string   `json:"source,omitempty"`
	Roles    []string `json:"roles,omitempty"`
	Category string   `json:"category,omitempty"`
}

// Layout limits.
const (
	MaxWidgets     = 24
	MaxWidgetLimit = 24
)

// widgetVariants lists the allowed variants per type (first = default).
var widgetVariants = map[string][]string{
	"hero":       {"large", "compact"},
	"intro":      {"text", "card"},
	"stats":      {"grid", "list"},
	"categories": {"grid", "list"},
	"topics":     {"list", "cards", "compact"},
	"members":    {"list", "avatars"},
	"activity":   {"list"},
}

var (
	widgetColumns = map[string]bool{"full": true, "main": true, "side": true}
	topicSources  = map[string]bool{"recent": true, "popular": true, "trending": true, "featured": true, "roles": true, "category": true}
	widgetIDRe    = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,39}$`)
	slugRe        = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,139}$`)
)

// DefaultLayout reproduces a balanced two-column home page.
var DefaultLayout = []Widget{
	{ID: "hero", Type: "hero", Enabled: true, Column: "full", Variant: "large"},
	{ID: "recentes", Type: "topics", Enabled: true, Column: "main", Title: "Tópicos recentes", Variant: "list", Limit: 8, Source: "recent"},
	{ID: "sobre", Type: "intro", Enabled: true, Column: "side", Title: "Sobre a comunidade", Variant: "card"},
	{ID: "numeros", Type: "stats", Enabled: true, Column: "side", Title: "Em números", Variant: "grid"},
	{ID: "destaques", Type: "topics", Enabled: true, Column: "main", Title: "Em destaque", Variant: "cards", Limit: 4, Source: "featured"},
	{ID: "categorias", Type: "categories", Enabled: true, Column: "side", Title: "Categorias", Variant: "list", Limit: 8},
	{ID: "equipe", Type: "topics", Enabled: true, Column: "main", Title: "Da equipe", Variant: "list", Limit: 5, Source: "roles", Roles: []string{"owner", "admin", "moderator"}},
	{ID: "membros", Type: "members", Enabled: true, Column: "side", Title: "Membros em destaque", Variant: "list", Limit: 5},
	{ID: "populares", Type: "topics", Enabled: false, Column: "main", Title: "Mais populares", Variant: "compact", Limit: 6, Source: "popular"},
	{ID: "atividade", Type: "activity", Enabled: true, Column: "side", Title: "Últimas atividades", Variant: "list", Limit: 6},
}

// ParseLayout validates and normalizes a JSON layout.
func ParseLayout(raw string) ([]Widget, error) {
	var ws []Widget
	if err := json.Unmarshal([]byte(raw), &ws); err != nil {
		return nil, validationErr("layout inválido: JSON malformado")
	}
	if len(ws) > MaxWidgets {
		return nil, validationErr("layout com blocos demais (máximo " + itoa(MaxWidgets) + ")")
	}
	seen := map[string]bool{}
	heroes := 0
	for i := range ws {
		w := &ws[i]
		if !widgetIDRe.MatchString(w.ID) || seen[w.ID] {
			return nil, validationErr("bloco com id inválido ou repetido: " + w.ID)
		}
		seen[w.ID] = true
		variants, ok := widgetVariants[w.Type]
		if !ok {
			return nil, validationErr("tipo de bloco desconhecido: " + w.Type)
		}
		if !widgetColumns[w.Column] {
			return nil, validationErr("posição inválida no bloco " + w.ID + " (use full, main ou side)")
		}
		if w.Variant == "" {
			w.Variant = variants[0]
		} else if !contains(variants, w.Variant) {
			return nil, validationErr("estilo inválido no bloco " + w.ID)
		}
		w.Title = strings.TrimSpace(w.Title)
		if len([]rune(w.Title)) > 80 {
			return nil, validationErr("título longo demais no bloco " + w.ID)
		}
		if w.Limit < 1 {
			w.Limit = 6
		}
		if w.Limit > MaxWidgetLimit {
			w.Limit = MaxWidgetLimit
		}
		if w.Type == "hero" {
			heroes++
		}
		if err := normalizeSource(w); err != nil {
			return nil, err
		}
	}
	if heroes > 1 {
		return nil, validationErr("só pode haver um banner principal")
	}
	return ws, nil
}

func normalizeSource(w *Widget) error {
	if w.Type != "topics" {
		w.Source, w.Roles, w.Category = "", nil, ""
		return nil
	}
	if w.Source == "" {
		w.Source = "recent"
	}
	if !topicSources[w.Source] {
		return validationErr("fonte de tópicos inválida no bloco " + w.ID)
	}
	if w.Source == "roles" {
		roles := []string{}
		for _, r := range w.Roles {
			r = strings.TrimSpace(r)
			if slugRe.MatchString(r) && !contains(roles, r) {
				roles = append(roles, r)
			}
		}
		if len(roles) == 0 || len(roles) > 10 {
			return validationErr("escolha de 1 a 10 cargos no bloco " + w.ID)
		}
		w.Roles = roles
	} else {
		w.Roles = nil
	}
	if w.Source == "category" {
		if !slugRe.MatchString(w.Category) {
			return validationErr("escolha uma categoria no bloco " + w.ID)
		}
	} else {
		w.Category = ""
	}
	return nil
}

func contains(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

// defaultLayoutJSON is the seeded value of the home.layout setting.
func defaultLayoutJSON() string {
	raw, _ := json.Marshal(DefaultLayout)
	ws, err := ParseLayout(string(raw))
	if err != nil {
		panic("invalid DefaultLayout: " + err.Error())
	}
	b, _ := json.Marshal(ws)
	return string(b)
}
