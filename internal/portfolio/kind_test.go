package portfolio_test

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/portfolio"
)

func TestKindValidate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		kind    portfolio.Kind
		wantErr error
	}{
		{name: "one word", kind: portfolio.Kind{Key: "bonds", Name: "Bonds", DefaultRole: portfolio.RoleDefensive}},
		{name: "hyphenated with digits", kind: portfolio.Kind{Key: "ikze-2", Name: "IKZE 2", DefaultRole: portfolio.RoleDefensive}},
		{name: "longest key", kind: portfolio.Kind{Key: portfolio.KindKey(strings.Repeat("a", 64)), Name: "A", DefaultRole: portfolio.RoleOffensive}},
		{name: "key too long", kind: portfolio.Kind{Key: portfolio.KindKey(strings.Repeat("a", 65)), Name: "A", DefaultRole: portfolio.RoleOffensive}, wantErr: portfolio.ErrInvalidKind},
		{name: "empty key", kind: portfolio.Kind{Name: "Bonds", DefaultRole: portfolio.RoleDefensive}, wantErr: portfolio.ErrInvalidKind},
		{name: "uppercase key", kind: portfolio.Kind{Key: "Bonds", Name: "Bonds", DefaultRole: portfolio.RoleDefensive}, wantErr: portfolio.ErrInvalidKind},
		{name: "space in key", kind: portfolio.Kind{Key: "my bonds", Name: "Bonds", DefaultRole: portfolio.RoleDefensive}, wantErr: portfolio.ErrInvalidKind},
		{name: "leading hyphen", kind: portfolio.Kind{Key: "-bonds", Name: "Bonds", DefaultRole: portfolio.RoleDefensive}, wantErr: portfolio.ErrInvalidKind},
		{name: "double hyphen", kind: portfolio.Kind{Key: "agent--pool", Name: "Pool", DefaultRole: portfolio.RoleOffensive}, wantErr: portfolio.ErrInvalidKind},
		{name: "trailing hyphen", kind: portfolio.Kind{Key: "bonds-", Name: "Bonds", DefaultRole: portfolio.RoleDefensive}, wantErr: portfolio.ErrInvalidKind},
		{name: "no name", kind: portfolio.Kind{Key: "bonds", DefaultRole: portfolio.RoleDefensive}, wantErr: portfolio.ErrInvalidKind},
		{name: "no default role", kind: portfolio.Kind{Key: "bonds", Name: "Bonds"}, wantErr: portfolio.ErrInvalidKind},
		{name: "unknown default role", kind: portfolio.Kind{Key: "bonds", Name: "Bonds", DefaultRole: portfolio.Role(3)}, wantErr: portfolio.ErrInvalidKind},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.kind.Validate()
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Validate() error (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDefaultKinds(t *testing.T) {
	t.Parallel()
	got := portfolio.DefaultKinds()
	want := map[portfolio.KindKey]portfolio.Role{
		portfolio.KindAgentPool:     portfolio.RoleOffensive,
		portfolio.KindBonds:         portfolio.RoleDefensive,
		portfolio.KindCore:          portfolio.RoleDefensive,
		portfolio.KindEmergencyFund: portfolio.RoleDefensive,
		portfolio.KindRetirement:    portfolio.RoleDefensive,
		portfolio.KindSatellite:     portfolio.RoleOffensive,
	}
	roles := map[portfolio.KindKey]portfolio.Role{}
	for i, k := range got {
		if err := k.Validate(); err != nil {
			t.Errorf("default kind %s: %v", k.Key, err)
		}
		if i > 0 && got[i-1].Key >= k.Key {
			t.Errorf("default kinds not in key order: %s before %s", got[i-1].Key, k.Key)
		}
		roles[k.Key] = k.DefaultRole
	}
	if diff := cmp.Diff(want, roles); diff != "" {
		t.Errorf("default kinds' roles (-want +got):\n%s", diff)
	}
	got[0].Name = "changed"
	if portfolio.DefaultKinds()[0].Name == "changed" {
		t.Error("DefaultKinds shares its slice between calls")
	}
}

func TestCheckKind(t *testing.T) {
	t.Parallel()
	stored := portfolio.DefaultKinds()
	tests := []struct {
		name    string
		kind    portfolio.Kind
		wantErr error
	}{
		{name: "new custom kind", kind: portfolio.Kind{Key: "crypto", Name: "Crypto", DefaultRole: portfolio.RoleOffensive}},
		{name: "rename a built-in kind", kind: portfolio.Kind{Key: portfolio.KindRetirement, Name: "IKE and IKZE", DefaultRole: portfolio.RoleDefensive}},
		{name: "make a defensive kind offensive", kind: portfolio.Kind{Key: portfolio.KindBonds, Name: "Bonds", DefaultRole: portfolio.RoleOffensive}, wantErr: portfolio.ErrKindRoleChange},
		{name: "make an offensive kind defensive", kind: portfolio.Kind{Key: portfolio.KindAgentPool, Name: "Agent pool", DefaultRole: portfolio.RoleDefensive}, wantErr: portfolio.ErrKindRoleChange},
		{name: "invalid kind", kind: portfolio.Kind{Key: "Crypto", Name: "Crypto", DefaultRole: portfolio.RoleOffensive}, wantErr: portfolio.ErrInvalidKind},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := portfolio.CheckKind(stored, tt.kind)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("CheckKind error (-want +got):\n%s", diff)
			}
		})
	}
}

