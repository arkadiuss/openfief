package connectapi_test

import (
	"context"
	"log/slog"
	"maps"
	"net/http/httptest"
	"slices"
	"testing"
	"time"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/testing/protocmp"

	gocmp "github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/ingest"
	"github.com/arkadiuss/openfief/internal/ingest/app"
	"github.com/arkadiuss/openfief/internal/ingest/connectapi"
	"github.com/arkadiuss/openfief/internal/jurisdiction"
	"github.com/arkadiuss/openfief/internal/jurisdiction/pl"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/ledger"
	"github.com/arkadiuss/openfief/internal/portfolio"
	ingestv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/ingest/v1/ingestv1connect"
)

// memRegistry is the ledger's accounts over the default sleeves, in
// memory, or failing with err, with the sleeve each account is in, their
// payout settings, the accounts' current transactions and what was
// imported, for a payout change (D-162).
type memRegistry struct {
	accounts []ledger.Account
	in       map[ledger.AccountID]portfolio.SleeveID
	payouts  map[ledger.AccountID]ingest.PayoutSetting
	current  []ledger.Transaction
	imported []ledger.Transaction
	err      error
}

func (m *memRegistry) AccountSleeves(context.Context) (map[ledger.AccountID]portfolio.SleeveID, error) {
	return maps.Clone(m.in), m.err
}

func (m *memRegistry) PlaceAccount(_ context.Context, account ledger.AccountID, sleeve portfolio.SleeveID) error {
	m.in[account] = sleeve
	return nil
}

func (m *memRegistry) RemoveAccount(_ context.Context, account ledger.AccountID) error {
	delete(m.in, account)
	return nil
}

func (m *memRegistry) PayoutSetting(_ context.Context, account ledger.AccountID) (ingest.PayoutSetting, error) {
	return m.payouts[account], m.err
}

func (m *memRegistry) SavePayoutSetting(_ context.Context, account ledger.AccountID, p ingest.PayoutSetting) error {
	m.payouts[account] = p
	return nil
}

func (m *memRegistry) CurrentTransactions(context.Context, ledger.AccountID) ([]ledger.Transaction, error) {
	return m.current, m.err
}

func (m *memRegistry) Import(_ context.Context, txs []ledger.Transaction, _ []ledger.BalanceAssertion) (ingest.ImportSummary, error) {
	m.imported = append(m.imported, txs...)
	return ingest.ImportSummary{}, m.err
}

func (m *memRegistry) Accounts(context.Context) ([]ledger.Account, error) { return m.accounts, m.err }

func (m *memRegistry) Sleeves(context.Context) ([]portfolio.Sleeve, error) {
	return portfolio.DefaultSleeves(), m.err
}

func (m *memRegistry) SaveAccount(_ context.Context, a ledger.Account) error {
	m.accounts = append(slices.DeleteFunc(m.accounts, func(o ledger.Account) bool { return o.ID == a.ID }), a)
	return nil
}

// memLinks keeps connector bindings and the synthetic broker accounts
// U0000001, unmapped, and U0000002, mapped to acc-bonds, of ibkr-main.
type memLinks struct {
	bound   []string
	brokers []ingest.BrokerAccount
}

func (m *memLinks) Bind(_ context.Context, instance string, account ledger.AccountID, _ time.Time) error {
	m.bound = append(m.bound, instance+">"+string(account))
	return nil
}

func (m *memLinks) BoundAccounts(context.Context, string) ([]ledger.AccountID, error) {
	return nil, nil
}

func (m *memLinks) RecordBrokerAccounts(context.Context, string, ingest.InstitutionID, []ingest.BrokerAccountID, kernel.Date) error {
	return nil
}

func (m *memLinks) BrokerAccounts(context.Context, string) ([]ingest.BrokerAccount, error) {
	return m.brokers, nil
}

func (m *memLinks) MarkSeeded(context.Context, string, []ingest.BrokerAccountID) error { return nil }

func (m *memLinks) MapBrokerAccount(_ context.Context, _ string, id ingest.BrokerAccountID, account ledger.AccountID, _ time.Time) error {
	for i := range m.brokers {
		if m.brokers[i].ID == id {
			m.brokers[i].Account = account
		}
	}
	return nil
}

// memTerms keeps the accounts' tax wrappers and fee rules in memory.
type memTerms struct {
	wrappers map[ledger.AccountID]jurisdiction.Wrapper
	fees     map[ledger.AccountID]jurisdiction.FeeRule
}

