package category

import "time"

// Category organizes topics and may nest via ParentID (subcategories).
type Category struct {
	ID          string    `json:"id"`
	ParentID    *string   `json:"parentId"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description *string   `json:"description"`
	Icon        *string   `json:"icon"`
	CoverURL    *string   `json:"coverUrl"`
	Position    int       `json:"position"`
	IsActive    bool      `json:"isActive"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// CreateInput is the payload to create a category.
type CreateInput struct {
	ParentID    *string `json:"parentId"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Icon        *string `json:"icon"`
	CoverURL    *string `json:"coverUrl"`
	Position    int     `json:"position"`
}

// UpdateInput is the payload to update a category.
type UpdateInput struct {
	// Partial update (PATCH): nil keeps the current value; for the optional
	// text fields an empty string clears them.
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Icon        *string `json:"icon"`
	CoverURL    *string `json:"coverUrl"`
	Position    *int    `json:"position"`
	IsActive    *bool   `json:"isActive"`
}
