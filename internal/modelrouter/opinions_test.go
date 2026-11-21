package modelrouter_test

import (
	"errors"
	"testing"

	"github.com/arkadiuss/openfief/internal/modelrouter"
)

// errPolicy is the error of a route check that fails.
var errPolicy = errors.New("the policy could not be read")

func TestOpinionsTiersAreTheQuestionTheThesisAndThePrices(t *testing.T) {
	t.Parallel()
	got := modelrouter.OpinionsTiers()
	want := modelrouter.DataTierSet{modelrouter.DataTierT0, modelrouter.DataTierT1, modelrouter.DataTierT2}
	if len(got) != len(want) {
		t.Fatalf("OpinionsTiers() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("OpinionsTiers()[%d] = %s, want %s", i, got[i], want[i])
		}
	}
}

func TestOpinions(t *testing.T) {
	t.Parallel()
	route := map[modelrouter.Role]modelrouter.Route{
		modelrouter.RoleAsk: {Role: modelrouter.RoleAsk, Models: []modelrouter.Ref{{Instance: "p", Model: "m"}}},
	}
	on := modelrouter.Switch{Enabled: true, Toggles: map[modelrouter.Feature]bool{modelrouter.FeatureAsk: true}}
	ready := modelrouter.Settings{Switch: on, Routes: route, Opinions: true}
	eligible := func() (bool, error) { return true, nil }
	boom := errPolicy
	with := func(edit func(*modelrouter.OpinionsInput)) *modelrouter.OpinionsInput {
		in := &modelrouter.OpinionsInput{Settings: ready, Owner: true, RouteEligible: eligible}
		edit(in)
		return in
	}
	tests := []struct {
		name    string
		in      *modelrouter.OpinionsInput
		want    modelrouter.OpinionsStatus
		wantErr error
	}{
		{"everything on, the owner, an eligible route", with(func(*modelrouter.OpinionsInput) {}), modelrouter.OpinionsOn, nil},
		{"off by default", with(func(in *modelrouter.OpinionsInput) { in.Settings.Opinions = false }), modelrouter.OpinionsOff, nil},
		{"the zero settings", with(func(in *modelrouter.OpinionsInput) { in.Settings = modelrouter.Settings{} }), modelrouter.OpinionsOff, nil},
		{"the one switch off", with(func(in *modelrouter.OpinionsInput) { in.Settings.Switch.Enabled = false }), modelrouter.OpinionsOff, nil},
		{"the Ask box's toggle off", with(func(in *modelrouter.OpinionsInput) { in.Settings.Switch.Toggles = nil }), modelrouter.OpinionsOff, nil},
		{"no model on the route", with(func(in *modelrouter.OpinionsInput) { in.Settings.Routes = nil }), modelrouter.OpinionsOff, nil},
		{"a viewer asks", with(func(in *modelrouter.OpinionsInput) { in.Owner = false }), modelrouter.OpinionsOff, nil},
		{"a public demo, whatever is stored", with(func(in *modelrouter.OpinionsInput) { in.Demo = true }), modelrouter.OpinionsOff, nil},
		{"a route that may not receive the tiers", with(func(in *modelrouter.OpinionsInput) {
			in.RouteEligible = func() (bool, error) { return false, nil }
		}), modelrouter.OpinionsRouteIneligible, nil},
		{"the route can't be checked", with(func(in *modelrouter.OpinionsInput) {
			in.RouteEligible = func() (bool, error) { return false, boom }
		}), modelrouter.OpinionsOff, boom},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := modelrouter.Opinions(tt.in)
			if !errors.Is(err, tt.wantErr) || got != tt.want {
				t.Errorf("Opinions() = %q, %v; want %q, %v", got, err, tt.want, tt.wantErr)
			}
		})
	}
}

func TestOpinionsDoesNotAskTheRouteWhenAnythingElseKeepsThemOff(t *testing.T) {
	t.Parallel()
	asked := false
	in := &modelrouter.OpinionsInput{Owner: true, RouteEligible: func() (bool, error) { asked = true; return true, nil }}
	if got, err := modelrouter.Opinions(in); err != nil || got != modelrouter.OpinionsOff || asked {
		t.Errorf("Opinions() = %q, route asked = %v; want off without asking the route", got, asked)
	}
}
