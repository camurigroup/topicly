package moderation

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/forum-framework/api/internal/modules/audit"
	"github.com/forum-framework/api/internal/modules/notification"
	"github.com/forum-framework/api/internal/modules/rbac"
	"github.com/forum-framework/api/internal/shared/apperr"
	"github.com/forum-framework/api/internal/shared/id"
)

// Audit action names.
const (
	ActionBan          = "user.ban"
	ActionMute         = "user.mute"
	ActionRevoke       = "sanction.revoke"
	ActionWarn         = "user.warn"
	ActionReportStatus = "report.status"
)

// restrictionTTL bounds how long an enforcement lookup is cached in memory.
// Changes made through this service invalidate the cache immediately.
const restrictionTTL = 15 * time.Second

// TokenRevoker revokes every session (refresh token) of a user.
type TokenRevoker func(ctx context.Context, userID string) error

// Service implements moderation business rules.
type Service struct {
	repo    *Repository
	rbac    *rbac.Service
	audit   *audit.Service
	notify  *notification.Service
	revoker TokenRevoker

	mu    sync.Mutex
	cache map[string]cachedRestriction
}

type cachedRestriction struct {
	sanction *Sanction
	at       time.Time
}

// NewService builds the moderation Service. notifySvc and revoker may be nil.
func NewService(repo *Repository, rbacSvc *rbac.Service, auditSvc *audit.Service, notifySvc *notification.Service, revoker TokenRevoker) *Service {
	return &Service{
		repo: repo, rbac: rbacSvc, audit: auditSvc, notify: notifySvc, revoker: revoker,
		cache: make(map[string]cachedRestriction),
	}
}

// ---- Enforcement ----

// Restriction returns the active sanction of a type for a user, or nil.
func (s *Service) Restriction(ctx context.Context, userID, typ string) (*Sanction, error) {
	key := typ + ":" + userID
	now := time.Now()
	s.mu.Lock()
	if c, ok := s.cache[key]; ok && now.Sub(c.at) < restrictionTTL {
		s.mu.Unlock()
		if c.sanction != nil && !c.sanction.ActiveAt(now.UTC()) {
			return nil, nil
		}
		return c.sanction, nil
	}
	s.mu.Unlock()

	sanction, err := s.repo.ActiveSanction(ctx, userID, typ)
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	s.cache[key] = cachedRestriction{sanction: sanction, at: now}
	s.mu.Unlock()
	return sanction, nil
}

func (s *Service) invalidate(userID string) {
	s.mu.Lock()
	delete(s.cache, TypeBan+":"+userID)
	delete(s.cache, TypeMute+":"+userID)
	s.mu.Unlock()
}

// Describe renders a human message for an active sanction.
func Describe(sanction *Sanction) string {
	what := "Sua conta está banida"
	if sanction.Type == TypeMute {
		what = "Você está silenciado e não pode publicar ou interagir"
	}
	until := " permanentemente"
	if sanction.ExpiresAt != nil {
		until = " até " + sanction.ExpiresAt.UTC().Format("02/01/2006 15:04") + " (UTC)"
	}
	return what + until + ". Motivo: " + sanction.Reason
}

// ---- Sanctions ----

// target resolves a user by username and checks the actor's authority over it.
func (s *Service) target(ctx context.Context, actorID, username string) (rbac.Actor, *UserRef, error) {
	u, err := s.repo.UserByUsername(ctx, username)
	if err != nil {
		return rbac.Actor{}, nil, err
	}
	actor, err := s.rbac.ResolveActor(ctx, actorID)
	if err != nil {
		return rbac.Actor{}, nil, err
	}
	target, err := s.rbac.ResolveActor(ctx, u.ID)
	if err != nil {
		return rbac.Actor{}, nil, err
	}
	if err := actor.CanActOnUser(u.ID, target.Rank, target.IsOwner); err != nil {
		return rbac.Actor{}, nil, err
	}
	return actor, u, nil
}

// Ban bans a user (durationHours 0 = permanent) and ends their sessions.
func (s *Service) Ban(ctx context.Context, actorID, username string, in SanctionInput) (*Sanction, error) {
	return s.sanction(ctx, actorID, username, TypeBan, in)
}

// Mute silences a user (no topics, replies, likes or follows).
func (s *Service) Mute(ctx context.Context, actorID, username string, in SanctionInput) (*Sanction, error) {
	return s.sanction(ctx, actorID, username, TypeMute, in)
}

