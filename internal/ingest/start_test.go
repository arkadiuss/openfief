package ingest_test

import (
	"errors"
	"testing"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/kernel"
)

func TestStartMode(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		seeded   bool
		booked   bool
		override ingest.Mode
		want     ingest.Mode
		wantErr  error
	}{
		{name: "a new account starts from its institution's positions", want: ingest.ModeOpening},
		{name: "a new account may start from its whole history", override: ingest.ModeTransactions, want: ingest.ModeTransactions},
		{name: "a seeded account takes transactions", seeded: true, booked: true, want: ingest.ModeTransactions},
		{name: "a seeded account may be seeded again, which the ledger decides", seeded: true, booked: true, override: ingest.ModeOpening, want: ingest.ModeOpening},
		{name: "an account started from its whole history takes transactions", booked: true, want: ingest.ModeTransactions},
		{name: "an account started from its whole history keeps transactions", booked: true, override: ingest.ModeTransactions, want: ingest.ModeTransactions},
		{name: "an account started from its whole history is never seeded", booked: true, override: ingest.ModeOpening, wantErr: ingest.ErrStartedFromHistory},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := ingest.StartMode(tt.seeded, tt.booked, tt.override)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("StartMode() error = %v, want %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("StartMode(%v, %v, %s) = %s, want %s", tt.seeded, tt.booked, tt.override, got, tt.want)
			}
		})
	}
}

func TestCheckOpeningCash(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		cash    string
		mode    ingest.Mode
		seeded  bool
		wantErr error
	}{
		{name: "no cash entered, opening", mode: ingest.ModeOpening},
		{name: "no cash entered, transactions", mode: ingest.ModeTransactions},
		{name: "cash entered with an opening upload", cash: "1250.00", mode: ingest.ModeOpening},
		{name: "zero cash entered with an opening upload", cash: "0", mode: ingest.ModeOpening},
		{name: "cash entered with a transactions upload", cash: "1250.00", mode: ingest.ModeTransactions, wantErr: ingest.ErrOpeningCashOutsideOpening},
		{name: "zero cash entered with a transactions upload", cash: "0", mode: ingest.ModeTransactions, wantErr: ingest.ErrOpeningCashOutsideOpening},
		{name: "no cash entered when seeding a seeded account again", mode: ingest.ModeOpening, seeded: true},
		{name: "cash entered again for a seeded account", cash: "1250.00", mode: ingest.ModeOpening, seeded: true, wantErr: ingest.ErrOpeningCashEntered},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var cash *kernel.Money
			if tt.cash != "" {
				m := pln(t, tt.cash)
				cash = &m
			}
			if err := ingest.CheckOpeningCash(cash, tt.mode, tt.seeded); !errors.Is(err, tt.wantErr) {
				t.Errorf("CheckOpeningCash(%q, %s) = %v, want %v", tt.cash, tt.mode, err, tt.wantErr)
			}
		})
	}
}
