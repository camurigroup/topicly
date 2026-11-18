package site

import (
	"context"

	"github.com/forum-framework/api/internal/modules/category"
	"github.com/forum-framework/api/internal/modules/topic"
)

const (
	homeTopics     = 6
	homeFeatured   = 4
	homeUsers      = 6
	homeActivities = 10
)

// Service manages site settings and assembles the home page.
type Service struct {
	repo       *Repository
	topics     *topic.Repository
	categories *category.Repository
}

// NewService builds a site Service.
func NewService(repo *Repository, topics *topic.Repository, categories *category.Repository) *Service {
	return &Service{repo: repo, topics: topics, categories: categories}
}

// Seed stores the default value of every setting that is still missing.
func (s *Service) Seed(ctx context.Context) error {
	defaults := make(map[string]string, len(Definitions))
	for _, d := range Definitions {
		defaults[d.Key] = d.Default
	}
	return s.repo.InsertMissing(ctx, defaults)
}

// Settings returns the value of every defined setting (defaults fill gaps).
func (s *Service) Settings(ctx context.Context) (map[string]string, error) {
	stored, err := s.repo.Values(ctx)
	if err != nil {
		return nil, err
	}
	out := make(map[string]string, len(Definitions))
	for _, d := range Definitions {
		if v, ok := stored[d.Key]; ok {
			out[d.Key] = v
		} else {
			out[d.Key] = d.Default
		}
	}
	return out, nil
}

// UpdateSettings validates and saves a partial set of settings.
func (s *Service) UpdateSettings(ctx context.Context, in map[string]string) (map[string]string, error) {
	if len(in) == 0 {
		return nil, validationErr("nenhuma configuração enviada")
	}
	clean := make(map[string]string, len(in))
	for k, v := range in {
		nv, err := Normalize(k, v)
		if err != nil {
			return nil, err
		}
		clean[k] = nv
	}
	if err := s.repo.Save(ctx, clean); err != nil {
		return nil, err
	}
	return s.Settings(ctx)
}

// Home is everything the home page renders, in one response.
type Home struct {
	Settings      map[string]string   `json:"settings"`
	Stats         *Stats              `json:"stats,omitempty"`
	Categories    []category.Category `json:"categories"`
	Featured      []topic.Topic       `json:"featuredTopics"`
	Recent        []topic.Topic       `json:"recentTopics"`
	Popular       []topic.Topic       `json:"popularTopics"`
	FeaturedUsers []FeaturedUser      `json:"featuredUsers"`
	Activity      []Activity          `json:"activity"`
}

// Home assembles the home page, querying only the sections that are enabled.
func (s *Service) Home(ctx context.Context) (*Home, error) {
	settings, err := s.Settings(ctx)
	if err != nil {
		return nil, err
	}
	h := &Home{
		Settings:      settings,
		Categories:    []category.Category{},
		Featured:      []topic.Topic{},
		Recent:        []topic.Topic{},
		Popular:       []topic.Topic{},
		FeaturedUsers: []FeaturedUser{},
		Activity:      []Activity{},
	}
	if Bool(settings, "home.show.stats") {
		if h.Stats, err = s.repo.Stats(ctx); err != nil {
			return nil, err
		}
	}
	if Bool(settings, "home.show.categories") {
		if h.Categories, err = s.categories.List(ctx, true); err != nil {
			return nil, err
		}
	}
	if Bool(settings, "home.show.featured") {
		if h.Featured, err = s.topics.Featured(ctx, homeFeatured); err != nil {
			return nil, err
		}
	}
	if Bool(settings, "home.show.recent") {
		if h.Recent, _, err = s.topics.Discover(ctx, topic.DiscoverRecent, homeTopics, 0); err != nil {
			return nil, err
		}
	}
	if Bool(settings, "home.show.popular") {
		if h.Popular, _, err = s.topics.Discover(ctx, topic.DiscoverPopular, homeTopics, 0); err != nil {
			return nil, err
		}
	}
	if Bool(settings, "home.show.users") {
		if h.FeaturedUsers, err = s.repo.FeaturedUsers(ctx, homeUsers); err != nil {
			return nil, err
		}
	}
	if Bool(settings, "home.show.activity") {
		if h.Activity, err = s.repo.RecentActivity(ctx, homeActivities); err != nil {
			return nil, err
		}
	}
	return h, nil
}
