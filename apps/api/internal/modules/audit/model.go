// Package audit records and lists administrative actions.
package audit

import "time"

// Actor is the user who performed an action.
type Actor struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"displayName"`
}

// Entry is one audit log record.
type Entry struct {
	ID         string         `json:"id"`
	Actor      *Actor         `json:"actor"`
	Action     string         `json:"action"`
	TargetType *string        `json:"targetType"`
	TargetID   *string        `json:"targetId"`
	Details    map[string]any `json:"details"`
	IP         *string        `json:"ip"`
	CreatedAt  time.Time      `json:"createdAt"`
}

// Filter narrows an audit listing.
type Filter struct {
	Action  string
	ActorID string
	Target  string
	Limit   int
	Offset  int
}