func TestCheckSleeve(t *testing.T) {
	t.Parallel()
	kinds := portfolio.DefaultKinds()
	sleeve := func(kind portfolio.KindKey, role portfolio.Role, override bool) portfolio.Sleeve {
		return portfolio.Sleeve{ID: "s", Name: "S", Kind: kind, Role: role, RoleOverride: override}
	}
	tests := []struct {
		name    string
		sleeve  portfolio.Sleeve
		want    portfolio.Sleeve
		wantErr error
	}{
		{
			name:   "role defaults from a defensive kind",
			sleeve: sleeve(portfolio.KindCore, 0, false),
			want:   sleeve(portfolio.KindCore, portfolio.RoleDefensive, false),
		},
		{
			name:   "role defaults from an offensive kind",
			sleeve: sleeve(portfolio.KindAgentPool, 0, false),
			want:   sleeve(portfolio.KindAgentPool, portfolio.RoleOffensive, false),
		},
		{
			name:   "kind's own role",
			sleeve: sleeve(portfolio.KindSatellite, portfolio.RoleOffensive, false),
			want:   sleeve(portfolio.KindSatellite, portfolio.RoleOffensive, false),
		},
		{
			name:   "defensive in an offensive kind with the override",
			sleeve: sleeve(portfolio.KindSatellite, portfolio.RoleDefensive, true),
			want:   sleeve(portfolio.KindSatellite, portfolio.RoleDefensive, true),
		},
		{
			name:    "defensive in an offensive kind without the override",
			sleeve:  sleeve(portfolio.KindSatellite, portfolio.RoleDefensive, false),
			wantErr: portfolio.ErrRoleOverride,
		},
		{
			name:   "offensive in a defensive kind with the override",
			sleeve: sleeve(portfolio.KindCore, portfolio.RoleOffensive, true),
			want:   sleeve(portfolio.KindCore, portfolio.RoleOffensive, true),
		},
		{
			name:    "offensive in a defensive kind without the override",
			sleeve:  sleeve(portfolio.KindRetirement, portfolio.RoleOffensive, false),
			wantErr: portfolio.ErrRoleOverride,
		},
		{
			name:    "override on a defensive sleeve of a defensive kind",
			sleeve:  sleeve(portfolio.KindCore, portfolio.RoleDefensive, true),
			wantErr: portfolio.ErrRoleOverride,
		},
		{
			name:    "override on an offensive sleeve of an offensive kind",
			sleeve:  sleeve(portfolio.KindAgentPool, portfolio.RoleOffensive, true),
			wantErr: portfolio.ErrRoleOverride,
		},
		{
			name:    "override with no role, which defaults to defensive",
			sleeve:  sleeve(portfolio.KindBonds, 0, true),
			wantErr: portfolio.ErrRoleOverride,
		},
		{
			name:    "no kind",
			sleeve:  sleeve("", portfolio.RoleDefensive, false),
			wantErr: portfolio.ErrUnknownKind,
		},
		{
			name:    "kind not stored",
			sleeve:  sleeve("crypto", portfolio.RoleOffensive, false),
			wantErr: portfolio.ErrUnknownKind,
		},
		{
			name:    "unknown role",
			sleeve:  sleeve(portfolio.KindCore, portfolio.Role(3), false),
			wantErr: portfolio.ErrInvalidSleeve,
		},
		{
			name:    "no name",
			sleeve:  portfolio.Sleeve{ID: "s", Kind: portfolio.KindCore},
			wantErr: portfolio.ErrInvalidSleeve,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := portfolio.CheckSleeve(kinds, tt.sleeve)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("CheckSleeve error (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("CheckSleeve (-want +got):\n%s", diff)
			}
		})
	}
}