func (m *memTerms) RecordTaxWrapper(_ context.Context, account ledger.AccountID, w jurisdiction.Wrapper) error {
	m.wrappers[account] = w
	return nil
}

func (m *memTerms) RecordFeeRule(_ context.Context, account ledger.AccountID, r *jurisdiction.FeeRule) error {
	m.fees[account] = *r
	return nil
}

func (m *memTerms) TaxWrapper(_ context.Context, account ledger.AccountID) (jurisdiction.Wrapper, bool, error) {
	w, ok := m.wrappers[account]
	return w, ok, nil
}

func (m *memTerms) FeeRule(_ context.Context, account ledger.AccountID) (jurisdiction.FeeRule, error) {
	return m.fees[account], nil
}

// accountsClient is the generated AccountService client.
type accountsClient struct {
	ingestv1connect.AccountServiceClient
}

// newAccountsClient serves the AccountService over the fakes and returns
// a client calling it with the caller token.
func newAccountsClient(t *testing.T, reg *memRegistry, links *memLinks, caller string) accountsClient {
	t.Helper()
	accounts := &app.Accounts{
		Owners: owners{}, Ledger: reg, Bindings: links, Brokers: links, Pack: pl.NewPack(), Payouts: reg, Sleeves: reg, PayoutSettings: reg,
		Terms: &memTerms{wrappers: map[ledger.AccountID]jurisdiction.Wrapper{}, fees: map[ledger.AccountID]jurisdiction.FeeRule{}},
		Now:   func() time.Time { return time.Date(2026, time.March, 2, 6, 0, 0, 0, time.UTC) },
	}
	_, handler := connectapi.NewAccounts(accounts, viewOnly(), slog.New(slog.DiscardHandler))
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return accountsClient{ingestv1connect.NewAccountServiceClient(srv.Client(), srv.URL, connect.WithInterceptors(bearer(caller)))}
}

// fixture returns the fakes: acc-bonds stored, ibkr-main's broker accounts.
func fixture() (*memRegistry, *memLinks) {
	return &memRegistry{
			accounts: []ledger.Account{{ID: bondsAccount, Institution: "treasury-bonds", Name: "Bonds"}},
			in:       map[ledger.AccountID]portfolio.SleeveID{bondsAccount: "bonds"},
			payouts:  map[ledger.AccountID]ingest.PayoutSetting{},
		},
		&memLinks{brokers: []ingest.BrokerAccount{
			{Instance: "ibkr-main", ID: "U0000001", Institution: "ibkr"},
			{Instance: "ibkr-main", ID: "U0000002", Institution: "ibkr", Account: bondsAccount},
		}}
}

