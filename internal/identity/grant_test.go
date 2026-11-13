package identity_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/identity"
)

// Synthetic IDs of grants.
const (
	grantToken  = "0011223344556677"
	grantPerson = "8899aabbccddeeff"
)

func grantAt() time.Time { return time.Date(2026, 11, 13, 8, 0, 0, 0, time.UTC) }

func tokenHolder(id string) identity.GrantHolder {
	return identity.GrantHolder{Kind: identity.GrantHolderScriptToken, ID: id}
}

func pluginHolder(id string) identity.GrantHolder {
	return identity.GrantHolder{Kind: identity.GrantHolderViewPlugin, ID: id}
}

func TestNewGrant(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		holder  identity.GrantHolder
		target  identity.GrantTarget
		wantErr bool
	}{
		{"a token an account", tokenHolder(grantToken), identity.GrantTarget{Kind: identity.GrantTargetAccount, ID: "main-account"}, false},
		{"a token a sleeve", tokenHolder(grantToken), identity.GrantTarget{Kind: identity.GrantTargetSleeve, ID: "core"}, false},
		{"a plugin the household", pluginHolder("ui"), identity.Household(), false},
		{"a person the household", identity.GrantHolder{Kind: identity.GrantHolderPerson, ID: grantPerson}, identity.Household(), false},
		{"an account ID a seed gave it", tokenHolder(grantToken), identity.GrantTarget{Kind: identity.GrantTargetAccount, ID: "Acc_1.b"}, false},

		{"no holder", identity.GrantHolder{}, identity.Household(), true},
		{"a holder of an unknown kind", identity.GrantHolder{Kind: "agent", ID: grantToken}, identity.Household(), true},
		{"a token named by a secret", tokenHolder(identity.SecretPrefix + "0011223344556677_AAAA"), identity.Household(), true},
		{"a token ID in capitals", tokenHolder("0011223344556677"[:8] + "ABCDEF01"), identity.Household(), true},
		{"a token ID too short", tokenHolder("0011"), identity.Household(), true},
		{"a plugin named by a sentence", pluginHolder("my plugin"), identity.Household(), true},
		{"a plugin name in capitals", pluginHolder("UI"), identity.Household(), true},
		{"a plugin name over 63 characters", pluginHolder(strings.Repeat("a", 64)), identity.Household(), true},
		{"a plugin name that ends in a dash", pluginHolder("ui-"), identity.Household(), true},
		{"no target", tokenHolder(grantToken), identity.GrantTarget{}, true},
		{"a target of an unknown kind", tokenHolder(grantToken), identity.GrantTarget{Kind: "fief", ID: "x"}, true},
		{"an account with no ID", tokenHolder(grantToken), identity.GrantTarget{Kind: identity.GrantTargetAccount}, true},
		{"an account named by a secret", tokenHolder(grantToken), identity.GrantTarget{Kind: identity.GrantTargetAccount, ID: identity.SecretPrefix + "abc"}, true},
		{"an account named by a session token", tokenHolder(grantToken), identity.GrantTarget{Kind: identity.GrantTargetAccount, ID: identity.SessionPrefix + "abc"}, true},
		{"an account ID with a space", tokenHolder(grantToken), identity.GrantTarget{Kind: identity.GrantTargetSleeve, ID: "my sleeve"}, true},
		{"an account ID with a slash", tokenHolder(grantToken), identity.GrantTarget{Kind: identity.GrantTargetSleeve, ID: "a/b"}, true},
		{"an account ID over 128 characters", tokenHolder(grantToken), identity.GrantTarget{Kind: identity.GrantTargetAccount, ID: strings.Repeat("a", 129)}, true},
		{"a household that is not home", tokenHolder(grantToken), identity.GrantTarget{Kind: identity.GrantTargetHousehold, ID: "away"}, true},
		{"a household with no ID", tokenHolder(grantToken), identity.GrantTarget{Kind: identity.GrantTargetHousehold}, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			g, err := identity.NewGrant(tc.holder, tc.target, grantAt())
			if tc.wantErr {
				wantInvalidGrant(t, err, tc.holder.ID)
				return
			}
			if err != nil {
				t.Fatalf("NewGrant() error = %v", err)
			}
			if !g.Active() || !g.CreatedAt.Equal(grantAt()) {
				t.Errorf("NewGrant() = %+v, want one in force, made at %v", g, grantAt())
			}
		})
	}
}

