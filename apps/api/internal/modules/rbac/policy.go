package rbac

import (
	"fmt"
	"regexp"

	"github.com/forum-framework/api/internal/shared/apperr"
)

// OwnerRank is the hierarchy rank of the Owner role. No other role may reach it.
const OwnerRank = 100

// Actor is the resolved authority of a user performing an administrative
// action: its highest role rank, whether it is the Owner and its permissions.
type Actor struct {
	ID      string
	Rank    int
	IsOwner bool
	Perms   map[string]bool
}

// policyError wraps a sentinel with a user-facing message.
type policyError struct {
	msg  string
	kind error
}

func (e *policyError) Error() string { return e.msg }
func (e *policyError) Unwrap() error { return e.kind }

func forbidden(msg string) error { return &policyError{msg: msg, kind: apperr.ErrForbidden} }
func invalid(msg string) error   { return &policyError{msg: msg, kind: apperr.ErrValidation} }

// CanManageRole reports whether the actor may edit, delete or change the
// permissions of role. The Owner role is immutable; other roles can only be
// managed by actors ranked strictly above them.
func (a Actor) CanManageRole(role Role) error {
	if role.Slug == RoleOwner {
		return forbidden("o cargo Dono é imutável")
	}
	if role.RankOrder >= a.Rank {
		return forbidden("você só pode gerenciar cargos abaixo do seu na hierarquia")
	}
	return nil
}

// CanUseRank reports whether the actor may create or move a role to rank.
func (a Actor) CanUseRank(rank int) error {
	if rank < 1 {
		return invalid("a hierarquia deve ser um número maior que zero")
	}
	if rank >= OwnerRank || rank >= a.Rank {
		return forbidden("a hierarquia do cargo deve ficar abaixo da sua")
	}
	return nil
}

// CanGrant reports whether the actor holds every permission it tries to grant
// (nobody can hand out a capability they do not have).
func (a Actor) CanGrant(perms []string) error {
	if a.IsOwner {
		return nil
	}
	for _, p := range perms {
		if !a.Perms[p] {
			return forbidden(fmt.Sprintf("você não pode conceder a permissão %q, que não possui", p))
		}
	}
	return nil
}

// CanActOnUser reports whether the actor may change roles of, or sanction,
// a target user with the given highest rank.
func (a Actor) CanActOnUser(targetID string, targetRank int, targetIsOwner bool) error {
	if targetID == a.ID {
		return forbidden("você não pode aplicar esta ação a si mesmo")
	}
	if targetIsOwner {
		return forbidden("a conta do Dono não pode ser alterada por outros cargos")
	}
	if targetRank >= a.Rank {
		return forbidden("o usuário possui hierarquia igual ou superior à sua")
	}
	return nil
}

// Diff returns the elements added to and removed from old to reach next.
func Diff(old, next []string) (added, removed []string) {
	oldSet := make(map[string]bool, len(old))
	for _, s := range old {
		oldSet[s] = true
	}
	nextSet := make(map[string]bool, len(next))
	for _, s := range next {
		nextSet[s] = true
		if !oldSet[s] {
			added = append(added, s)
		}
	}
	for _, s := range old {
		if !nextSet[s] {
			removed = append(removed, s)
		}
	}
	return added, removed
}

var roleSlugRe = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{1,48}[a-z0-9]$`)

// ValidRoleSlug checks a custom role or permission slug.
func ValidRoleSlug(slug string) error {
	if !roleSlugRe.MatchString(slug) {
		return invalid("slug inválido (use 3 a 50 letras minúsculas, números ou hífens)")
	}
	return nil
}