func TestAccountServiceRoundTrip(t *testing.T) {
	t.Parallel()
	reg, links := fixture()
	c := newAccountsClient(t, reg, links, viewCaller)
	ctx := t.Context()
	created, err := c.CreateAccount(ctx, connect.NewRequest(&ingestv1.CreateAccountRequest{
		Origin: local, SessionToken: ownerSession, Name: "IBKR main", Sleeve: "satellite",
		ConnectorInstance: "ibkr-main", BrokerAccount: "U0000001", OpenedOnSeedingDay: true,
	}))
	want := &ingestv1.Account{
		Id: "ibkr-main", Institution: "ibkr", Name: "IBKR main", Sleeve: "satellite", TaxWrapper: "taxable",
		FeeRule: &ingestv1.FeeRule{Rate: "0", Minimum: "0", ConversionRate: "0"}, BondPayout: ingestv1.BondPayout_BOND_PAYOUT_CASH,
		OpenedOnSeedingDay: true,
	}
	if err != nil || !gocmp.Equal(want, created.Msg.GetAccount(), protocmp.Transform()) {
		t.Fatalf("CreateAccount = %v, %v; want %v", created, err, want)
	}
	fee := &ingestv1.FeeRule{Rate: "0.0005", Minimum: "1", MinimumCurrency: "USD", ConversionRate: "0.002"}
	updated, err := c.UpdateAccount(ctx, connect.NewRequest(&ingestv1.UpdateAccountRequest{
		Origin: local, SessionToken: ownerSession, AccountId: "ibkr-main", Name: "IBKR", Sleeve: "core",
		TaxWrapper: "ikze", FeeRule: fee,
	}))
	// An update that doesn't name the mark of an account opened on its
	// seeding day keeps it (D-168); one that does changes it.
	want = &ingestv1.Account{
		Id: "ibkr-main", Institution: "ibkr", Name: "IBKR", Sleeve: "core", TaxWrapper: "ikze", FeeRule: fee,
		BondPayout: ingestv1.BondPayout_BOND_PAYOUT_CASH, OpenedOnSeedingDay: true,
	}
	if err != nil || !gocmp.Equal(want, updated.Msg.GetAccount(), protocmp.Transform()) {
		t.Errorf("UpdateAccount = %v, %v; want %v", updated, err, want)
	}
	unmark := false
	updated, err = c.UpdateAccount(ctx, connect.NewRequest(&ingestv1.UpdateAccountRequest{
		Origin: local, SessionToken: ownerSession, AccountId: "ibkr-main", Name: "IBKR", Sleeve: "core", OpenedOnSeedingDay: &unmark,
	}))
	if err != nil || updated.Msg.GetAccount().GetOpenedOnSeedingDay() {
		t.Errorf("UpdateAccount unmarking = %v, %v; want the mark cleared", updated, err)
	}
	if _, err := c.BindAccount(ctx, connect.NewRequest(&ingestv1.BindAccountRequest{
		Origin: local, SessionToken: ownerSession, AccountId: bondsAccount, ConnectorInstance: "bonds-main",
	})); err != nil {
		t.Errorf("BindAccount: %v", err)
	}
	if diff := gocmp.Diff([]string{"ibkr-main>ibkr-main", "bonds-main>acc-bonds"}, links.bound); diff != "" || links.brokers[0].Account != "ibkr-main" {
		t.Errorf("bindings (-want +got):\n%s\nU0000001 mapped to %q", diff, links.brokers[0].Account)
	}
}

// TestAccountServiceBondPayout pays out a hand-entered bond account's
// coupons and maturities from today (D-135), then keeps them as cash from
// an earlier day (D-162).
func TestAccountServiceBondPayout(t *testing.T) {
	t.Parallel()
	reg, links := fixture()
	c := newAccountsClient(t, reg, links, viewCaller)
	ctx := t.Context()
	// A hand-entered bond account's coupons and maturities paid out from today (D-135).
	paid, err := c.UpdateAccount(ctx, connect.NewRequest(&ingestv1.UpdateAccountRequest{
		Origin: local, SessionToken: ownerSession, AccountId: bondsAccount, Name: "Bonds", Sleeve: "bonds",
		BondPayout: ingestv1.BondPayout_BOND_PAYOUT_PAID_OUT,
	}))
	if err != nil || paid.Msg.GetAccount().GetBondPayout() != ingestv1.BondPayout_BOND_PAYOUT_PAID_OUT ||
		paid.Msg.GetAccount().GetBondPayoutSince() != "2026-03-02" {
		t.Errorf("UpdateAccount paying out = %v, %v; want paid out since 2026-03-02", paid, err)
	}

	if got := reg.payouts[bondsAccount]; got.Payout != ingest.PayoutPaidOut || got.Since.String() != "2026-03-02" {
		t.Errorf("stored payout = %+v, want paid out since 2026-03-02", got)
	}
	checkPayoutBackToCash(t, c, reg)
}

