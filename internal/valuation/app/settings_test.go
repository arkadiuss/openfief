package app_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/jurisdiction"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/valuation"
	"github.com/arkadiuss/openfief/internal/valuation/app"
)

// newSettings returns the settings use cases over store and rec, with the
// Polish pack, PLN reporting and ownerToken as the owner's session.
func newSettings(t *testing.T, store *fakeStore, rec *fakeRecorder) *app.Settings {
	t.Helper()
	l := twoAccounts(t)
	pack := pl.NewPack()
	return app.NewSettings(app.NewInputRecorder(app.NewPlanner(l, l), rec, pack), store, fakeOwners{err: app.ErrNotOwner}, pack, valuation.PLN())
}

func TestSettingsDisplay(t *testing.T) {
	t.Parallel()
	store := newFakeStore()
	s := newSettings(t, store, &fakeRecorder{})
	got, err := s.Display(t.Context())
	want := app.DisplaySettings{ReportingCurrency: valuation.PLN(), NumberFormat: pl.NewPack().NumberFormat(), NumberFormatDefault: true}
	if err != nil {
		t.Fatalf("Display: %v", err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("the pack's default (-want +got):\n%s", diff)
	}
	chosen := jurisdiction.NumberFormat{Group: jurisdiction.SeparatorComma, Decimal: jurisdiction.SeparatorDot}
	store.format = &chosen
	got, err = s.Display(t.Context())
	if err != nil {
		t.Fatalf("Display: %v", err)
	}
	if diff := cmp.Diff(app.DisplaySettings{ReportingCurrency: valuation.PLN(), NumberFormat: chosen}, got); diff != "" {
		t.Errorf("the owner's choice (-want +got):\n%s", diff)
	}
}

func TestOwnerSettings(t *testing.T) {
	t.Parallel()
	store := newFakeStore()
	store.settings = jurisdiction.Settings{MarginalRate: dec(t, "0.32"), HasMarginalRate: true}
	s := newSettings(t, store, &fakeRecorder{})
	got, err := s.Owner(t.Context(), "https://fief.example.org", ownerToken)
	if err != nil {
		t.Fatalf("Owner: %v", err)
	}
	want := app.OwnerSettings{
		DisplaySettings: app.DisplaySettings{ReportingCurrency: valuation.PLN(), NumberFormat: pl.NewPack().NumberFormat(), NumberFormatDefault: true},
		Wrappers:        []jurisdiction.Wrapper{"taxable", "ike", "ikze", "oki"},
		Values:          store.settings,
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("(-want +got):\n%s", diff)
	}
	if _, err := s.Owner(t.Context(), "https://fief.example.org", "read-session"); !cmp.Equal(app.ErrNotOwner, err, cmpopts.EquateErrors()) {
		t.Errorf("a read session: error = %v, want ErrNotOwner", err)
	}
}

func TestChangeSettings(t *testing.T) {
	t.Parallel()
	comma := jurisdiction.NumberFormat{Group: jurisdiction.SeparatorComma, Decimal: jurisdiction.SeparatorDot}
	tests := []struct {
		name    string
		token   string
		change  func(s *app.Settings, token string) error
		want    []string
		wantErr error
	}{
		{
			name: "the owner sets the number format", token: ownerToken, want: []string{"number format comma-dot"},
			change: func(s *app.Settings, token string) error { return s.SetNumberFormat(t.Context(), "o", token, comma) },
		},
		{
			name: "the owner sets the marginal rate, revaluing every account", token: ownerToken, want: []string{"marginal rate 0.12"},
			change: func(s *app.Settings, token string) error {
				return s.SetMarginalRate(t.Context(), "o", token, dec(t, "0.12"))
			},
		},
		{
			name: "a read session changes nothing", token: "read-session", wantErr: app.ErrNotOwner,
			change: func(s *app.Settings, token string) error {
				return s.SetMarginalRate(t.Context(), "o", token, dec(t, "0.12"))
			},
		},
		{
			name: "a rate of 0 is refused", token: ownerToken, wantErr: app.ErrInvalidInput,
			change: func(s *app.Settings, token string) error {
				return s.SetMarginalRate(t.Context(), "o", token, kernel.Decimal{})
			},
		},
		{
			name: "a format without a decimal separator is refused", token: ownerToken, wantErr: app.ErrInvalidInput,
			change: func(s *app.Settings, token string) error {
				return s.SetNumberFormat(t.Context(), "o", token, jurisdiction.NumberFormat{Group: jurisdiction.SeparatorSpace})
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rec := &fakeRecorder{}
			err := tt.change(newSettings(t, newFakeStore(), rec), tt.token)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("error (-want +got):\n%s", diff)
			}
			got := make([]string, 0, len(rec.calls))
			for _, c := range rec.calls {
				got = append(got, c.what)
			}
			if diff := cmp.Diff(tt.want, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("stored (-want +got):\n%s", diff)
			}
		})
	}
}
