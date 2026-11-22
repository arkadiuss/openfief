package main

import (
	"errors"
	"testing"
	"time"

	"github.com/arkadiuss/openfief/internal/modelrouter"
	modelsapp "github.com/arkadiuss/openfief/internal/modelrouter/app"
	"github.com/arkadiuss/openfief/internal/views/browserproxy"
)

// TestGuardFlags checks D-208's flags: off by default for an installation that
// sets none, and the demo's values when it does.
func TestGuardFlags(t *testing.T) {
	t.Parallel()
	demo := []string{
		"--model-daily-budget-usd=0.50", "--ask-questions-per-hour=10", "--ask-questions-per-day=30", "--ask-concurrent-calls=3",
		"--ask-queue-wait=2s", "--client-ip-header=X-Forwarded-For", "--trusted-proxies=10.42.0.0/16", "--allow-turnstile",
	}
	tests := []struct {
		name    string
		args    []string
		want    guardConfig
		wantErr error
		fail    bool
	}{
		{name: "none set: nothing is limited", want: guardConfig{QueueWait: modelsapp.DefaultQueueWait}},
		{
			name: "the demo's values", args: demo,
			want: guardConfig{
				DailyBudgetUSD: "0.50", QuestionsPerHour: 10, QuestionsPerDay: 30, AskCalls: 3, QueueWait: 2 * time.Second,
				ClientIPHeader: "X-Forwarded-For", TrustedProxies: "10.42.0.0/16", AllowTurnstile: true,
			},
		},
		{name: "a budget that is not a number", args: []string{"--model-daily-budget-usd=lots"}, wantErr: errGuards},
		{name: "a negative budget", args: []string{"--model-daily-budget-usd=-1"}, wantErr: errGuards},
		{name: "a negative limit", args: []string{"--ask-questions-per-hour=-1"}, wantErr: errGuards},
		{name: "more an hour than a day", args: []string{"--ask-questions-per-hour=31", "--ask-questions-per-day=30"}, wantErr: errGuards},
		{name: "negative calls at once", args: []string{"--ask-concurrent-calls=-1"}, wantErr: errGuards},
		{name: "a negative queue wait", args: []string{"--ask-queue-wait=-1s"}, wantErr: errGuards},
		{name: "a header with no proxy to trust", args: []string{"--client-ip-header=X-Forwarded-For"}, wantErr: browserproxy.ErrClientIPConfig},
		{name: "proxies with no header", args: []string{"--trusted-proxies=10.0.0.0/8"}, wantErr: browserproxy.ErrClientIPConfig},
		{name: "a proxy that is no network", args: []string{"--client-ip-header=X-Forwarded-For", "--trusted-proxies=ingress"}, wantErr: browserproxy.ErrClientIPConfig},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := parseServe(append([]string{"--authz=policy"}, tt.args...), func(string) string { return "" })
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && got.Guards != tt.want {
				t.Errorf("guards = %+v, want %+v", got.Guards, tt.want)
			}
		})
	}
}

func TestGuardConfigBuildsTheRoutersGuardsAndTheAsksLimits(t *testing.T) {
	t.Parallel()
	g := guardConfig{DailyBudgetUSD: "0.5", QuestionsPerHour: 10, QuestionsPerDay: 30, AskCalls: 3, QueueWait: time.Second}
	b, err := g.budget()
	if err != nil {
		t.Fatal(err)
	}
	if !b.Enabled() || b.Currency.String() != "USD" || b.DailyCap.String() != "0.5" {
		t.Errorf("budget = %+v, want USD 0.5", b)
	}
	if got, want := g.limits(), (modelrouter.QuestionLimits{PerHour: 10, PerDay: 30}); got != want {
		t.Errorf("limits = %+v, want %+v", got, want)
	}
	none, err := (&guardConfig{}).budget()
	if err != nil || none.Enabled() {
		t.Errorf("no flag: budget = %+v, %v, want none", none, err)
	}
	zero, err := (&guardConfig{DailyBudgetUSD: "0"}).budget()
	if err != nil || zero.Enabled() {
		t.Errorf("a zero budget = %+v, %v, want none", zero, err)
	}
}

func TestGuardConfigProxyOptions(t *testing.T) {
	t.Parallel()
	for _, g := range []guardConfig{{}, {ClientIPHeader: "X-Forwarded-For", TrustedProxies: "10.42.0.0/16", AllowTurnstile: true}} {
		if _, err := g.proxyOptions(); err != nil {
			t.Errorf("proxyOptions(%+v): %v", g, err)
		}
	}
	bad := guardConfig{ClientIPHeader: "X-Forwarded-For"}
	if _, err := bad.proxyOptions(); !errors.Is(err, browserproxy.ErrClientIPConfig) {
		t.Errorf("a header with no proxy: error = %v", err)
	}
}
