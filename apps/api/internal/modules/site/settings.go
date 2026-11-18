package site

import (
	"strings"
)

// Setting kinds drive validation and the admin form.
const (
	KindText     = "text"
	KindLongText = "longtext"
	KindBool     = "bool"
	KindPath     = "path"  // internal path such as /categories
	KindURL      = "url"   // internal path or absolute http(s) URL
	KindColor    = "color" // #rgb or #rrggbb (empty = theme default)
	KindLinks    = "links" // one "Label|url" per line
)

// Definition describes one administrable setting.
type Definition struct {
	Key     string `json:"key"`
	Label   string `json:"label"`
	Group   string `json:"group"`
	Kind    string `json:"kind"`
	Max     int    `json:"maxLength"`
	Default string `json:"default"`
}

// Definitions is the allow-list of administrable settings and their defaults.
// Defaults are seeded into the database on boot, so the texts shown on the
// home page always come from the database, never from the client code.
var Definitions = []Definition{
	{"home.banner.title", "Título do banner", "Banner", KindText, 120, "Bem-vindo à comunidade"},
	{"home.banner.subtitle", "Subtítulo do banner", "Banner", KindLongText, 300, "Um lugar para conversar, aprender e compartilhar."},
	{"home.banner.image", "Imagem do banner (URL)", "Banner", KindURL, 500, ""},
	{"home.banner.ctaLabel", "Texto do botão", "Banner", KindText, 40, "Explorar categorias"},
	{"home.banner.ctaLink", "Link do botão", "Banner", KindPath, 200, "/categories"},
	{"home.welcome.member", "Boas-vindas ao membro ({name} = nome)", "Banner", KindText, 200, "Bem-vindo de volta, {name}!"},

	{"home.intro.title", "Título da apresentação", "Apresentação", KindText, 120, "Sobre a comunidade"},
	{"home.intro.text", "Texto da apresentação", "Apresentação", KindLongText, 2000,
		"Participe das discussões, siga os temas do seu interesse e receba notificações em tempo real."},

	{"home.section.categories.title", "Título: categorias", "Seções", KindText, 80, "Categorias"},
	{"home.section.featured.title", "Título: tópicos em destaque", "Seções", KindText, 80, "Em destaque"},
	{"home.section.recent.title", "Título: tópicos recentes", "Seções", KindText, 80, "Tópicos recentes"},
	{"home.section.popular.title", "Título: populares", "Seções", KindText, 80, "Mais populares"},
	{"home.section.users.title", "Título: usuários em destaque", "Seções", KindText, 80, "Membros em destaque"},
	{"home.section.stats.title", "Título: estatísticas", "Seções", KindText, 80, "Estatísticas"},
	{"home.section.activity.title", "Título: últimas atividades", "Seções", KindText, 80, "Últimas atividades"},

	{"home.show.categories", "Mostrar categorias", "Visibilidade", KindBool, 5, "true"},
	{"home.show.featured", "Mostrar tópicos em destaque", "Visibilidade", KindBool, 5, "true"},
	{"home.show.recent", "Mostrar tópicos recentes", "Visibilidade", KindBool, 5, "true"},
	{"home.show.popular", "Mostrar populares", "Visibilidade", KindBool, 5, "true"},
	{"home.show.users", "Mostrar usuários em destaque", "Visibilidade", KindBool, 5, "true"},
	{"home.show.stats", "Mostrar estatísticas", "Visibilidade", KindBool, 5, "true"},
	{"home.show.activity", "Mostrar últimas atividades", "Visibilidade", KindBool, 5, "true"},

	// White-label identity (Etapa 6), read by the public layout.
	{"brand.name", "Nome do fórum", "Identidade", KindText, 60, "Forum"},
	{"brand.description", "Descrição (SEO)", "Identidade", KindLongText, 300, "Uma comunidade para conversar, aprender e compartilhar."},
	{"brand.logo", "Logo (URL)", "Identidade", KindURL, 500, ""},
	{"brand.favicon", "Favicon (URL)", "Identidade", KindURL, 500, ""},

	{"theme.primary", "Cor primária", "Cores", KindColor, 7, ""},
	{"theme.accent", "Cor de destaque", "Cores", KindColor, 7, ""},
	{"theme.radius", "Arredondamento (px)", "Cores", KindText, 2, "8"},

	{"social.website", "Site", "Redes sociais", KindURL, 300, ""},
	{"social.github", "GitHub", "Redes sociais", KindURL, 300, ""},
	{"social.x", "X / Twitter", "Redes sociais", KindURL, 300, ""},
	{"social.instagram", "Instagram", "Redes sociais", KindURL, 300, ""},
	{"social.discord", "Discord", "Redes sociais", KindURL, 300, ""},
	{"social.youtube", "YouTube", "Redes sociais", KindURL, 300, ""},

	{"footer.text", "Texto do rodapé", "Rodapé", KindText, 200, "Construído com o Forum Framework."},
	{"footer.links", "Links institucionais (Rótulo|URL, um por linha)", "Rodapé", KindLinks, 2000,
		"Termos de uso|/terms\nPrivacidade|/privacy"},

	{"legal.terms", "Termos de uso", "Documentos legais", KindLongText, 15000, "Ao usar esta comunidade você concorda em tratar os demais com respeito."},
	{"legal.privacy", "Política de privacidade", "Documentos legais", KindLongText, 15000, "Coletamos apenas os dados necessários para o funcionamento da conta."},
}

