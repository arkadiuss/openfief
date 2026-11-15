package main

import (
	"errors"
	"testing"
	"time"

	"github.com/arkadiuss/openfief/internal/views/ui"
)

func TestAskLimitsFromConfig(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		raw     rawConfig
		want    ui.AskLimits
		wantErr bool
	}{
		{name: "no config leaves every limit to its default", raw: rawConfig{}, want: ui.AskLimits{}},
		{
			name: "every limit set", raw: rawConfig{AskToolRounds: 3, AskToolCalls: 6, AskOutputTokens: 500, AskDeadline: "90s"},
			want: ui.AskLimits{ToolRounds: 3, ToolCalls: 6, OutputTokens: 500, Deadline: 90 * time.Second},
		},
		{name: "one limit set", raw: rawConfig{AskOutputTokens: 800}, want: ui.AskLimits{OutputTokens: 800}},
		{
			name: "the most", raw: rawConfig{AskToolRounds: maxToolRounds, AskToolCalls: maxToolCalls, AskOutputTokens: maxOutputTokens, AskDeadline: "5m"},
			want: ui.AskLimits{ToolRounds: maxToolRounds, ToolCalls: maxToolCalls, OutputTokens: maxOutputTokens, Deadline: maxDeadline},
		},
		{name: "rounds below zero", raw: rawConfig{AskToolRounds: -1}, wantErr: true},
		{name: "rounds above the most", raw: rawConfig{AskToolRounds: maxToolRounds + 1}, wantErr: true},
		{name: "calls above the most", raw: rawConfig{AskToolCalls: maxToolCalls + 1}, wantErr: true},
		{name: "tokens below zero", raw: rawConfig{AskOutputTokens: -5}, wantErr: true},
		{name: "tokens above the most", raw: rawConfig{AskOutputTokens: maxOutputTokens + 1}, wantErr: true},
		{name: "a deadline that is no duration", raw: rawConfig{AskDeadline: "soon"}, wantErr: true},
		{name: "a deadline too short", raw: rawConfig{AskDeadline: "1s"}, wantErr: true},
		{name: "a deadline too long", raw: rawConfig{AskDeadline: "6m"}, wantErr: true},
		{name: "a negative deadline", raw: rawConfig{AskDeadline: "-2m"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := tt.raw.askLimits()
			if (err != nil) != tt.wantErr || (err != nil && !errors.Is(err, errBadConfig)) {
				t.Fatalf("error = %v, want error: %v", err, tt.wantErr)
			}
			if err == nil && got != tt.want {
				t.Errorf("limits = %+v, want %+v", got, tt.want)
			}
		})
	}
}

// TestAskLimitsDefaultsAreD208s checks that a config that sets nothing leaves
// the box with D-208's limits: five rounds that may call tools, ten tool calls,
// 800 output tokens and two minutes.
func TestAskLimitsDefaultsAreD208s(t *testing.T) {
	t.Parallel()
	got := ui.DefaultAskLimits()
	want := ui.AskLimits{ToolRounds: 5, ToolCalls: 10, OutputTokens: 800, Deadline: 2 * time.Minute}
	if got != want {
		t.Errorf("DefaultAskLimits = %+v, want D-208's %+v", got, want)
	}
}
