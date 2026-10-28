package ibkrflex_test

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/ibkrflex"
	"github.com/arkadiuss/openfief/internal/ledger"
)

// TestMapRefusesHostileInput feeds malformed and hostile statements to Map:
// each must fail with ErrMalformed or ErrTooLarge, without reading DTDs or
// expanding entities.
func TestMapRefusesHostileInput(t *testing.T) {
	t.Parallel()
	small := ibkrflex.Limits{MaxBytes: 4096, MaxDepth: 8, MaxElements: 20, MaxAttrs: 10, MaxAttrLen: 64}
	tests := []struct {
		name    string
		input   string
		limits  ibkrflex.Limits
		wantErr error
	}{
		{
			name:    "billion laughs",
			input:   `<?xml version="1.0"?><!DOCTYPE lolz [<!ENTITY lol "lol"><!ENTITY lol2 "&lol;&lol;&lol;&lol;">]><FlexQueryResponse>&lol2;</FlexQueryResponse>`,
			wantErr: ibkrflex.ErrMalformed,
		},
		{
			name:    "external entity",
			input:   `<?xml version="1.0"?><!DOCTYPE r [<!ENTITY xxe SYSTEM "file:///etc/passwd">]><FlexQueryResponse a="&xxe;"/>`,
			wantErr: ibkrflex.ErrMalformed,
		},
		{
			name:    "undeclared entity reference",
			input:   `<FlexQueryResponse a="&xxe;"/>`,
			wantErr: ibkrflex.ErrMalformed,
		},
		{
			name:    "non-UTF-8 encoding",
			input:   `<?xml version="1.0" encoding="ISO-8859-2"?><FlexQueryResponse/>`,
			wantErr: ibkrflex.ErrMalformed,
		},
		{
			name:    "truncated",
			input:   `<FlexQueryResponse><FlexStatements><FlexStatement accountId="U0000000" toDate="20260313">`,
			wantErr: ibkrflex.ErrMalformed,
		},
		{
			name:    "empty",
			input:   "",
			wantErr: ibkrflex.ErrMalformed,
		},
		{
			name:    "another root element",
			input:   `<FlexStatementResponse timestamp="x"><Status>Fail</Status><ErrorCode>1015</ErrorCode></FlexStatementResponse>`,
			wantErr: ibkrflex.ErrMalformed,
		},
		{
			name:    "statement without a readable toDate",
			input:   `<FlexQueryResponse><FlexStatements><FlexStatement accountId="U0000000" toDate="soon"/></FlexStatements></FlexQueryResponse>`,
			wantErr: ibkrflex.ErrMalformed,
		},
		{
			name:    "over the size cap",
			input:   statementXML(`<CashReport>` + strings.Repeat(`<CashReportCurrency currency="USD" endingCash="1" />`, 100) + `</CashReport>`),
			limits:  ibkrflex.Limits{MaxBytes: 1024},
			wantErr: ibkrflex.ErrTooLarge,
		},
		{
			name:    "nested too deep",
			input:   `<FlexQueryResponse>` + strings.Repeat("<a>", 9) + strings.Repeat("</a>", 9) + `</FlexQueryResponse>`,
			limits:  small,
			wantErr: ibkrflex.ErrTooLarge,
		},
		{
			name:    "too many elements",
			input:   statementXML(`<Trades>` + strings.Repeat(`<Trade/>`, 30) + `</Trades>`),
			limits:  small,
			wantErr: ibkrflex.ErrTooLarge,
		},
		{
			name:    "too many attributes",
			input:   `<FlexQueryResponse a="1" b="1" c="1" d="1" e="1" f="1" g="1" h="1" i="1" j="1" k="1"/>`,
			limits:  small,
			wantErr: ibkrflex.ErrTooLarge,
		},
		{
			name:    "attribute too long",
			input:   `<FlexQueryResponse a="` + strings.Repeat("x", 65) + `"/>`,
			limits:  small,
			wantErr: ibkrflex.ErrTooLarge,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			m := testMapper()
			m.Limits = tt.limits
			got, err := m.Map(strings.NewReader(tt.input), ingest.ModeTransactions)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("Map error (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(ingest.Mapping{}, got, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("Map returned a mapping with its error:\n%s", diff)
			}
		})
	}
}

// TestMapAtTheLimits checks that input exactly at each limit is accepted.
func TestMapAtTheLimits(t *testing.T) {
	t.Parallel()
	input := statementXML(`<CashReport><CashReportCurrency currency="USD" endingCash="1" /></CashReport>`)
	m := ibkrflex.Mapper{
		Accounts: map[string]ledger.AccountID{"U0000000": "acc-ibkr"},
		Limits:   ibkrflex.Limits{MaxBytes: int64(len(input)), MaxDepth: 5, MaxElements: 5, MaxAttrs: 3, MaxAttrLen: 8},
	}
	got, err := m.Map(strings.NewReader(input), ingest.ModeTransactions)
	if err != nil {
		t.Fatalf("Map: %v", err)
	}
	if want := "assert acc-ibkr 2026-03-13 cash:USD 1\n"; render(&got) != want {
		t.Errorf("Map = %q, want %q", render(&got), want)
	}
}

func TestMapUnknownMode(t *testing.T) {
	t.Parallel()
	_, err := testMapper().Map(strings.NewReader(statementXML("")), ingest.Mode(0))
	if !errors.Is(err, ibkrflex.ErrUnknownMode) {
		t.Errorf("Map error = %v, want ErrUnknownMode", err)
	}
}

// errReader fails every read.
type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }

func TestMapReadFailure(t *testing.T) {
	t.Parallel()
	_, err := testMapper().Map(errReader{}, ingest.ModeTransactions)
	if !errors.Is(err, ibkrflex.ErrMalformed) || !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Errorf("Map error = %v, want ErrMalformed wrapping the read error", err)
	}
}