func TestDefaultSleeves(t *testing.T) {
	t.Parallel()
	want := []portfolio.Sleeve{
		{ID: "bonds", Name: "Bonds", Kind: portfolio.KindBonds, Role: portfolio.RoleDefensive},
		{ID: "core", Name: "Core", Kind: portfolio.KindCore, Role: portfolio.RoleDefensive},
		{ID: "satellite", Name: "Satellite", Kind: portfolio.KindSatellite, Role: portfolio.RoleOffensive},
	}
	got := portfolio.DefaultSleeves()
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("DefaultSleeves (-want +got):\n%s", diff)
	}
	for _, s := range got {
		checked, err := portfolio.CheckSleeve(portfolio.DefaultKinds(), s)
		if err != nil {
			t.Errorf("default sleeve %s: %v", s.ID, err)
		}
		if diff := cmp.Diff(s, checked); diff != "" {
			t.Errorf("default sleeve %s changed by CheckSleeve (-want +got):\n%s", s.ID, diff)
		}
	}
}

func TestMissingDefaults(t *testing.T) {
	t.Parallel()
	kinds, sleeves := portfolio.DefaultKinds(), portfolio.DefaultSleeves()
	renamedCore := portfolio.Kind{Key: portfolio.KindCore, Name: "Long term", DefaultRole: portfolio.RoleDefensive}
	movedCore := portfolio.Sleeve{ID: "core", Name: "Core", Kind: portfolio.KindRetirement, Role: portfolio.RoleDefensive}
	custom := portfolio.Kind{Key: "crypto", Name: "Crypto", DefaultRole: portfolio.RoleOffensive}
	tests := []struct {
		name        string
		kinds       []portfolio.Kind
		sleeves     []portfolio.Sleeve
		wantKinds   []portfolio.Kind
		wantSleeves []portfolio.Sleeve
	}{
		{name: "fresh install", wantKinds: kinds, wantSleeves: sleeves},
		{name: "everything stored", kinds: kinds, sleeves: sleeves},
		{name: "kinds stored by the migration, no sleeve yet", kinds: kinds, wantSleeves: sleeves},
		{
			name:      "renamed kind, moved sleeve and custom kind are kept; no default sleeve is added to a started install",
			kinds:     []portfolio.Kind{custom, renamedCore},
			sleeves:   []portfolio.Sleeve{movedCore},
			wantKinds: []portfolio.Kind{kinds[0], kinds[1], kinds[3], kinds[4], kinds[5]},
		},
		{
			name:    "the demo dataset's own sleeves are not joined by the defaults",
			kinds:   kinds,
			sleeves: []portfolio.Sleeve{{ID: "demo-defensive", Name: "Demo defensive", Kind: portfolio.KindBonds, Role: portfolio.RoleDefensive}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotKinds, gotSleeves := portfolio.MissingDefaults(tt.kinds, tt.sleeves)
			if diff := cmp.Diff(tt.wantKinds, gotKinds, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("missing kinds (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.wantSleeves, gotSleeves, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("missing sleeves (-want +got):\n%s", diff)
			}
		})
	}
}

func TestKindBySleeve(t *testing.T) {
	t.Parallel()
	want := map[portfolio.SleeveID]portfolio.KindKey{
		"bonds":     portfolio.KindBonds,
		"core":      portfolio.KindCore,
		"satellite": portfolio.KindSatellite,
	}
	if diff := cmp.Diff(want, portfolio.KindBySleeve(portfolio.DefaultSleeves())); diff != "" {
		t.Errorf("KindBySleeve (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(map[portfolio.SleeveID]portfolio.KindKey{}, portfolio.KindBySleeve(nil)); diff != "" {
		t.Errorf("KindBySleeve(nil) (-want +got):\n%s", diff)
	}
}
