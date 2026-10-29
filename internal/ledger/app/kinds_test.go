package app_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/ledger/app"
	"github.com/arkadiuss/openfief/internal/portfolio"
)

func TestSaveKind(t *testing.T) {
	t.Parallel()
	defaults := portfolio.DefaultKinds()
	crypto := portfolio.Kind{Key: "crypto", Name: "Crypto", DefaultRole: portfolio.RoleOffensive}
	renamed := portfolio.Kind{Key: portfolio.KindRetirement, Name: "IKE and IKZE", DefaultRole: portfolio.RoleDefensive}
	tests := []struct {
		name    string
		kind    portfolio.Kind
		wantErr error
		want    []portfolio.Kind
	}{
		{
			name: "custom kind",
			kind: crypto,
			want: []portfolio.Kind{defaults[0], defaults[1], defaults[2], crypto, defaults[3], defaults[4], defaults[5]},
		},
		{
			name: "rename a built-in kind",
			kind: renamed,
			want: []portfolio.Kind{defaults[0], defaults[1], defaults[2], defaults[3], renamed, defaults[5]},
		},
		{
			name:    "change a kind's default role",
			kind:    portfolio.Kind{Key: portfolio.KindCore, Name: "Core", DefaultRole: portfolio.RoleOffensive},
			wantErr: portfolio.ErrKindRoleChange,
			want:    defaults,
		},
		{
			name:    "invalid kind",
			kind:    portfolio.Kind{Key: "Crypto", Name: "Crypto", DefaultRole: portfolio.RoleOffensive},
			wantErr: portfolio.ErrInvalidKind,
			want:    defaults,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			l, _ := seeded(t)
			err := l.SaveKind(t.Context(), tt.kind)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("SaveKind error (-want +got):\n%s", diff)
			}
			got, err := l.Kinds(t.Context())
			checkResult(t, "Kinds", tt.want, got, err)
		})
	}
}

func TestEnsureDefaults(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	l := app.New(newMemStore())
	for range 2 {
		if err := l.EnsureDefaults(ctx); err != nil {
			t.Fatalf("EnsureDefaults: %v", err)
		}
		kinds, err := l.Kinds(ctx)
		checkResult(t, "Kinds", portfolio.DefaultKinds(), kinds, err)
		sleeves, err := l.Sleeves(ctx)
		checkResult(t, "Sleeves", portfolio.DefaultSleeves(), sleeves, err)
	}

	// The owner renames a kind and moves a default sleeve to another kind;
	// ensuring the defaults again keeps both changes.
	renamed := portfolio.Kind{Key: portfolio.KindCore, Name: "Long term", DefaultRole: portfolio.RoleDefensive}
	if err := l.SaveKind(ctx, renamed); err != nil {
		t.Fatalf("SaveKind: %v", err)
	}
	moved := portfolio.Sleeve{ID: "satellite", Name: "Satellite", Kind: portfolio.KindAgentPool, Role: portfolio.RoleOffensive}
	if err := l.SaveSleeve(ctx, moved); err != nil {
		t.Fatalf("SaveSleeve: %v", err)
	}
	if err := l.EnsureDefaults(ctx); err != nil {
		t.Fatalf("EnsureDefaults after changes: %v", err)
	}
	wantKinds := portfolio.DefaultKinds()
	wantKinds[2] = renamed
	kinds, err := l.Kinds(ctx)
	checkResult(t, "Kinds after changes", wantKinds, kinds, err)
	wantSleeves := portfolio.DefaultSleeves()
	wantSleeves[2] = moved
	sleeves, err := l.Sleeves(ctx)
	checkResult(t, "Sleeves after changes", wantSleeves, sleeves, err)
}

func TestEnsureDefaultsOnlyOnAFreshInstall(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	// The demo dataset, or the owner, stored sleeves of their own before
	// the defaults were ensured: the install has started, so only the
	// built-in kinds are added (D-71: an install starts with the defaults).
	own := portfolio.Sleeve{ID: "long-term", Name: "Long term", Kind: portfolio.KindCore, Role: portfolio.RoleDefensive}
	fresh := app.New(newMemStore())
	if err := fresh.SaveKind(ctx, portfolio.DefaultKinds()[2]); err != nil {
		t.Fatalf("SaveKind: %v", err)
	}
	if err := fresh.SaveSleeve(ctx, own); err != nil {
		t.Fatalf("SaveSleeve: %v", err)
	}
	if err := fresh.EnsureDefaults(ctx); err != nil {
		t.Fatalf("EnsureDefaults: %v", err)
	}
	kinds, err := fresh.Kinds(ctx)
	checkResult(t, "Kinds", portfolio.DefaultKinds(), kinds, err)
	sleeves, err := fresh.Sleeves(ctx)
	checkResult(t, "Sleeves", []portfolio.Sleeve{own}, sleeves, err)
}
