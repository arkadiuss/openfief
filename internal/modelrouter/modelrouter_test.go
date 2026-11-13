package modelrouter_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/modelrouter"
)

func ref(t *testing.T, s string) modelrouter.Ref {
	t.Helper()
	r, err := modelrouter.ParseRef(s)
	if err != nil {
		t.Fatalf("ParseRef(%q): %v", s, err)
	}
	return r
}

func TestSwitchAllows(t *testing.T) {
	t.Parallel()
	on := func(fs ...modelrouter.Feature) map[modelrouter.Feature]bool {
		m := map[modelrouter.Feature]bool{}
		for _, f := range fs {
			m[f] = true
		}
		return m
	}
	tests := []struct {
		name string
		sw   modelrouter.Switch
		f    modelrouter.Feature
		want bool
	}{
		{"the zero value is off", modelrouter.Switch{}, modelrouter.FeatureAsk, false},
		{"switch off, toggle on", modelrouter.Switch{Toggles: on(modelrouter.FeatureAsk)}, modelrouter.FeatureAsk, false},
		{"switch on, toggle off", modelrouter.Switch{Enabled: true}, modelrouter.FeatureAsk, false},
		{"switch on, toggle on", modelrouter.Switch{Enabled: true, Toggles: on(modelrouter.FeatureAsk)}, modelrouter.FeatureAsk, true},
		{"another feature's toggle", modelrouter.Switch{Enabled: true, Toggles: on(modelrouter.FeatureMCP)}, modelrouter.FeatureAsk, false},
		{"mcp", modelrouter.Switch{Enabled: true, Toggles: on(modelrouter.FeatureMCP)}, modelrouter.FeatureMCP, true},
		{"an unknown feature is never allowed", modelrouter.Switch{Enabled: true, Toggles: on("other")}, "other", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.sw.Allows(tt.f); got != tt.want {
				t.Errorf("Allows(%q) = %v, want %v", tt.f, got, tt.want)
			}
		})
	}
}

func TestSwitchChanges(t *testing.T) {
	t.Parallel()
	var sw modelrouter.Switch
	sw2, err := sw.WithToggle(modelrouter.FeatureMaintainer, true)
	if err != nil {
		t.Fatal(err)
	}
	if sw.Toggles != nil || sw.Enabled {
		t.Error("WithToggle changed its receiver")
	}
	sw3 := sw2.WithEnabled(true)
	if sw2.Enabled || !sw3.Allows(modelrouter.FeatureMaintainer) {
		t.Errorf("WithEnabled: %+v then %+v", sw2, sw3)
	}
	sw4, err := sw3.WithToggle(modelrouter.FeatureMaintainer, false)
	if err != nil || sw4.Allows(modelrouter.FeatureMaintainer) || !sw3.Allows(modelrouter.FeatureMaintainer) {
		t.Errorf("turning a toggle off: %+v, %v", sw4, err)
	}
	if _, err := sw.WithToggle("other", true); !errors.Is(err, modelrouter.ErrUnknownFeature) {
		t.Errorf("an unknown toggle: %v, want ErrUnknownFeature", err)
	}
}

func TestParseRef(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in      string
		want    modelrouter.Ref
		wantErr bool
	}{
		{"openrouter/openai/gpt-oss-120b", modelrouter.Ref{Provider: "openrouter", Model: "openai/gpt-oss-120b"}, false},
		{"ollama/qwen3:8b", modelrouter.Ref{Provider: "ollama", Model: "qwen3:8b"}, false},
		{"local-1/m", modelrouter.Ref{Provider: "local-1", Model: "m"}, false},
		{"", modelrouter.Ref{}, true},
		{"noslash", modelrouter.Ref{}, true},
		{"/model", modelrouter.Ref{}, true},
		{"provider/", modelrouter.Ref{}, true},
		{"Upper/model", modelrouter.Ref{}, true},
		{"a--b/model", modelrouter.Ref{}, true},
		{"-a/model", modelrouter.Ref{}, true},
		{"a-/model", modelrouter.Ref{}, true},
		{"provider/has space", modelrouter.Ref{}, true},
		{"provider/new\nline", modelrouter.Ref{}, true},
		{"provider/<script>", modelrouter.Ref{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()
			got, err := modelrouter.ParseRef(tt.in)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseRef(%q) error = %v, wantErr %v", tt.in, err, tt.wantErr)
			}
			if err != nil && !errors.Is(err, modelrouter.ErrInvalidRef) {
				t.Errorf("error %v does not wrap ErrInvalidRef", err)
			}
			if got != tt.want {
				t.Errorf("ParseRef(%q) = %v, want %v", tt.in, got, tt.want)
			}
			if err == nil && got.String() != tt.in {
				t.Errorf("String() = %q, want %q", got.String(), tt.in)
			}
		})
	}
}

