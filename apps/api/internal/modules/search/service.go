package search

import (
	"context"
	"strings"
	"time"

	"github.com/forum-framework/api/internal/database"
	"github.com/forum-framework/api/internal/modules/category"
	"github.com/forum-framework/api/internal/modules/topic"
	"github.com/forum-framework/api/internal/shared/apperr"
)

// Result types accepted by Params.Type.
const (
	TypeAll        = "all"
	TypeTopics     = "topics"
	TypeReplies    = "replies"
	TypeUsers      = "users"
	TypeCategories = "categories"
)

const (
	maxQueryLen  = 100
	maxTerms     = 6
	previewLimit = 5 // per section when Type is "all"
	dateLayout   = "2006-01-02"
)

// Params are the raw search inputs (already parsed from the query string).
type Params struct {
	Query      string
	Type       string
	Category   string // category slug (includes its subcategories)
	Author     string // username
	From, To   string // YYYY-MM-DD, inclusive
	MinReplies int
	MinLikes   int
	Sort       string // relevance | recent | popular | replies | likes | activity
	Limit      int
	Offset     int
}

// Section is one group of results with its total match count.
type Section[T any] struct {
	Items []T   `json:"items"`
	Total int64 `json:"total"`
}

// Result groups the matches per content type. Only requested sections are set.
type Result struct {
	Query      string                      `json:"query"`
	Type       string                      `json:"type"`
	Topics     *Section[topic.Topic]       `json:"topics,omitempty"`
	Replies    *Section[topic.ReplyHit]    `json:"replies,omitempty"`
	Users      *Section[UserHit]           `json:"users,omitempty"`
	Categories *Section[category.Category] `json:"categories,omitempty"`
}

// Service runs global searches across topics, replies, users and categories.
type Service struct {
	repo   *Repository
	topics *topic.Repository
}

// NewService builds a search Service.
func NewService(repo *Repository, topics *topic.Repository) *Service {
	return &Service{repo: repo, topics: topics}
}

// Search validates params, applies every filter in the database and returns
// the requested sections. With Type "all" each section holds a short preview.
func (s *Service) Search(ctx context.Context, p Params) (*Result, error) {
	p.Query = strings.TrimSpace(p.Query)
	if len([]rune(p.Query)) > maxQueryLen {
		return nil, validationErr("a busca deve ter no máximo 100 caracteres")
	}
	if p.Type == "" {
		p.Type = TypeAll
	}
	switch p.Type {
	case TypeAll, TypeTopics, TypeReplies, TypeUsers, TypeCategories:
	default:
		return nil, validationErr("tipo de busca inválido")
	}
	terms := database.SearchTerms(p.Query, maxTerms)
	hasFilters := p.Category != "" || p.Author != ""
	if len(terms) == 0 && !hasFilters {
		return nil, validationErr("informe um termo de busca ou um filtro de autor/categoria")
	}

	filter, empty, err := s.contentFilter(ctx, p, terms)
	if err != nil {
		return nil, err
	}

	res := &Result{Query: p.Query, Type: p.Type}
	limit, offset := p.Limit, p.Offset
	if p.Type == TypeAll {
		limit, offset = previewLimit, 0
	}
	filter.Limit, filter.Offset = limit, offset
	want := func(t string) bool { return p.Type == TypeAll || p.Type == t }

	if want(TypeTopics) {
		res.Topics = &Section[topic.Topic]{Items: []topic.Topic{}}
		if !empty {
			if res.Topics.Items, res.Topics.Total, err = s.topics.SearchTopics(ctx, filter); err != nil {
				return nil, err
			}
		}
	}
	if want(TypeReplies) {
		res.Replies = &Section[topic.ReplyHit]{Items: []topic.ReplyHit{}}
		if !empty {
			if res.Replies.Items, res.Replies.Total, err = s.topics.SearchReplies(ctx, filter); err != nil {
				return nil, err
			}
		}
	}
	// Users and categories have no author/category dimension: they need terms.
	if want(TypeUsers) {
		res.Users = &Section[UserHit]{Items: []UserHit{}}
		if len(terms) > 0 {
			if res.Users.Items, res.Users.Total, err = s.repo.SearchUsers(ctx, terms, p.Sort, limit, offset); err != nil {
				return nil, err
			}
		}
	}
	if want(TypeCategories) {
		res.Categories = &Section[category.Category]{Items: []category.Category{}}
		if len(terms) > 0 {
			if res.Categories.Items, res.Categories.Total, err = s.repo.SearchCategories(ctx, terms, limit, offset); err != nil {
				return nil, err
			}
		}
	}
	return res, nil
}

// contentFilter resolves the topic/reply filter. empty reports that a filter
// value (unknown author or category) can match nothing.
func (s *Service) contentFilter(ctx context.Context, p Params, terms []string) (topic.SearchFilter, bool, error) {
	f := topic.SearchFilter{Terms: terms, MinReplies: max(p.MinReplies, 0), MinLikes: max(p.MinLikes, 0), Sort: p.Sort}
	if f.Sort == "" {
		f.Sort = topic.SortRelevance
	}
	switch f.Sort {
	case topic.SortRelevance, topic.SortRecent, topic.SortPopular, topic.SortReplies, topic.SortLikes, topic.SortActivity:
	default:
		return f, false, validationErr("ordenação inválida")
	}
	if p.From != "" {
		t, err := time.Parse(dateLayout, p.From)
		if err != nil {
			return f, false, validationErr("data inicial inválida (use AAAA-MM-DD)")
		}
		f.From = &t
	}
	if p.To != "" {
		t, err := time.Parse(dateLayout, p.To)
		if err != nil {
			return f, false, validationErr("data final inválida (use AAAA-MM-DD)")
		}
		end := t.Add(24 * time.Hour) // inclusive end date
		f.To = &end
	}
	if f.From != nil && f.To != nil && !f.From.Before(*f.To) {
		return f, false, validationErr("a data inicial deve ser anterior à final")
	}
	if p.Author != "" {
		uid, err := s.repo.UserIDByUsername(ctx, strings.TrimPrefix(p.Author, "@"))
		if err != nil {
			return f, false, err
		}
		if uid == "" {
			return f, true, nil
		}
		f.AuthorID = uid
	}
	if p.Category != "" {
		cid, err := s.repo.CategoryIDBySlug(ctx, p.Category)
		if err != nil {
			return f, false, err
		}
		if cid == "" {
			return f, true, nil
		}
		f.CategoryID = cid
	}
	return f, false, nil
}

func validationErr(msg string) error { return &validationError{msg: msg} }

type validationError struct{ msg string }

func (e *validationError) Error() string { return e.msg }
func (e *validationError) Unwrap() error { return apperr.ErrValidation }
