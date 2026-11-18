package rbac

import "time"

// Role is a named set of permissions with a hierarchy rank.
type Role struct {
	ID        string    `json:"id"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	RankOrder int       `json:"rankOrder"`
	IsSystem  bool      `json:"isSystem"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Permission is an atomic, toggleable capability.
type Permission struct {
	ID          string    `json:"id"`
	Slug        string    `json:"slug"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	IsActive    bool      `json:"isActive"`
	CreatedAt   time.Time `json:"createdAt"`
}

// Built-in role slugs.
const (
	RoleOwner     = "owner"
	RoleAdmin     = "admin"
	RoleModerator = "moderator"
)

// Permission slugs (verified server-side).
const (
	PermUsersManage       = "users.manage"
	PermUsersBan          = "users.ban"
	PermUsersMute         = "users.mute"
	PermRolesManage       = "roles.manage"
	PermPermissionsManage = "permissions.manage"
	PermCategoriesManage  = "categories.manage"
	PermTopicsEdit        = "topics.edit"
	PermTopicsDelete      = "topics.delete"
	PermTopicsLock        = "topics.lock"
	PermTopicsPin         = "topics.pin"
	PermRepliesEdit       = "replies.edit"
	PermRepliesDelete     = "replies.delete"
	PermReportsManage     = "reports.manage"
	PermWarningsApply     = "warnings.apply"
	PermSettingsManage    = "settings.manage"
	PermStorageManage     = "storage.manage"
	PermLogsView          = "logs.view"
	PermMetricsView       = "metrics.view"
)