func TestNewRoute(t *testing.T) {
	t.Parallel()
	a, b := ref(t, "p/a"), ref(t, "p/b")
	many := make([]modelrouter.Ref, 0, modelrouter.MaxRouteModels+1)
	for i := range modelrouter.MaxRouteModels + 1 {
		many = append(many, modelrouter.Ref{Provider: "p", Model: string(rune('a' + i))})
	}
	tests := []struct {
		name    string
		role    modelrouter.Role
		refs    []modelrouter.Ref
		wantErr error
	}{
		{"primary and fallback", modelrouter.RoleAsk, []modelrouter.Ref{a, b}, nil},
		{"primary only", modelrouter.RoleMaintainer, []modelrouter.Ref{a}, nil},
		{"unknown role", "other", []modelrouter.Ref{a}, modelrouter.ErrUnknownRole},
		{"no models", modelrouter.RoleAsk, nil, modelrouter.ErrInvalidRoute},
		{"one twice", modelrouter.RoleAsk, []modelrouter.Ref{a, b, a}, modelrouter.ErrInvalidRoute},
		{"too many", modelrouter.RoleAsk, many, modelrouter.ErrInvalidRoute},
		{"an invalid reference", modelrouter.RoleAsk, []modelrouter.Ref{{Provider: "P", Model: "m"}}, modelrouter.ErrInvalidRef},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := modelrouter.NewRoute(tt.role, tt.refs)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("NewRoute error = %v, want %v", err, tt.wantErr)
			}
			if err == nil && !cmp.Equal(got.Models, tt.refs) {
				t.Errorf("models = %v, want %v", got.Models, tt.refs)
			}
		})
	}
}

func TestRoleFeature(t *testing.T) {
	t.Parallel()
	if got := modelrouter.RoleAsk.Feature(); got != modelrouter.FeatureAsk {
		t.Errorf("ask role's feature = %q", got)
	}
	if got := modelrouter.RoleMaintainer.Feature(); got != modelrouter.FeatureMaintainer {
		t.Errorf("maintainer role's feature = %q", got)
	}
}

func TestNewTierSet(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		tags    []modelrouter.Tier
		want    modelrouter.TierSet
		wantErr error
	}{
		{"one", []modelrouter.Tier{modelrouter.TierT2}, modelrouter.TierSet{modelrouter.TierT2}, nil},
		{"distinct, in tier order", []modelrouter.Tier{modelrouter.TierT2, modelrouter.TierT0, modelrouter.TierT2, modelrouter.TierTF}, modelrouter.TierSet{modelrouter.TierT0, modelrouter.TierTF, modelrouter.TierT2}, nil},
		{"none", nil, nil, modelrouter.ErrUntagged},
		{"an untagged piece among tagged", []modelrouter.Tier{modelrouter.TierT2, ""}, nil, modelrouter.ErrUntagged},
		{"an unknown tier", []modelrouter.Tier{modelrouter.TierT2, "T9"}, nil, modelrouter.ErrUnknownTier},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := modelrouter.NewTierSet(tt.tags)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("set (-want +got):\n%s", diff)
			}
		})
	}
}

