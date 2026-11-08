package connectapi_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/views/app"
	"github.com/arkadiuss/openfief/internal/views/connectapi"
)

// errNamesDown is the names store failing.
var errNamesDown = errors.New("names store down")

// fakeNames names the synthetic ETF, or fails with err.
type fakeNames struct {
	err error
}

func (f fakeNames) Current(context.Context) (map[ledger.ISIN]string, error) {
	if f.err != nil {
		return nil, f.err
	}
	isin, err := ledger.ParseISIN(isinIE)
	if err != nil {
		return nil, fmt.Errorf("parsing: %w", err)
	}
	return map[ledger.ISIN]string{isin: "Synthetic World ETF"}, nil
}

// namesAPI is the read API over the uiv2 fixture with names from n.
func namesAPI(t *testing.T, n fakeNames) api {
	t.Helper()
	var log bytes.Buffer
	h, err := connectapi.New(newViews(t, newLedger(t), app.WithInstrumentNames(n)), fakeVerifier{},
		connectapi.WithLogger(slog.New(slog.NewJSONHandler(&log, nil))))
	if err != nil {
		t.Fatalf("connectapi.New: %v", err)
	}
	return api{handler: h, log: &log}
}

// TestRESTInstrumentNames checks that every instrument an answer names
// carries its display name (D-197), one without a name none, and that a
// store that can't read the names fails the answer.
func TestRESTInstrumentNames(t *testing.T) {
	t.Parallel()
	a := namesAPI(t, fakeNames{})
	var got struct {
		Holdings []struct {
			Instrument struct {
				Isin string `json:"isin"`
				Name string `json:"name"`
			} `json:"instrument"`
		} `json:"holdings"`
	}
	serveJSON(t, a, "/v1/accounts/acc-1/holdings?date=2026-03-02", http.StatusOK, &got)
	names := map[string]string{}
	for _, h := range got.Holdings {
		names[h.Instrument.Isin] = h.Instrument.Name
	}
	if names[isinIE] != "Synthetic World ETF" || names[isinPL] != "" || len(names) != 2 {
		t.Errorf("names = %v, want the ETF's only", names)
	}
	var detail struct {
		Instrument struct {
			Name string `json:"name"`
		} `json:"instrument"`
	}
	serveJSON(t, a, "/v1/instruments/"+isinIE, http.StatusOK, &detail)
	if detail.Instrument.Name != "Synthetic World ETF" {
		t.Errorf("GetInstrument name = %q", detail.Instrument.Name)
	}
	failing := namesAPI(t, fakeNames{err: errNamesDown})
	serveJSON(t, failing, "/v1/accounts/acc-1/holdings?date=2026-03-02", http.StatusInternalServerError, nil)
	serveJSON(t, failing, "/v1/sleeves", http.StatusOK, nil)
}

// isinProvisional is the synthetic provisional ISIN of series EDO0536
// (D-214).
const isinProvisional = "PLEDO0536000"

// TestRESTProvisionalISIN checks that an instrument held under its
// series' provisional ISIN is marked as such in a holding and in its
// instrument detail, and that a registry-like ISIN or another
// instrument's is not (D-214).
func TestRESTProvisionalISIN(t *testing.T) {
	t.Parallel()
	l := newLedger(t)
	bond, on := instrument(t, isinProvisional, "EDO0536"), date(t, "2026-03-02")
	l.txs = append(l.txs, ledger.Transaction{
		Key: ledger.TransactionKey{Source: "test", ExternalID: "buy-2"}, Account: "acc-1", Kind: ledger.KindBuy, Date: on,
		Description: "Synthetic bond buy", Postings: []ledger.Posting{
			{Book: ledger.HoldingBook(bond), Amount: money(t, "500.00", "PLN"), Quantity: kernel.NewQuantity(dec(t, "5"))},
			{Book: ledger.CashBook(cur(t, "PLN")), Amount: money(t, "-500.00", "PLN")},
		},
	})
	l.holdings = append(l.holdings, ledger.Holding{Account: "acc-1", Instrument: bond, Date: on, Quantity: kernel.NewQuantity(dec(t, "5"))})
	var log bytes.Buffer
	h, err := connectapi.New(newViews(t, l), fakeVerifier{}, connectapi.WithLogger(slog.New(slog.NewJSONHandler(&log, nil))))
	if err != nil {
		t.Fatalf("connectapi.New: %v", err)
	}
	a := api{handler: h, log: &log}
	var got struct {
		Holdings []struct {
			Instrument struct {
				Isin        string `json:"isin"`
				Provisional bool   `json:"isinProvisional"`
			} `json:"instrument"`
		} `json:"holdings"`
	}
	serveJSON(t, a, "/v1/accounts/acc-1/holdings?date=2026-03-02", http.StatusOK, &got)
	marked := map[string]bool{}
	for _, h := range got.Holdings {
		marked[h.Instrument.Isin] = h.Instrument.Provisional
	}
	want := map[string]bool{isinPL: false, isinIE: false, isinProvisional: true}
	if diff := cmp.Diff(want, marked); diff != "" {
		t.Errorf("isin_provisional by ISIN (-want +got):\n%s", diff)
	}
	for isin, want := range map[string]bool{isinProvisional: true, isinPL: false} {
		var detail struct {
			Instrument struct {
				Provisional bool `json:"isinProvisional"`
			} `json:"instrument"`
		}
		serveJSON(t, a, "/v1/instruments/"+isin, http.StatusOK, &detail)
		if detail.Instrument.Provisional != want {
			t.Errorf("GetInstrument(%s) isin_provisional = %t, want %t", isin, detail.Instrument.Provisional, want)
		}
	}
}