var definitionByKey = func() map[string]Definition {
	m := make(map[string]Definition, len(Definitions))
	for _, d := range Definitions {
		m[d.Key] = d
	}
	return m
}()

// Normalize validates a value for key and returns its canonical form.
func Normalize(key, value string) (string, error) {
	d, ok := definitionByKey[key]
	if !ok {
		return "", validationErr("configuração desconhecida: " + key)
	}
	v := strings.TrimSpace(value)
	if len([]rune(v)) > d.Max {
		return "", validationErr(d.Label + ": máximo de " + itoa(d.Max) + " caracteres")
	}
	switch d.Kind {
	case KindText:
		if strings.ContainsAny(v, "\r\n") {
			return "", validationErr(d.Label + ": não use quebras de linha")
		}
	case KindBool:
		switch strings.ToLower(v) {
		case "true", "1", "on":
			v = "true"
		case "false", "0", "off", "":
			v = "false"
		default:
			return "", validationErr(d.Label + ": use true ou false")
		}
	case KindPath:
		if v != "" && !isInternalPath(v) {
			return "", validationErr(d.Label + ": use um caminho interno, ex.: /categories")
		}
	case KindURL:
		if v != "" && !isSafeURL(v) {
			return "", validationErr(d.Label + ": use um caminho interno ou uma URL http(s)")
		}
	case KindColor:
		c, ok := normalizeColor(v)
		if !ok {
			return "", validationErr(d.Label + ": use uma cor hexadecimal, ex.: #7c3aed")
		}
		v = c
	case KindLinks:
		if err := validateLinks(v); err != nil {
			return "", validationErr(d.Label + ": " + err.Error())
		}
	}
	return v, nil
}

func isSafeURL(v string) bool {
	return isInternalPath(v) || strings.HasPrefix(v, "https://") || strings.HasPrefix(v, "http://")
}

// normalizeColor accepts "", #rgb or #rrggbb and returns lowercase #rrggbb.
func normalizeColor(v string) (string, bool) {
	if v == "" {
		return "", true
	}
	v = strings.ToLower(v)
	if !strings.HasPrefix(v, "#") || (len(v) != 4 && len(v) != 7) {
		return "", false
	}
	for _, r := range v[1:] {
		if !strings.ContainsRune("0123456789abcdef", r) {
			return "", false
		}
	}
	if len(v) == 4 {
		v = "#" + strings.Repeat(v[1:2], 2) + strings.Repeat(v[2:3], 2) + strings.Repeat(v[3:4], 2)
	}
	return v, true
}

type linksError string

func (e linksError) Error() string { return string(e) }

// validateLinks checks "Label|url" lines with safe URLs.
func validateLinks(v string) error {
	for _, line := range strings.Split(v, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		label, url, ok := strings.Cut(line, "|")
		if !ok || strings.TrimSpace(label) == "" {
			return linksError("use o formato Rótulo|URL em cada linha")
		}
		if !isSafeURL(strings.TrimSpace(url)) {
			return linksError("URL inválida em \"" + strings.TrimSpace(label) + "\"")
		}
	}
	return nil
}

func isInternalPath(v string) bool {
	return strings.HasPrefix(v, "/") && !strings.HasPrefix(v, "//") && !strings.Contains(v, `\`)
}

// Bool reads a boolean setting from a value map.
func Bool(values map[string]string, key string) bool {
	return values[key] == "true"
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [12]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