func TestTierPolicyDenied(t *testing.T) {
	t.Parallel()
	cloud, local := ref(t, "openrouter/qwen"), ref(t, "ollama/qwen3:8b")
	policy, err := modelrouter.NewTierPolicy(map[modelrouter.Tier][]modelrouter.Ref{
		modelrouter.TierT0: {local},
		modelrouter.TierTF: {local},
		modelrouter.TierT1: {local, cloud},
		modelrouter.TierT2: {local, cloud},
	})
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name  string
		model modelrouter.Ref
		tiers modelrouter.TierSet
		want  []modelrouter.Tier
	}{
		{"local model, everything", local, modelrouter.Tiers(), nil},
		{"cloud model, public data", cloud, modelrouter.TierSet{modelrouter.TierT2}, nil},
		{"cloud model, thesis and public data", cloud, modelrouter.TierSet{modelrouter.TierT1, modelrouter.TierT2}, nil},
		{"cloud model, portfolio data", cloud, modelrouter.TierSet{modelrouter.TierT0}, []modelrouter.Tier{modelrouter.TierT0}},
		{"cloud model, one denied tier among allowed ones", cloud, modelrouter.TierSet{modelrouter.TierT0, modelrouter.TierT1, modelrouter.TierT2}, []modelrouter.Tier{modelrouter.TierT0}},
		{"cloud model, two denied tiers", cloud, modelrouter.TierSet{modelrouter.TierT0, modelrouter.TierTF}, []modelrouter.Tier{modelrouter.TierT0, modelrouter.TierTF}},
		{"an unlisted model is denied everything", ref(t, "other/m"), modelrouter.TierSet{modelrouter.TierT2}, []modelrouter.Tier{modelrouter.TierT2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tt.want, policy.Denied(tt.model, tt.tiers)); diff != "" {
				t.Errorf("denied (-want +got):\n%s", diff)
			}
		})
	}
	t.Run("the zero policy denies by default", func(t *testing.T) {
		t.Parallel()
		var zero modelrouter.TierPolicy
		for _, tier := range modelrouter.Tiers() {
			if zero.Allows(tier, local) {
				t.Errorf("the zero policy allows %s", tier)
			}
		}
	})
}

