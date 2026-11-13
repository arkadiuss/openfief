package main

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"

	attentiondomain "github.com/arkadiuss/openfief/internal/attention"
	"github.com/arkadiuss/openfief/internal/identity"
	"github.com/arkadiuss/openfief/internal/views"
	viewsapp "github.com/arkadiuss/openfief/internal/views/app"
)

// errGrantsDown is the error of a grants table that can't be read.
var errGrantsDown = errors.New("grants down")

// grantsHeld is a grantsInForce over the grants it holds, or one that fails.
type grantsHeld struct {
	grants []identity.Grant
	err    error
}

func (g grantsHeld) InForce(context.Context) ([]identity.Grant, error) { return g.grants, g.err }

// pluginGrant is a grant of target to the view plugin instance.
func pluginGrant(instance string, target identity.GrantTarget) identity.Grant {
	return identity.Grant{Holder: identity.GrantHolder{Kind: identity.GrantHolderViewPlugin, ID: instance}, Target: target}
}

// readyView is a view plugin that runs: Ready, with the view capability.
func readyView(instance string) views.PluginRecord {
	return views.PluginRecord{Instance: instance, Requested: []string{"view"}, Effective: []string{"view"}, Ready: views.Condition{Status: "True"}}
}

// ungranted is the condition of an ungranted view plugin instance.
func ungranted(instance string) attentiondomain.Condition {
	return attentiondomain.Condition{
		Kind: attentiondomain.KindViewPluginUngranted, Subject: instance, Detail: instance + " can't read anything until you grant it accounts",
	}
}

// TestUngrantedViewPluginsCheck checks D-196's item for a view plugin that
// can't read anything while SpiceDB answers: each installed view plugin that is
// Ready and holds no grant in force is a condition naming its instance, one
// with a grant of any target is not, nor is a plugin that is not Ready, being
// removed, without the view capability or the one serving browsers (the
// built-in UI reads with the signed-in person's session), and with the policy,
// or outside a cluster, there is none.
func TestUngrantedViewPluginsCheck(t *testing.T) {
	t.Parallel()
	notReady := readyView("starting")
	notReady.Ready = views.Condition{Status: "False", Reason: "WaitingForSecrets"}
	removing := readyView("removing")
	removing.Deleting = true
	asksOnly := readyView("asks-only")
	asksOnly.Effective = nil
	builtIn := readyView("ui")
	builtIn.BrowserPath = "/"
	connector := views.PluginRecord{Instance: "ibkr", Effective: []string{"connector"}, Ready: views.Condition{Status: "True"}}
	home := identity.Household()
	tests := []struct {
		name     string
		spicedb  bool
		records  []views.PluginRecord
		grants   []identity.Grant
		outside  bool
		want     []attentiondomain.Condition
		wantFail bool
	}{
		{name: "a view plugin with no grant", spicedb: true, records: []views.PluginRecord{readyView("ghostfolio")}, want: []attentiondomain.Condition{ungranted("ghostfolio")}},
		{
			name: "a grant of the household", spicedb: true, records: []views.PluginRecord{readyView("ghostfolio")},
			grants: []identity.Grant{pluginGrant("ghostfolio", home)},
		},
		{
			name: "a grant of one account is a grant", spicedb: true, records: []views.PluginRecord{readyView("ghostfolio")},
			grants: []identity.Grant{pluginGrant("ghostfolio", identity.GrantTarget{Kind: identity.GrantTargetAccount, ID: "main"})},
		},
		{
			name: "another plugin's grant is not its own", spicedb: true, records: []views.PluginRecord{readyView("ghostfolio"), readyView("second")},
			grants: []identity.Grant{pluginGrant("second", home)}, want: []attentiondomain.Condition{ungranted("ghostfolio")},
		},
		{
			name: "a script token's grant is not a plugin's", spicedb: true, records: []views.PluginRecord{readyView("ghostfolio")},
			grants: []identity.Grant{{Holder: identity.GrantHolder{Kind: identity.GrantHolderScriptToken, ID: "ghostfolio"}, Target: home}},
			want:   []attentiondomain.Condition{ungranted("ghostfolio")},
		},
		{
			name: "two ungranted plugins", spicedb: true, records: []views.PluginRecord{readyView("a"), readyView("b")},
			want: []attentiondomain.Condition{ungranted("a"), ungranted("b")},
		},
		{
			name: "plugins that are not asked for", spicedb: true,
			records: []views.PluginRecord{notReady, removing, asksOnly, builtIn, connector},
		},
		{name: "the policy needs no grant", records: []views.PluginRecord{readyView("ghostfolio")}},
		{name: "outside a cluster", spicedb: true, outside: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var installed installedPlugins
			if !tt.outside {
				installed = &fakePlugins{records: tt.records}
			}
			check := ungrantedViewPlugins(tt.spicedb, installed, grantsHeld{grants: tt.grants})
			if check.Kind != attentiondomain.KindViewPluginUngranted {
				t.Errorf("Kind = %q, want %q", check.Kind, attentiondomain.KindViewPluginUngranted)
			}
			got, err := check.Find(t.Context())
			if err != nil {
				t.Fatalf("Find: %v", err)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("conditions (-want +got):\n%s", diff)
			}
		})
	}
}

// TestUngrantedViewPluginsCheckNeverGuesses checks that a check that can't tell
// fails instead of finding nothing, which would clear every item: the plugin
// cache not synced yet, or the grants not readable.
func TestUngrantedViewPluginsCheckNeverGuesses(t *testing.T) {
	t.Parallel()
	unsynced := ungrantedViewPlugins(true, &fakePlugins{unsynced: true}, grantsHeld{})
	if _, err := unsynced.Find(t.Context()); !errors.Is(err, viewsapp.ErrPluginsUnknown) {
		t.Errorf("Find before the plugins are known: %v, want ErrPluginsUnknown", err)
	}
	down := ungrantedViewPlugins(true, &fakePlugins{records: []views.PluginRecord{readyView("ghostfolio")}}, grantsHeld{err: errGrantsDown})
	if _, err := down.Find(t.Context()); !errors.Is(err, errGrantsDown) {
		t.Errorf("Find with the grants down: %v, want the grants' error", err)
	}
}

// TestUngrantedViewPluginsInOutsideACluster checks that a core with no cluster,
// whose plugins are a nil pointer, finds nothing instead of calling it, with
// SpiceDB or without.
func TestUngrantedViewPluginsInOutsideACluster(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{authzPolicy, authzSpiceDB} {
		check := ungrantedViewPluginsIn(&serveConfig{Authz: authzConfig{Mode: mode}}, nil, grantsHeld{})
		if got, err := check.Find(t.Context()); err != nil || len(got) != 0 {
			t.Errorf("Find with --authz=%s and no cluster = %v, %v, want nothing", mode, got, err)
		}
	}
}
