package main

import "testing"

// TestClientEventsAreOnUnlessTheConfigTurnsThemOff checks D-227's default: the
// pages' reports are on with no config, and only an explicit false turns them off.
func TestClientEventsAreOnUnlessTheConfigTurnsThemOff(t *testing.T) {
	t.Parallel()
	on, off := true, false
	tests := []struct {
		name    string
		raw     rawConfig
		wantOff bool
	}{
		{"no config", rawConfig{}, false},
		{"set on", rawConfig{ClientEvents: &on}, false},
		{"set off", rawConfig{ClientEvents: &off}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.raw.clientEventsOff(); got != tt.wantOff {
				t.Errorf("clientEventsOff = %v, want %v", got, tt.wantOff)
			}
		})
	}
}