func TestNewTierPolicyErrors(t *testing.T) {
	t.Parallel()
	a := ref(t, "p/a")
	tests := []struct {
		name    string
		lists   map[modelrouter.Tier][]modelrouter.Ref
		wantErr error
	}{
		{"unknown tier", map[modelrouter.Tier][]modelrouter.Ref{"T9": {a}}, modelrouter.ErrUnknownTier},
		{"one twice", map[modelrouter.Tier][]modelrouter.Ref{modelrouter.TierT2: {a, a}}, modelrouter.ErrInvalidTierModels},
		{"an invalid reference", map[modelrouter.Tier][]modelrouter.Ref{modelrouter.TierT2: {{Provider: "p"}}}, modelrouter.ErrInvalidRef},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if _, err := modelrouter.NewTierPolicy(tt.lists); !errors.Is(err, tt.wantErr) {
				t.Errorf("error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestMakePlan(t *testing.T) {
	t.Parallel()
	cloud, local, other := ref(t, "openrouter/qwen"), ref(t, "ollama/qwen3:8b"), ref(t, "vllm/big")
	on := modelrouter.Switch{Enabled: true, Toggles: map[modelrouter.Feature]bool{modelrouter.FeatureAsk: true}}
	route, err := modelrouter.NewRoute(modelrouter.RoleAsk, []modelrouter.Ref{cloud, local, other})
	if err != nil {
		t.Fatal(err)
	}
	policy, err := modelrouter.NewTierPolicy(map[modelrouter.Tier][]modelrouter.Ref{
		modelrouter.TierT0: {local, other},
		modelrouter.TierT2: {cloud, local, other},
	})
	if err != nil {
		t.Fatal(err)
	}
	catalog := map[modelrouter.Ref]modelrouter.Entry{
		local: {ID: "qwen3:8b", ToolCalling: true},
		other: {ID: "big", ToolCalling: true, StructuredOut: true},
	}
	tests := []struct {
		name      string
		mod       func(in *modelrouter.PlanInput)
		want      []modelrouter.Ref
		wantSkips []modelrouter.Skip
		wantErr   error
	}{
		{
			name: "public data goes to the whole route in order",
			mod:  func(in *modelrouter.PlanInput) { in.Tiers = modelrouter.TierSet{modelrouter.TierT2} },
			want: []modelrouter.Ref{cloud, local, other},
		},
		{
			name: "portfolio data skips the cloud model",
			mod: func(in *modelrouter.PlanInput) {
				in.Tiers = modelrouter.TierSet{modelrouter.TierT0, modelrouter.TierT2}
			},
			want:      []modelrouter.Ref{local, other},
			wantSkips: []modelrouter.Skip{{Model: cloud, Reason: modelrouter.SkipTierDenied, Tiers: []modelrouter.Tier{modelrouter.TierT0}}},
		},
		{
			name: "structured output skips a model whose entry lacks it",
			mod: func(in *modelrouter.PlanInput) {
				in.Tiers = modelrouter.TierSet{modelrouter.TierT2}
				in.Needs = modelrouter.Needs{StructuredOutput: true}
			},
			want:      []modelrouter.Ref{cloud, other},
			wantSkips: []modelrouter.Skip{{Model: local, Reason: modelrouter.SkipIncapable}},
		},
		{
			name: "a model the catalog does not list stays a candidate",
			mod: func(in *modelrouter.PlanInput) {
				in.Tiers = modelrouter.TierSet{modelrouter.TierT2}
				in.Needs = modelrouter.Needs{Tools: true}
			},
			want: []modelrouter.Ref{cloud, local, other},
		},
		{
			name: "the switch off refuses everything",
			mod: func(in *modelrouter.PlanInput) {
				in.Switch.Enabled = false
				in.Tiers = modelrouter.TierSet{modelrouter.TierT2}
			},
			wantErr: modelrouter.ErrSwitchOff,
		},
		{
			name: "the role's toggle off refuses",
			mod: func(in *modelrouter.PlanInput) {
				in.Switch.Toggles = map[modelrouter.Feature]bool{modelrouter.FeatureMCP: true}
				in.Tiers = modelrouter.TierSet{modelrouter.TierT2}
			},
			wantErr: modelrouter.ErrFeatureOff,
		},
		{
			name: "another role's toggle does not enable the role",
			mod: func(in *modelrouter.PlanInput) {
				in.Role = modelrouter.RoleMaintainer
				in.Tiers = modelrouter.TierSet{modelrouter.TierT2}
			},
			wantErr: modelrouter.ErrFeatureOff,
		},
		{
			name: "no route",
			mod: func(in *modelrouter.PlanInput) {
				in.Route = modelrouter.Route{}
				in.Tiers = modelrouter.TierSet{modelrouter.TierT2}
			},
			wantErr: modelrouter.ErrNoRoute,
		},
		{
			name:    "no tiers: untagged data is not sent",
			mod:     func(in *modelrouter.PlanInput) { in.Tiers = nil },
			wantErr: modelrouter.ErrUntagged,
		},
		{
			name:    "no model of the route is allowed for a tier",
			mod:     func(in *modelrouter.PlanInput) { in.Tiers = modelrouter.TierSet{modelrouter.TierTF} },
			wantErr: modelrouter.ErrTierDenied,
		},
		{
			name: "no model supports tools and structured output",
			mod: func(in *modelrouter.PlanInput) {
				in.Route = modelrouter.Route{Role: modelrouter.RoleAsk, Models: []modelrouter.Ref{local}}
				in.Tiers = modelrouter.TierSet{modelrouter.TierT2}
				in.Needs = modelrouter.Needs{Tools: true, StructuredOutput: true}
			},
			wantErr: modelrouter.ErrNoCapableModel,
		},
		{
			name:    "an empty policy denies by default",
			mod:     func(in *modelrouter.PlanInput) { in.Policy = nil; in.Tiers = modelrouter.TierSet{modelrouter.TierT2} },
			wantErr: modelrouter.ErrTierDenied,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			in := modelrouter.PlanInput{Switch: on, Role: modelrouter.RoleAsk, Route: route, Policy: policy, Catalog: catalog}
			tt.mod(&in)
			plan, err := modelrouter.MakePlan(&in)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				return
			}
			if diff := cmp.Diff(tt.want, plan.Candidates); diff != "" {
				t.Errorf("candidates (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.wantSkips, plan.Skipped); diff != "" {
				t.Errorf("skipped (-want +got):\n%s", diff)
			}
		})
	}
}

func TestRequest(t *testing.T) {
	t.Parallel()
	user := modelrouter.Message{Role: modelrouter.MessageUser, Content: "q", Tier: modelrouter.TierT0}
	tests := []struct {
		name      string
		req       modelrouter.Request
		wantValid error
		wantTiers modelrouter.TierSet
		wantErr   error
	}{
		{"one tagged message", modelrouter.Request{Messages: []modelrouter.Message{user}}, nil, modelrouter.TierSet{modelrouter.TierT0}, nil},
		{
			"system, user and a tool result",
			modelrouter.Request{Messages: []modelrouter.Message{
				{Role: modelrouter.MessageSystem, Tier: modelrouter.TierT2},
				user,
				{Role: modelrouter.MessageAssistant, Tier: modelrouter.TierT0, ToolCalls: []modelrouter.ToolCall{{ID: "1", Name: "x"}}},
				{Role: modelrouter.MessageTool, ToolCallID: "1", Tier: modelrouter.TierTF},
			}},
			nil,
			modelrouter.TierSet{modelrouter.TierT0, modelrouter.TierTF, modelrouter.TierT2},
			nil,
		},
		{"no messages", modelrouter.Request{}, modelrouter.ErrInvalidRequest, nil, modelrouter.ErrUntagged},
		{"ends with the assistant", modelrouter.Request{Messages: []modelrouter.Message{user, {Role: modelrouter.MessageAssistant, Tier: modelrouter.TierT0}}}, modelrouter.ErrInvalidRequest, modelrouter.TierSet{modelrouter.TierT0}, nil},
		{"unknown role", modelrouter.Request{Messages: []modelrouter.Message{{Role: "other", Tier: modelrouter.TierT2}}}, modelrouter.ErrInvalidRequest, modelrouter.TierSet{modelrouter.TierT2}, nil},
		{"unnamed tool", modelrouter.Request{Messages: []modelrouter.Message{user}, Tools: []modelrouter.Tool{{}}}, modelrouter.ErrInvalidRequest, modelrouter.TierSet{modelrouter.TierT0}, nil},
		{"negative limit", modelrouter.Request{Messages: []modelrouter.Message{user}, MaxOutputTokens: -1}, modelrouter.ErrInvalidRequest, modelrouter.TierSet{modelrouter.TierT0}, nil},
		{"an untagged message", modelrouter.Request{Messages: []modelrouter.Message{user, {Role: modelrouter.MessageUser}}}, nil, nil, modelrouter.ErrUntagged},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := tt.req.Validate(); !errors.Is(err, tt.wantValid) {
				t.Errorf("Validate error = %v, want %v", err, tt.wantValid)
			}
			got, err := tt.req.Tiers()
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Tiers error = %v, want %v", err, tt.wantErr)
			}
			if diff := cmp.Diff(tt.wantTiers, got); diff != "" {
				t.Errorf("tiers (-want +got):\n%s", diff)
			}
		})
	}
}