// checkPayoutBackToCash changes acc-bonds, paid out since 2026-03-02, back
// to cash from an earlier day, then refuses a change from tomorrow.
func checkPayoutBackToCash(t *testing.T, c accountsClient, reg *memRegistry) {
	t.Helper()
	ctx := t.Context()
	// Back to cash from a day before the change the owner gives (D-162): the
	// account keeps its payouts as cash over its whole history again, and the
	// payout withdrawals booked from that day on are superseded, earlier ones stay.
	reg.current = []ledger.Transaction{
		withdrawal(t, "lot:COI1028:2024-10-10:coupon:2025-10-10:payout", "2025-10-10"),
		withdrawal(t, "lot:COI1028:2024-10-10:coupon:2026-01-10:payout", "2026-01-10"),
	}
	cash, err := c.UpdateAccount(ctx, connect.NewRequest(&ingestv1.UpdateAccountRequest{
		Origin: local, SessionToken: ownerSession, AccountId: bondsAccount, Name: "Bonds", Sleeve: "bonds",
		BondPayout: ingestv1.BondPayout_BOND_PAYOUT_CASH, BondPayoutFrom: "2026-01-01",
	}))
	if err != nil || cash.Msg.GetAccount().GetBondPayout() != ingestv1.BondPayout_BOND_PAYOUT_CASH ||
		cash.Msg.GetAccount().GetBondPayoutSince() != "" {
		t.Errorf("UpdateAccount back to cash = %v, %v; want cash over the whole history", cash, err)
	}
	if len(reg.imported) != 1 || reg.imported[0].Kind != ledger.KindDeletion || reg.imported[0].Date.String() != "2026-01-10" {
		t.Errorf("imported %v, want the deletion of the payout withdrawal of 2026-01-10", reg.imported)
	}
	// A change can't take effect after today.
	_, err = c.UpdateAccount(ctx, connect.NewRequest(&ingestv1.UpdateAccountRequest{
		Origin: local, SessionToken: ownerSession, AccountId: bondsAccount, Name: "Bonds", Sleeve: "bonds",
		BondPayout: ingestv1.BondPayout_BOND_PAYOUT_PAID_OUT, BondPayoutFrom: "2026-03-03",
	}))
	if connect.CodeOf(err) != connect.CodeInvalidArgument {
		t.Errorf("UpdateAccount from tomorrow = %v, want InvalidArgument", err)
	}
}

// withdrawal returns a synthetic payout withdrawal of acc-bonds keyed id.
func withdrawal(t *testing.T, id, day string) ledger.Transaction {
	t.Helper()
	d, err := kernel.ParseDate(day)
	if err != nil {
		t.Fatal(err)
	}
	pln, err := kernel.ParseCurrency("PLN")
	if err != nil {
		t.Fatal(err)
	}
	w, err := ledger.NewWithdrawal(ledger.TransactionHeader{
		Key: ledger.TransactionKey{Source: "bonds-main:treasury-bonds:acc-bonds", ExternalID: id}, Account: bondsAccount, Date: d,
	}, kernel.NewMoney(kernel.NewDecimal(5305, -2), pln))
	if err != nil {
		t.Fatal(err)
	}
	return w
}