// wantInvalidGrant fails the test unless err wraps ErrInvalidGrant without
// quoting the holder's ID, which may be a secret in the wrong field (GR-12).
func wantInvalidGrant(t *testing.T, err error, holderID string) {
	t.Helper()
	if !errors.Is(err, identity.ErrInvalidGrant) {
		t.Fatalf("NewGrant() error = %v, want %v", err, identity.ErrInvalidGrant)
	}
	if len(holderID) > 4 && strings.Contains(err.Error(), holderID) {
		t.Errorf("the error %q quotes the holder's ID", err)
	}
}

func TestGrantRevoked(t *testing.T) {
	t.Parallel()
	g, err := identity.NewGrant(tokenHolder(grantToken), identity.Household(), grantAt())
	if err != nil {
		t.Fatalf("NewGrant: %v", err)
	}
	later := g.Revoked(grantAt().Add(time.Hour))
	if later.Active() || !later.RevokedAt.Equal(grantAt().Add(time.Hour)) || !g.Active() {
		t.Errorf("Revoked(+1h) = %+v of %+v, want the copy ended an hour in and the original in force", later, g)
	}
	before := g.Revoked(grantAt().Add(-time.Hour))
	if !before.RevokedAt.Equal(grantAt()) {
		t.Errorf("Revoked(-1h).RevokedAt = %v, want the grant's own time %v", before.RevokedAt, grantAt())
	}
	g.RevokedAt = grantAt().Add(-time.Second)
	if err := g.Validate(); !errors.Is(err, identity.ErrInvalidGrant) {
		t.Errorf("Validate() of a grant revoked before it was made = %v, want %v", err, identity.ErrInvalidGrant)
	}
}

func TestGrantMatch(t *testing.T) {
	t.Parallel()
	tok := identity.Grant{Holder: tokenHolder(grantToken), Target: identity.Household()}
	acc := identity.Grant{Holder: pluginHolder("ui"), Target: identity.GrantTarget{Kind: identity.GrantTargetAccount, ID: "a"}}
	tests := []struct {
		name  string
		match identity.GrantMatch
		want  [2]bool
	}{
		{"every grant", identity.GrantMatch{}, [2]bool{true, true}},
		{"one holder", identity.GrantMatch{Holder: tokenHolder(grantToken)}, [2]bool{true, false}},
		{"one target", identity.GrantMatch{Target: acc.Target}, [2]bool{false, true}},
		{"a holder and a target that don't go together", identity.GrantMatch{Holder: tok.Holder, Target: acc.Target}, [2]bool{false, false}},
		{"a holder and its target", identity.GrantMatch{Holder: acc.Holder, Target: acc.Target}, [2]bool{false, true}},
	}
	for _, tc := range tests {
		got := [2]bool{tc.match.Matches(&tok), tc.match.Matches(&acc)}
		if got != tc.want {
			t.Errorf("%s: matches %v, want %v", tc.name, got, tc.want)
		}
	}
	if err := (identity.GrantMatch{Holder: identity.GrantHolder{Kind: "x", ID: "y"}}).Validate(); !errors.Is(err, identity.ErrInvalidGrant) {
		t.Errorf("Validate() of a match with an unknown holder kind = %v, want %v", err, identity.ErrInvalidGrant)
	}
}

