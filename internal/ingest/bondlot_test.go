package ingest_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
)

// bondLotCase is a lot in text form: a synthetic EDO lot unless a field
// says otherwise.
type bondLotCase struct {
	name      string
	series    string
	quantity  string
	purchased string
	first     string
	margin    string
	fee       string
	feeCcy    string
	purchase  ingest.BondPurchase
	wantErr   error
}

// lot builds the case's lot, filling the blank fields with a valid
// synthetic EDO0536 lot bought on 2026-05-04.
func (c *bondLotCase) lot(t *testing.T) ingest.BondLot {
	t.Helper()
	or := func(s, def string) string {
		if s == "" {
			return def
		}
		return s
	}
	series, err := pl.ParseSeries(or(c.series, "EDO0536"))
	if err != nil {
		t.Fatalf("ParseSeries: %v", err)
	}
	q, err := kernel.ParseQuantity(or(c.quantity, "40"))
	if err != nil {
		t.Fatalf("ParseQuantity: %v", err)
	}
	first, err1 := kernel.ParseDecimal(or(c.first, "0.06"))
	margin, err2 := kernel.ParseDecimal(or(c.margin, "0.02"))
	ccy, err3 := kernel.ParseCurrency(or(c.feeCcy, "PLN"))
	fee, err4 := kernel.ParseDecimal(or(c.fee, "2.00"))
	for _, e := range []error{err1, err2, err3, err4} {
		if e != nil {
			t.Fatalf("parsing the case: %v", e)
		}
	}
	l := ingest.BondLot{
		Terms:    pl.Terms{Series: series, FirstPeriodRate: first, Margin: margin, EarlyRedemptionFee: kernel.NewMoney(fee, ccy)},
		Quantity: q, Purchase: c.purchase,
	}
	if c.purchased != "none" {
		if l.Purchased, err = kernel.ParseDate(or(c.purchased, "2026-05-04")); err != nil {
			t.Fatalf("ParseDate: %v", err)
		}
	}
	return l
}

func TestBondLotValidate(t *testing.T) {
	t.Parallel()
	tests := []bondLotCase{
		{name: "EDO bought in its series' month ten years before maturity"},
		{name: "bought today", purchased: "2026-10-29", series: "EDO1036"},
		{name: "fixed-rate TOS with no margin", series: "TOS1029", purchased: "2026-10-01", margin: "0"},
		{name: "a fee of zero", fee: "0"},
		{name: "no purchase date", purchased: "none", wantErr: ingest.ErrInvalidBondLot},
		{name: "bought after today", series: "EDO1136", purchased: "2026-11-02", wantErr: ingest.ErrInvalidBondLot},
		{name: "purchase date of another series", purchased: "2026-06-01", wantErr: ingest.ErrInvalidBondLot},
		{name: "a fraction of a bond", quantity: "1.5", wantErr: ingest.ErrInvalidBondLot},
		{name: "no bonds", quantity: "0", wantErr: ingest.ErrInvalidBondLot},
		{name: "negative quantity", quantity: "-3", wantErr: ingest.ErrInvalidBondLot},
		{name: "negative first-period rate", first: "-0.01", wantErr: ingest.ErrInvalidBondLot},
		{name: "negative margin", margin: "-0.01", wantErr: ingest.ErrInvalidBondLot},
		{name: "negative fee", fee: "-1", wantErr: ingest.ErrInvalidBondLot},
		{name: "fee not in PLN", feeCcy: "EUR", wantErr: ingest.ErrInvalidBondLot},
		{name: "bought by exchange of maturing bonds", purchase: ingest.PurchaseExchange},
		{name: "a purchase neither cash nor exchange", purchase: ingest.BondPurchase(5), wantErr: ingest.ErrInvalidBondLot},
	}
	today, err := kernel.ParseDate("2026-10-29")
	if err != nil {
		t.Fatalf("ParseDate: %v", err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			l := tt.lot(t)
			if diff := cmp.Diff(tt.wantErr, l.Validate(today), cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Validate (-want +got):\n%s", diff)
			}
		})
	}
}

func TestBondLotCost(t *testing.T) {
	t.Parallel()
	c := bondLotCase{quantity: "40"}
	l := c.lot(t)
	want, err := kernel.ParseMoney("4000", pl.Nominal().Currency())
	if err != nil {
		t.Fatalf("ParseMoney: %v", err)
	}
	if got := l.Cost(); !got.Equal(want) {
		t.Errorf("Cost() = %s, want %s: 40 bonds at the 100 PLN nominal", got, want)
	}
	if got := l.Series().String(); got != "EDO0536" {
		t.Errorf("Series() = %s, want EDO0536", got)
	}
}

func TestBondLotCostByPurchase(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		series   string
		purchase ingest.BondPurchase
		want     string
	}{
		{name: "cash at the nominal", series: "EDO0536", purchase: ingest.PurchaseCash, want: "4000"},
		{name: "exchange at 99.90 PLN a bond (D-100)", series: "EDO0536", purchase: ingest.PurchaseExchange, want: "3996"},
		{name: "OTS exchanges at the nominal", series: "OTS0127", purchase: ingest.PurchaseExchange, want: "4000"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := bondLotCase{series: tt.series, purchased: "2026-10-27", purchase: tt.purchase}
			l := c.lot(t)
			want, err := kernel.ParseMoney(tt.want, pl.Nominal().Currency())
			if err != nil {
				t.Fatalf("ParseMoney: %v", err)
			}
			if got := l.Cost(); !got.Equal(want) {
				t.Errorf("Cost() = %s, want %s", got, want)
			}
			if got := l.Key(); got.Series != l.Series() || got.Purchased != l.Purchased {
				t.Errorf("Key() = %v, want the series and purchase date", got)
			}
		})
	}
	if got := []string{ingest.PurchaseCash.String(), ingest.PurchaseExchange.String(), ingest.BondPurchase(5).String()}; !cmp.Equal(got, []string{"cash", "exchange", "BondPurchase(5)"}) {
		t.Errorf("BondPurchase names = %v", got)
	}
}

func TestBondLotKeyValidate(t *testing.T) {
	t.Parallel()
	c := bondLotCase{}
	l := c.lot(t)
	tests := []struct {
		name    string
		key     ingest.BondLotKey
		wantErr error
	}{
		{name: "a series and a date", key: l.Key()},
		{name: "no date", key: ingest.BondLotKey{Series: l.Series()}, wantErr: ingest.ErrInvalidBondLot},
		{name: "no series", key: ingest.BondLotKey{Purchased: l.Purchased}, wantErr: ingest.ErrInvalidBondLot},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tt.wantErr, tt.key.Validate(), cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Validate (-want +got):\n%s", diff)
			}
		})
	}
}
