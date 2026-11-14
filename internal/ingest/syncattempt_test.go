package ingest_test

import (
	"errors"
	"testing"
	"time"

	"github.com/arkadiuss/openfief/internal/ingest"
)

func TestSyncErrorClassNames(t *testing.T) {
	t.Parallel()
	tests := []struct {
		class ingest.SyncErrorClass
		name  string
	}{
		{0, ""},
		{ingest.SyncErrorTimeout, "timeout"},
		{ingest.SyncErrorUnavailable, "unavailable"},
		{ingest.SyncErrorConnector, "connector-error"},
		{ingest.SyncErrorInvalidDelivery, "invalid-delivery"},
		{ingest.SyncErrorRefused, "refused"},
		{ingest.SyncErrorBooking, "booking"},
		{ingest.SyncErrorInterrupted, "interrupted"},
		{ingest.SyncErrorOther, "other"},
		{ingest.SyncErrorDeliveryTimeout, "delivery-timeout"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.class.String(); got != tt.name {
				t.Errorf("String = %q, want %q", got, tt.name)
			}
			got, err := ingest.ParseSyncErrorClass(tt.name)
			if err != nil || got != tt.class {
				t.Errorf("ParseSyncErrorClass(%q) = %v, %v; want %v", tt.name, got, err, tt.class)
			}
		})
	}
	if _, err := ingest.ParseSyncErrorClass("network"); !errors.Is(err, ingest.ErrUnknownSyncErrorClass) {
		t.Errorf("ParseSyncErrorClass(network) = %v, want ErrUnknownSyncErrorClass", err)
	}
	if got := ingest.SyncErrorClass(42).String(); got != "SyncErrorClass(42)" {
		t.Errorf("an invalid class = %q", got)
	}
	failedAttempt, okAttempt := ingest.SyncAttempt{Class: ingest.SyncErrorOther, Failures: 1}, ingest.SyncAttempt{}
	if !failedAttempt.Failed() || okAttempt.Failed() {
		t.Error("Failed should follow the class")
	}
}

func TestKeepsFailing(t *testing.T) {
	t.Parallel()
	since := time.Date(2026, time.November, 3, 6, 0, 0, 0, time.UTC)
	tests := []struct {
		name   string
		status ingest.SyncStatus
		want   bool
	}{
		{name: "never synced", status: ingest.SyncStatus{}},
		{name: "synced", status: ingest.SyncStatus{LastAttempt: since, LastSuccess: since}},
		{
			name:   "two failures over an hour",
			status: ingest.SyncStatus{Failures: 2, FailingSince: since, LastAttempt: since.Add(time.Hour)},
		},
		{
			name:   "three failures within a minute",
			status: ingest.SyncStatus{Failures: 3, FailingSince: since, LastAttempt: since.Add(time.Minute)},
		},
		{
			name:   "three failures just short of the span",
			status: ingest.SyncStatus{Failures: 3, FailingSince: since, LastAttempt: since.Add(ingest.KeepsFailingFor - time.Second)},
		},
		{
			name:   "three failures over the span",
			status: ingest.SyncStatus{Failures: 3, FailingSince: since, LastAttempt: since.Add(ingest.KeepsFailingFor)},
			want:   true,
		},
		{
			name:   "many failures over hours",
			status: ingest.SyncStatus{Failures: 9, FailingSince: since, LastAttempt: since.Add(6 * time.Hour), LastClass: ingest.SyncErrorTimeout},
			want:   true,
		},
		{
			name:   "failures with no start recorded",
			status: ingest.SyncStatus{Failures: 5, LastAttempt: since},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.status.KeepsFailing(); got != tt.want {
				t.Errorf("KeepsFailing = %v, want %v", got, tt.want)
			}
		})
	}
}
