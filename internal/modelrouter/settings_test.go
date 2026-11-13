package modelrouter_test

import (
	"testing"

	"github.com/arkadiuss/openfief/internal/modelrouter"
)

func TestSettingsHasModel(t *testing.T) {
	t.Parallel()
	route := map[modelrouter.Role]modelrouter.Route{
		modelrouter.RoleAsk: {Role: modelrouter.RoleAsk, Models: []modelrouter.Ref{{Instance: "p", Model: "m"}}},
	}
	on := func(fs ...modelrouter.Feature) modelrouter.Switch {
		sw := modelrouter.Switch{Enabled: true, Toggles: map[modelrouter.Feature]bool{}}
		for _, f := range fs {
			sw.Toggles[f] = true
		}
		return sw
	}
	tests := []struct {
		name string
		s    modelrouter.Settings
		role modelrouter.Role
		want bool
	}{
		{"nothing saved", modelrouter.Settings{}, modelrouter.RoleAsk, false},
		{"switch on, toggle on, a route", modelrouter.Settings{Switch: on(modelrouter.FeatureAsk), Routes: route}, modelrouter.RoleAsk, true},
		{"the switch off with everything else set", modelrouter.Settings{
			Switch: modelrouter.Switch{Toggles: map[modelrouter.Feature]bool{modelrouter.FeatureAsk: true}}, Routes: route,
		}, modelrouter.RoleAsk, false},
		{"the toggle off", modelrouter.Settings{Switch: on(), Routes: route}, modelrouter.RoleAsk, false},
		{"another feature's toggle", modelrouter.Settings{Switch: on(modelrouter.FeatureMaintainer), Routes: route}, modelrouter.RoleAsk, false},
		{"no route", modelrouter.Settings{Switch: on(modelrouter.FeatureAsk)}, modelrouter.RoleAsk, false},
		{"an empty route", modelrouter.Settings{
			Switch: on(modelrouter.FeatureAsk), Routes: map[modelrouter.Role]modelrouter.Route{modelrouter.RoleAsk: {Role: modelrouter.RoleAsk}},
		}, modelrouter.RoleAsk, false},
		{"another role's route", modelrouter.Settings{Switch: on(modelrouter.FeatureAsk, modelrouter.FeatureMaintainer), Routes: route}, modelrouter.RoleMaintainer, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.s.HasModel(tt.role); got != tt.want {
				t.Errorf("HasModel(%s) = %v, want %v", tt.role, got, tt.want)
			}
		})
	}
}
