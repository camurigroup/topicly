// Package moderation implements bans, mutes, warnings, reports and their
// enforcement.
package moderation

import (
	"strings"
	"time"

	"github.com/forum-framework/api/internal/shared/apperr"
)

// Sanction types.
const (
	TypeBan  = "ban"
	TypeMute = "mute"
)

// Report target types.
const (
	TargetUser  = "user"
	TargetTopic = "topic"
	TargetReply = "reply"
)

// Report statuses.
const (
	StatusPending   = "pending"
	StatusReviewing = "reviewing"
	StatusResolved  = "resolved"
	StatusRejected  = "rejected"
)

// MaxDurationHours caps temporary sanctions (5 years); 0 means permanent.
const MaxDurationHours = 24 * 365 * 5

// ReportReasons is the closed list of report reasons.
var ReportReasons = []string{"spam", "harassment", "offensive", "inappropriate", "misinformation", "other"}

// UserRef is a compact user reference.
type UserRef struct {
	ID          string  `json:"id"`
	Username    string  `json:"username"`
	DisplayName string  `json:"displayName"`
	AvatarURL   *string `json:"avatarUrl,omitempty"`
}

// Sanction is a ban or a mute.
type Sanction struct {
	ID           string     `json:"id"`
	User         UserRef    `json:"user"`
	Type         string     `json:"type"`
	Reason       string     `json:"reason"`
	IssuedBy     *UserRef   `json:"issuedBy"`
	ExpiresAt    *time.Time `json:"expiresAt"`
	RevokedAt    *time.Time `json:"revokedAt"`
	RevokedBy    *UserRef   `json:"revokedBy"`
	RevokeReason *string    `json:"revokeReason"`
	CreatedAt    time.Time  `json:"createdAt"`
	Active       bool       `json:"active"`
}

// ActiveAt reports whether the sanction is in force at t.
func (s Sanction) ActiveAt(t time.Time) bool {
	if s.RevokedAt != nil {
		return false
	}
	return s.ExpiresAt == nil || s.ExpiresAt.After(t)
}

// Warning is a formal notice to a user.
type Warning struct {
	ID        string    `json:"id"`
	User      UserRef   `json:"user"`
	Reason    string    `json:"reason"`
	IssuedBy  *UserRef  `json:"issuedBy"`
	CreatedAt time.Time `json:"createdAt"`
}

// ReportTarget summarizes the reported entity.
type ReportTarget struct {
	Type    string  `json:"type"`
	ID      string  `json:"id"`
	Title   string  `json:"title"`
	Excerpt string  `json:"excerpt"`
	Link    string  `json:"link"`
	Author  *string `json:"author"`
	Exists  bool    `json:"exists"`
}

// Report is a user complaint about content or another user.
type Report struct {
	ID             string       `json:"id"`
	Reporter       *UserRef     `json:"reporter"`
	Target         ReportTarget `json:"target"`
	Reason         string       `json:"reason"`
	Details        *string      `json:"details"`
	Status         string       `json:"status"`
	HandledBy      *UserRef     `json:"handledBy"`
	ResolutionNote *string      `json:"resolutionNote"`
	CreatedAt      time.Time    `json:"createdAt"`
	UpdatedAt      time.Time    `json:"updatedAt"`
}

// SanctionInput is the payload to ban or mute a user.
type SanctionInput struct {
	Reason        string `json:"reason"`
	DurationHours int    `json:"durationHours"`
}

// RevokeInput is the payload to lift a sanction.
type RevokeInput struct {
	Reason string `json:"reason"`
}

// WarningInput is the payload to warn a user.
type WarningInput struct {
	Reason string `json:"reason"`
}

// CreateReportInput is the payload to file a report.
type CreateReportInput struct {
	TargetType string `json:"targetType"`
	TargetID   string `json:"targetId"`
	Reason     string `json:"reason"`
	Details    string `json:"details"`
}

// UpdateReportInput is the payload to change a report status.
type UpdateReportInput struct {
	Status string `json:"status"`
	Note   string `json:"note"`
}

// ReportFilter narrows the reports queue.
type ReportFilter struct {
	Status     string
	TargetType string
	Limit      int
	Offset     int
}

// SanctionFilter narrows the sanctions listing.
type SanctionFilter struct {
	Type       string
	ActiveOnly bool
	UserID     string
	Limit      int
	Offset     int
}

// UserModeration is the moderation summary of a user.
type UserModeration struct {
	User       UserRef    `json:"user"`
	Roles      []string   `json:"roles"`
	Rank       int        `json:"rank"`
	IsOwner    bool       `json:"isOwner"`
	ActiveBan  *Sanction  `json:"activeBan"`
	ActiveMute *Sanction  `json:"activeMute"`
	Sanctions  []Sanction `json:"sanctions"`
	Warnings   []Warning  `json:"warnings"`
	CanAct     bool       `json:"canAct"`
}

type ruleError struct{ msg string }

func (e *ruleError) Error() string { return e.msg }
func (e *ruleError) Unwrap() error { return apperr.ErrValidation }

func invalid(msg string) error { return &ruleError{msg: msg} }

// Expiry converts a duration in hours to an expiry time (nil = permanent).
func Expiry(hours int, now time.Time) (*time.Time, error) {
	if hours < 0 || hours > MaxDurationHours {
		return nil, invalid("duração inválida (0 = permanente, máximo de 5 anos)")
	}
	if hours == 0 {
		return nil, nil
	}
	t := now.Add(time.Duration(hours) * time.Hour)
	return &t, nil
}

// CleanReason trims and validates a moderation reason.
func CleanReason(reason string) (string, error) {
	reason = strings.TrimSpace(reason)
	if l := len([]rune(reason)); l < 3 || l > 500 {
		return "", invalid("o motivo deve ter entre 3 e 500 caracteres")
	}
	return reason, nil
}

// ValidStatus reports whether s is a known report status.
func ValidStatus(s string) bool {
	switch s {
	case StatusPending, StatusReviewing, StatusResolved, StatusRejected:
		return true
	}
	return false
}

// Validate checks a report payload.
func (in CreateReportInput) Validate() error {
	switch in.TargetType {
	case TargetUser, TargetTopic, TargetReply:
	default:
		return invalid("tipo de alvo inválido (user, topic ou reply)")
	}
	if strings.TrimSpace(in.TargetID) == "" {
		return invalid("alvo da denúncia ausente")
	}
	valid := false
	for _, r := range ReportReasons {
		if r == in.Reason {
			valid = true
			break
		}
	}
	if !valid {
		return invalid("motivo da denúncia inválido")
	}
	if len([]rune(in.Details)) > 1000 {
		return invalid("os detalhes devem ter no máximo 1000 caracteres")
	}
	return nil
}
