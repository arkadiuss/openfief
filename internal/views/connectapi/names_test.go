package connectapi_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"testing"

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
