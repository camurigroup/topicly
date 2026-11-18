package rbac

import (
	"errors"
	"testing"

	"github.com/forum-framework/api/internal/shared/apperr"
)

func actor(rank int, owner bool, perms ...string) Actor {
	set := make(map[string]bool, len(perms))
	for _, p := range perms {
		set[p] = true
	}
	return Actor{ID: "a", Rank: rank, IsOwner: owner, Perms: set}
}

func TestCanManageRole(t *testing.T) {
	owner := actor(100, true)
	admin := actor(50, false)
	cases := []struct {
		name  string
		actor Actor
		role  Role
		ok    bool
	}{
		{"owner edits admin", owner, Role{Slug: RoleAdmin, RankOrder: 50}, true},
		{"owner cannot edit owner role", owner, Role{Slug: RoleOwner, RankOrder: 100, IsSystem: true}, false},
		{"admin edits lower role", admin, Role{Slug: "helper", RankOrder: 10}, true},
		{"admin cannot edit same rank", admin, Role{Slug: RoleAdmin, RankOrder: 50}, false},
		{"admin cannot edit higher rank", admin, Role{Slug: "senior", RankOrder: 70}, false},
	}
	for _, c := range cases {
		err := c.actor.CanManageRole(c.role)
		if c.ok && err != nil {
			t.Errorf("%s: unexpected error %v", c.name, err)
		}
		if !c.ok && !errors.Is(err, apperr.ErrForbidden) {
			t.Errorf("%s: expected forbidden, got %v", c.name, err)
		}
	}
}

func TestCanUseRank(t *testing.T) {
	admin := actor(50, false)
	if err := admin.CanUseRank(49); err != nil {
		t.Fatalf("rank below actor should be allowed: %v", err)
	}
	for _, r := range []int{50, 51, 100} {
		if err := admin.CanUseRank(r); !errors.Is(err, apperr.ErrForbidden) {
			t.Errorf("rank %d should be forbidden for rank 50, got %v", r, err)
		}
	}
	if err := admin.CanUseRank(0); !errors.Is(err, apperr.ErrValidation) {
		t.Errorf("rank 0 should be invalid, got %v", err)
	}
	if err := actor(100, true).CanUseRank(100); !errors.Is(err, apperr.ErrForbidden) {
		t.Errorf("nobody may create a role at owner rank, got %v", err)
	}
}

func TestCanGrant(t *testing.T) {
	mod := actor(20, false, PermUsersBan, PermReportsManage)
	if err := mod.CanGrant([]string{PermUsersBan}); err != nil {
		t.Fatalf("granting a held permission should be allowed: %v", err)
	}
	if err := mod.CanGrant([]string{PermUsersBan, PermRolesManage}); !errors.Is(err, apperr.ErrForbidden) {
		t.Fatalf("granting a permission not held must be forbidden, got %v", err)
	}
	if err := actor(100, true).CanGrant([]string{PermRolesManage, PermPermissionsManage}); err != nil {
		t.Fatalf("owner may grant anything: %v", err)
	}
}

func TestCanActOnUser(t *testing.T) {
	admin := actor(50, false)
	if err := admin.CanActOnUser("b", 20, false); err != nil {
		t.Fatalf("admin over moderator should be allowed: %v", err)
	}
	if err := admin.CanActOnUser("b", 50, false); !errors.Is(err, apperr.ErrForbidden) {
		t.Fatalf("same rank must be forbidden, got %v", err)
	}
	if err := admin.CanActOnUser("a", 0, false); !errors.Is(err, apperr.ErrForbidden) {
		t.Fatalf("acting on self must be forbidden, got %v", err)
	}
	if err := actor(100, true).CanActOnUser("b", 100, true); !errors.Is(err, apperr.ErrForbidden) {
		t.Fatalf("owner account is immutable, got %v", err)
	}
}

func TestDiff(t *testing.T) {
	added, removed := Diff([]string{"a", "b"}, []string{"b", "c"})
	if len(added) != 1 || added[0] != "c" || len(removed) != 1 || removed[0] != "a" {
		t.Fatalf("unexpected diff added=%v removed=%v", added, removed)
	}
}

func TestValidRoleSlug(t *testing.T) {
	for _, ok := range []string{"helper", "vip-member", "mod2"} {
		if err := ValidRoleSlug(ok); err != nil {
			t.Errorf("%q should be valid: %v", ok, err)
		}
	}
	for _, bad := range []string{"", "a", "Has Space", "UPPER", "-lead", "x_y"} {
		if err := ValidRoleSlug(bad); err == nil {
			t.Errorf("%q should be invalid", bad)
		}
	}
}