func TestErrorsNeverQuoteTheRequest(t *testing.T) {
	t.Parallel()
	const secret = "ACCOUNT-1234-SECRET"
	req := modelrouter.Request{Messages: []modelrouter.Message{{Role: modelrouter.MessageUser, Content: secret}}}
	_, tiersErr := req.Tiers()
	for _, err := range []error{req.Validate(), tiersErr} {
		if err != nil && strings.Contains(err.Error(), secret) {
			t.Errorf("an error quotes the request text: %v", err)
		}
	}
	in := modelrouter.PlanInput{Role: modelrouter.RoleAsk, Tiers: modelrouter.TierSet{modelrouter.TierT0}}
	if _, err := modelrouter.MakePlan(&in); err == nil || strings.Contains(err.Error(), secret) {
		t.Errorf("MakePlan error = %v", err)
	}
}

func TestEntryServes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		e     modelrouter.Entry
		needs modelrouter.Needs
		want  bool
	}{
		{"asks nothing", modelrouter.Entry{}, modelrouter.Needs{}, true},
		{"tools, entry has them", modelrouter.Entry{ToolCalling: true}, modelrouter.Needs{Tools: true}, true},
		{"tools, entry lacks them", modelrouter.Entry{}, modelrouter.Needs{Tools: true}, false},
		{"structured output, entry lacks it", modelrouter.Entry{ToolCalling: true}, modelrouter.Needs{StructuredOutput: true}, false},
		{"both", modelrouter.Entry{ToolCalling: true, StructuredOut: true}, modelrouter.Needs{Tools: true, StructuredOutput: true}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.e.Serves(tt.needs); got != tt.want {
				t.Errorf("Serves = %v, want %v", got, tt.want)
			}
		})
	}
}
