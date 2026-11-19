package site

import (
	"context"

	"github.com/forum-framework/api/internal/modules/category"
	"github.com/forum-framework/api/internal/modules/topic"
)

// layout returns the configured home layout (the default when unset/invalid).
func layout(settings map[string]string) []Widget {
	if ws, err := ParseLayout(settings["home.layout"]); err == nil {
		return ws
	}
	return DefaultLayout
}

// Home assembles the home page following the configured layout: only the
// enabled blocks are queried, and each topics block gets its own feed.
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
		Layout:        []Widget{},
		Feeds:         map[string][]topic.Topic{},
	}
	maxMembers, maxActivity := 0, 0
	needStats, needCategories := false, false
	for _, w := range layout(settings) {
		if !w.Enabled {
			continue
		}
		h.Layout = append(h.Layout, w)
		switch w.Type {
		case "stats":
			needStats = true
		case "categories":
			needCategories = true
		case "members":
			maxMembers = max(maxMembers, w.Limit)
		case "activity":
			maxActivity = max(maxActivity, w.Limit)
		case "topics":
			feed, err := s.feed(ctx, w)
			if err != nil {
				return nil, err
			}
			h.Feeds[w.ID] = feed
		}
	}
	if needStats {
		if h.Stats, err = s.repo.Stats(ctx); err != nil {
			return nil, err
		}
	}
	if needCategories {
		if h.Categories, err = s.categories.List(ctx, true); err != nil {
			return nil, err
		}
	}
	if maxMembers > 0 {
		if h.FeaturedUsers, err = s.repo.FeaturedUsers(ctx, maxMembers); err != nil {
			return nil, err
		}
	}
	if maxActivity > 0 {
		if h.Activity, err = s.repo.RecentActivity(ctx, maxActivity); err != nil {
			return nil, err
		}
	}
	return h, nil
}

// feed loads the topics of one block according to its source.
func (s *Service) feed(ctx context.Context, w Widget) ([]topic.Topic, error) {
	var (
		out []topic.Topic
		err error
	)
	switch w.Source {
	case "featured":
		out, err = s.topics.Featured(ctx, w.Limit)
	case "popular":
		out, _, err = s.topics.Discover(ctx, topic.DiscoverPopular, w.Limit, 0)
	case "trending":
		out, _, err = s.topics.Discover(ctx, topic.DiscoverTrending, w.Limit, 0)
	case "roles":
		out, err = s.topics.ByAuthorRoles(ctx, w.Roles, w.Limit)
	case "category":
		id, cerr := s.topics.CategoryIDBySlug(ctx, w.Category)
		if cerr != nil {
			return []topic.Topic{}, nil // category removed: show the block empty
		}
		out, err = s.topics.ListTopics(ctx, id, w.Limit, 0)
	default:
		out, _, err = s.topics.Discover(ctx, topic.DiscoverRecent, w.Limit, 0)
	}
	if out == nil {
		out = []topic.Topic{}
	}
	return out, err
}