func TestAccountServiceErrors(t *testing.T) {
	t.Parallel()
	ctx := t.Context()
	create := func(session, name, sleeve, instance, broker string) *connect.Request[ingestv1.CreateAccountRequest] {
		return connect.NewRequest(&ingestv1.CreateAccountRequest{
			Origin: local, SessionToken: session, Name: name, Institution: "x", Sleeve: sleeve, ConnectorInstance: instance, BrokerAccount: broker,
		})
	}
	tests := []struct {
		name   string
		caller string
		down   bool
		call   func(c accountsClient) connect.Code
		want   connect.Code
	}{
		{
			name: "not a view plugin", caller: "another-plugin", want: connect.CodePermissionDenied,
			call: func(c accountsClient) connect.Code {
				_, err := c.CreateAccount(ctx, create(ownerSession, "Cash", "core", "", ""))
				return connect.CodeOf(err)
			},
		},
		{name: "no session", want: connect.CodeUnauthenticated, call: func(c accountsClient) connect.Code {
			_, err := c.CreateAccount(ctx, create("nope", "Cash", "core", "", ""))
			return connect.CodeOf(err)
		}},
		{name: "a read session", want: connect.CodePermissionDenied, call: func(c accountsClient) connect.Code {
			_, err := c.CreateAccount(ctx, create(readSession, "Cash", "core", "", ""))
			return connect.CodeOf(err)
		}},
		{name: "a taken ID", want: connect.CodeAlreadyExists, call: func(c accountsClient) connect.Code {
			_, err := c.CreateAccount(ctx, create(ownerSession, "Acc bonds", "core", "", ""))
			return connect.CodeOf(err)
		}},
		{name: "an unknown sleeve", want: connect.CodeInvalidArgument, call: func(c accountsClient) connect.Code {
			_, err := c.CreateAccount(ctx, create(ownerSession, "Cash", "gold", "", ""))
			return connect.CodeOf(err)
		}},
		{name: "a malformed instance", want: connect.CodeInvalidArgument, call: func(c accountsClient) connect.Code {
			_, err := c.CreateAccount(ctx, create(ownerSession, "Cash", "core", "IBKR", ""))
			return connect.CodeOf(err)
		}},
		{name: "a broker account mapped elsewhere", want: connect.CodeFailedPrecondition, call: func(c accountsClient) connect.Code {
			_, err := c.CreateAccount(ctx, create(ownerSession, "Cash", "core", "ibkr-main", "U0000002"))
			return connect.CodeOf(err)
		}},
		{name: "a broker account never reported", want: connect.CodeFailedPrecondition, call: func(c accountsClient) connect.Code {
			_, err := c.CreateAccount(ctx, create(ownerSession, "Cash", "core", "ibkr-main", "U0000009"))
			return connect.CodeOf(err)
		}},
		{name: "a wrapper the pack doesn't have", want: connect.CodeInvalidArgument, call: func(c accountsClient) connect.Code {
			req := create(ownerSession, "Cash", "core", "", "")
			req.Msg.TaxWrapper = "ppk"
			_, err := c.CreateAccount(ctx, req)
			return connect.CodeOf(err)
		}},
		{name: "a fee rate that is no decimal", want: connect.CodeInvalidArgument, call: func(c accountsClient) connect.Code {
			_, err := c.UpdateAccount(ctx, connect.NewRequest(&ingestv1.UpdateAccountRequest{
				Origin: local, SessionToken: ownerSession, AccountId: bondsAccount, Name: "X", Sleeve: "core",
				FeeRule: &ingestv1.FeeRule{Rate: "0,29%"},
			}))
			return connect.CodeOf(err)
		}},
		{name: "paying out an account that is not a hand-entered bond account", want: connect.CodeInvalidArgument, call: func(c accountsClient) connect.Code {
			req := create(ownerSession, "Cash", "core", "", "")
			req.Msg.BondPayout = ingestv1.BondPayout_BOND_PAYOUT_PAID_OUT
			_, err := c.CreateAccount(ctx, req)
			return connect.CodeOf(err)
		}},
		{name: "a payout setting outside the declared ones", want: connect.CodeInvalidArgument, call: func(c accountsClient) connect.Code {
			_, err := c.UpdateAccount(ctx, connect.NewRequest(&ingestv1.UpdateAccountRequest{
				Origin: local, SessionToken: ownerSession, AccountId: bondsAccount, Name: "Bonds", Sleeve: "bonds", BondPayout: 7,
			}))
			return connect.CodeOf(err)
		}},
		{name: "an account the owner may not change", want: connect.CodePermissionDenied, call: func(c accountsClient) connect.Code {
			_, err := c.UpdateAccount(ctx, connect.NewRequest(&ingestv1.UpdateAccountRequest{Origin: local, SessionToken: ownerSession, AccountId: deniedAccount, Name: "X", Sleeve: "core"}))
			return connect.CodeOf(err)
		}},
		{name: "binding an account the owner may not change", want: connect.CodePermissionDenied, call: func(c accountsClient) connect.Code {
			_, err := c.BindAccount(ctx, connect.NewRequest(&ingestv1.BindAccountRequest{Origin: local, SessionToken: ownerSession, AccountId: deniedAccount, ConnectorInstance: "ibkr-main"}))
			return connect.CodeOf(err)
		}},
		{name: "an unknown account", want: connect.CodeNotFound, call: func(c accountsClient) connect.Code {
			_, err := c.UpdateAccount(ctx, connect.NewRequest(&ingestv1.UpdateAccountRequest{Origin: local, SessionToken: ownerSession, AccountId: "nope", Name: "X", Sleeve: "core"}))
			return connect.CodeOf(err)
		}},
		{name: "the ledger down", down: true, want: connect.CodeInternal, call: func(c accountsClient) connect.Code {
			_, err := c.BindAccount(ctx, connect.NewRequest(&ingestv1.BindAccountRequest{Origin: local, SessionToken: ownerSession, AccountId: bondsAccount, ConnectorInstance: "x"}))
			return connect.CodeOf(err)
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			reg, links := fixture()
			if tt.down {
				reg.err = errDown
			}
			caller := viewCaller
			if tt.caller != "" {
				caller = tt.caller
			}
			if got := tt.call(newAccountsClient(t, reg, links, caller)); got != tt.want {
				t.Errorf("code = %v, want %v", got, tt.want)
			}
			if len(links.bound) != 0 {
				t.Errorf("a refused call bound %v", links.bound)
			}
		})
	}
}