// TestGrantAuditEntries checks that a grant's entries name its holder, then
// its target, by ID and nothing else, and that nothing else passes for one.
func TestGrantAuditEntries(t *testing.T) {
	t.Parallel()
	owner := identity.Actor{Kind: identity.ActorOwner, ID: auditSession}
	system := identity.Actor{Kind: identity.ActorSystem, ID: identity.SystemActorID}
	acct := identity.GrantTarget{Kind: identity.GrantTargetAccount, ID: "main-account"}
	tests := []struct {
		name    string
		actor   identity.Actor
		action  identity.AuditAction
		refs    []identity.Ref
		wantErr bool
	}{
		{"created by the owner", owner, identity.ActionGrantCreated, []identity.Ref{tokenHolder(grantToken).Ref(), acct.Ref()}, false},
		{"revoked by the core", system, identity.ActionGrantRevoked, []identity.Ref{pluginHolder("ui").Ref(), identity.Household().Ref()}, false},
		{"a sleeve", owner, identity.ActionGrantCreated, []identity.Ref{tokenHolder(grantToken).Ref(), {Kind: identity.RefSleeve, ID: "core"}}, false},
		{"a person", owner, identity.ActionGrantCreated, []identity.Ref{{Kind: identity.RefPerson, ID: grantPerson}, identity.Household().Ref()}, false},

		{"no references", owner, identity.ActionGrantCreated, nil, true},
		{"the holder alone", owner, identity.ActionGrantCreated, []identity.Ref{tokenHolder(grantToken).Ref()}, true},
		{"three references", owner, identity.ActionGrantRevoked, []identity.Ref{tokenHolder(grantToken).Ref(), acct.Ref(), identity.Household().Ref()}, true},
		{"the target first", owner, identity.ActionGrantCreated, []identity.Ref{acct.Ref(), tokenHolder(grantToken).Ref()}, true},
		{"two holders", owner, identity.ActionGrantCreated, []identity.Ref{tokenHolder(grantToken).Ref(), pluginHolder("ui").Ref()}, true},
		{"a session as the target", owner, identity.ActionGrantCreated, []identity.Ref{tokenHolder(grantToken).Ref(), {Kind: identity.RefSession, ID: auditSession}}, true},
		{"an account named by a sentence", owner, identity.ActionGrantCreated, []identity.Ref{tokenHolder(grantToken).Ref(), {Kind: identity.RefAccount, ID: "my brokerage"}}, true},
		{"an account named by a secret", owner, identity.ActionGrantCreated, []identity.Ref{tokenHolder(grantToken).Ref(), {Kind: identity.RefAccount, ID: identity.SecretPrefix + "x"}}, true},
		{"a household that is not home", owner, identity.ActionGrantCreated, []identity.Ref{tokenHolder(grantToken).Ref(), {Kind: identity.RefHousehold, ID: "x"}}, true},
		{"a plugin named by a sentence", owner, identity.ActionGrantCreated, []identity.Ref{{Kind: identity.RefViewPlugin, ID: "my plugin"}, identity.Household().Ref()}, true},
		{"a grant named by a script token entry's kind", owner, identity.ActionScriptTokenCreated, []identity.Ref{{Kind: identity.RefAccount, ID: "main-account"}}, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := identity.NewAuditEntry(auditAt(), tc.actor, tc.action, identity.OutcomeOK, tc.refs...)
			if tc.wantErr != (err != nil) || (err != nil && !errors.Is(err, identity.ErrInvalidAuditEntry)) {
				t.Errorf("NewAuditEntry() error = %v, want an error: %v", err, tc.wantErr)
			}
		})
	}
	if diff := cmp.Diff([]identity.Ref{{Kind: identity.RefViewPlugin, ID: "ui"}, {Kind: identity.RefHousehold, ID: "home"}},
		[]identity.Ref{pluginHolder("ui").Ref(), identity.Household().Ref()}); diff != "" {
		t.Errorf("refs of a grant (-want +got):\n%s", diff)
	}
	if _, err := identity.ParseAuditAction("grant_created"); err != nil {
		t.Errorf("ParseAuditAction(grant_created) = %v", err)
	}
}