func (s *Service) sanction(ctx context.Context, actorID, username, typ string, in SanctionInput) (*Sanction, error) {
	reason, err := CleanReason(in.Reason)
	if err != nil {
		return nil, err
	}
	expires, err := Expiry(in.DurationHours, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	_, u, err := s.target(ctx, actorID, username)
	if err != nil {
		return nil, err
	}
	// A new sanction replaces the active one of the same type.
	if _, err := s.repo.RevokeActive(ctx, u.ID, typ, actorID, "substituída por nova sanção"); err != nil {
		return nil, err
	}
	sid := id.New()
	if err := s.repo.CreateSanction(ctx, sid, u.ID, typ, reason, actorID, expires); err != nil {
		return nil, err
	}
	s.invalidate(u.ID)
	if typ == TypeBan && s.revoker != nil {
		_ = s.revoker(ctx, u.ID)
	}

	action, verb := ActionMute, "silenciou"
	if typ == TypeBan {
		action, verb = ActionBan, "baniu"
	}
	details := map[string]any{"username": u.Username, "reason": reason, "durationHours": in.DurationHours}
	if expires != nil {
		details["expiresAt"] = expires.Format(time.RFC3339)
	}
	s.audit.Record(ctx, actorID, action, "user", u.ID, details)

	created, err := s.repo.GetSanction(ctx, sid)
	if err != nil {
		return nil, err
	}
	s.notifyUser(ctx, u.ID, notification.TypePunishment, "A moderação "+verb+" sua conta. "+Describe(created))
	return created, nil
}

// Revoke lifts a sanction. Requires the matching permission and authority
// over the sanctioned user.
func (s *Service) Revoke(ctx context.Context, actorID, sanctionID string, in RevokeInput) (*Sanction, error) {
	reason, err := CleanReason(in.Reason)
	if err != nil {
		return nil, err
	}
	sanction, err := s.repo.GetSanction(ctx, sanctionID)
	if err != nil {
		return nil, err
	}
	actor, _, err := s.target(ctx, actorID, sanction.User.Username)
	if err != nil {
		return nil, err
	}
	need := rbac.PermUsersMute
	if sanction.Type == TypeBan {
		need = rbac.PermUsersBan
	}
	if !actor.Perms[need] {
		return nil, apperr.ErrForbidden
	}
	if err := s.repo.RevokeSanction(ctx, sanctionID, actorID, reason); err != nil {
		return nil, err
	}
	s.invalidate(sanction.User.ID)
	s.audit.Record(ctx, actorID, ActionRevoke, "sanction", sanctionID, map[string]any{
		"username": sanction.User.Username, "type": sanction.Type, "reason": reason,
	})
	what := "silenciamento"
	if sanction.Type == TypeBan {
		what = "banimento"
	}
	s.notifyUser(ctx, sanction.User.ID, notification.TypePunishment, "A moderação revogou seu "+what+".")
	return s.repo.GetSanction(ctx, sanctionID)
}

// Warn records a formal warning to a user.
func (s *Service) Warn(ctx context.Context, actorID, username string, in WarningInput) (*Warning, error) {
	reason, err := CleanReason(in.Reason)
	if err != nil {
		return nil, err
	}
	_, u, err := s.target(ctx, actorID, username)
	if err != nil {
		return nil, err
	}
	wid := id.New()
	if err := s.repo.CreateWarning(ctx, wid, u.ID, reason, actorID); err != nil {
		return nil, err
	}
	s.audit.Record(ctx, actorID, ActionWarn, "user", u.ID, map[string]any{"username": u.Username, "reason": reason})
	s.notifyUser(ctx, u.ID, notification.TypePunishment, "Você recebeu uma advertência da moderação. Motivo: "+reason)
	warnings, err := s.repo.WarningsOf(ctx, u.ID, 1)
	if err != nil || len(warnings) == 0 {
		return nil, err
	}
	return &warnings[0], nil
}

// notifyUser sends an anonymous (moderation) notification.
func (s *Service) notifyUser(ctx context.Context, userID, typ, message string) {
	if s.notify == nil {
		return
	}
	s.notify.Deliver(ctx, notification.Delivery{
		Recipients: []string{userID},
		Event:      notification.Event{Type: typ, Message: message},
	})
}

// UserSummary returns the moderation summary of a user for the actor.
func (s *Service) UserSummary(ctx context.Context, actorID, username string) (*UserModeration, error) {
	u, err := s.repo.UserByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	actor, err := s.rbac.ResolveActor(ctx, actorID)
	if err != nil {
		return nil, err
	}
	target, err := s.rbac.ResolveActor(ctx, u.ID)
	if err != nil {
		return nil, err
	}
	roles, err := s.rbac.RolesForUser(ctx, u.ID)
	if err != nil {
		return nil, err
	}
	if roles == nil {
		roles = []string{}
	}
	sanctions, _, err := s.repo.ListSanctions(ctx, SanctionFilter{UserID: u.ID, Limit: 50})
	if err != nil {
		return nil, err
	}
	warnings, err := s.repo.WarningsOf(ctx, u.ID, 50)
	if err != nil {
		return nil, err
	}
	out := &UserModeration{
		User: *u, Roles: roles, Rank: target.Rank, IsOwner: target.IsOwner,
		Sanctions: sanctions, Warnings: warnings,
		CanAct: actor.CanActOnUser(u.ID, target.Rank, target.IsOwner) == nil,
	}
	for i := range sanctions {
		sc := &sanctions[i]
		if !sc.Active {
			continue
		}
		if sc.Type == TypeBan && out.ActiveBan == nil {
			out.ActiveBan = sc
		}
		if sc.Type == TypeMute && out.ActiveMute == nil {
			out.ActiveMute = sc
		}
	}
	return out, nil
}

// ListSanctions returns a filtered page of sanctions.
func (s *Service) ListSanctions(ctx context.Context, f SanctionFilter) ([]Sanction, int64, error) {
	if f.Type != "" && f.Type != TypeBan && f.Type != TypeMute {
		return nil, 0, invalid("tipo inválido (ban ou mute)")
	}
	return s.repo.ListSanctions(ctx, f)
}

// ---- Reports ----

// CreateReport files a report. A reporter can only have one open report per target.
func (s *Service) CreateReport(ctx context.Context, reporterID string, in CreateReportInput) (*Report, error) {
	in.TargetID = strings.TrimSpace(in.TargetID)
	if err := in.Validate(); err != nil {
		return nil, err
	}
	if in.TargetType == TargetUser && in.TargetID == reporterID {
		return nil, invalid("você não pode denunciar a si mesmo")
	}
	ok, err := s.repo.TargetExists(ctx, in.TargetType, in.TargetID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, apperr.ErrNotFound
	}
	open, err := s.repo.HasOpenReport(ctx, reporterID, in.TargetType, in.TargetID)
	if err != nil {
		return nil, err
	}
	if open {
		return nil, &conflictError{msg: "você já denunciou este conteúdo; a denúncia está em análise"}
	}
	rid := id.New()
	if err := s.repo.CreateReport(ctx, rid, reporterID, in); err != nil {
		return nil, err
	}
	return s.repo.GetReport(ctx, rid)
}

// ListReports returns a filtered page of reports.
func (s *Service) ListReports(ctx context.Context, f ReportFilter) ([]Report, int64, error) {
	if f.Status != "" && !ValidStatus(f.Status) {
		return nil, 0, invalid("status inválido")
	}
	return s.repo.ListReports(ctx, f)
}

// ReportCounts returns report totals per status.
func (s *Service) ReportCounts(ctx context.Context) (map[string]int64, error) {
	return s.repo.ReportCounts(ctx)
}

// UpdateReport changes the status of a report and notifies the reporter
// when it is closed.
func (s *Service) UpdateReport(ctx context.Context, actorID, reportID string, in UpdateReportInput) (*Report, error) {
	if !ValidStatus(in.Status) {
		return nil, invalid("status inválido")
	}
	note := strings.TrimSpace(in.Note)
	if len([]rune(note)) > 1000 {
		return nil, invalid("a nota deve ter no máximo 1000 caracteres")
	}
	current, err := s.repo.GetReport(ctx, reportID)
	if err != nil {
		return nil, err
	}
	if err := s.repo.UpdateReportStatus(ctx, reportID, in.Status, actorID, note); err != nil {
		return nil, err
	}
	s.audit.Record(ctx, actorID, ActionReportStatus, "report", reportID, map[string]any{
		"from": current.Status, "to": in.Status, "note": note,
		"targetType": current.Target.Type, "targetId": current.Target.ID,
	})
	closed := in.Status == StatusResolved || in.Status == StatusRejected
	if current.Reporter != nil && current.Status != in.Status && closed {
		verdict := "foi analisada e resolvida"
		if in.Status == StatusRejected {
			verdict = "foi analisada e rejeitada"
		}
		s.notifyUser(ctx, current.Reporter.ID, notification.TypeAdmin, fmt.Sprintf("Sua denúncia (%s) %s.", current.Reason, verdict))
	}
	return s.repo.GetReport(ctx, reportID)
}

type conflictError struct{ msg string }

func (e *conflictError) Error() string { return e.msg }
func (e *conflictError) Unwrap() error { return apperr.ErrConflict }
