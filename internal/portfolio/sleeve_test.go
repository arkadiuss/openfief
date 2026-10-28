package portfolio_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/portfolio"
)

func TestSleeveValidate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		sleeve  portfolio.Sleeve
		wantErr error
	}{
		{name: "defensive", sleeve: portfolio.Sleeve{ID: "core", Name: "Core", Role: portfolio.RoleDefensive}},
		{name: "offensive", sleeve: portfolio.Sleeve{ID: "theses", Name: "Theses", Role: portfolio.RoleOffensive}},
		{name: "no role", sleeve: portfolio.Sleeve{ID: "core", Name: "Core"}, wantErr: portfolio.ErrInvalidSleeve},
		{name: "unknown role", sleeve: portfolio.Sleeve{ID: "core", Name: "Core", Role: portfolio.Role(3)}, wantErr: portfolio.ErrInvalidSleeve},
		{name: "no ID", sleeve: portfolio.Sleeve{Name: "Core", Role: portfolio.RoleDefensive}, wantErr: portfolio.ErrInvalidSleeve},
		{name: "no name", sleeve: portfolio.Sleeve{ID: "core", Role: portfolio.RoleDefensive}, wantErr: portfolio.ErrInvalidSleeve},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.sleeve.Validate()
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Validate() error (-want +got):\n%s", diff)
			}
		})
	}
}

func TestParseRole(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		in      string
		want    portfolio.Role
		wantErr error
	}{
		{name: "defensive", in: "defensive", want: portfolio.RoleDefensive},
		{name: "offensive", in: "offensive", want: portfolio.RoleOffensive},
		{name: "capitalized", in: "Offensive", wantErr: portfolio.ErrInvalidSleeve},
		{name: "empty", in: "", wantErr: portfolio.ErrInvalidSleeve},
		{name: "placeholder of an invalid role", in: "Role(0)", wantErr: portfolio.ErrInvalidSleeve},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := portfolio.ParseRole(tt.in)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("ParseRole(%q) error (-want +got):\n%s", tt.in, diff)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ParseRole(%q) (-want +got):\n%s", tt.in, diff)
			}
			if err == nil && got.String() != tt.in {
				t.Errorf("String() = %q, want %q", got.String(), tt.in)
			}
		})
	}
}
