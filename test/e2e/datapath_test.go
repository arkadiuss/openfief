//go:build e2e

package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arkadiuss/openfief/internal/ingest/xtb/xtbtest"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/portfolio"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// The data path's synthetic data, which the fixture serves
// (test/e2e/fixture/datapath.go) and this test uploads: the Flex query IDs
// that pick its IBKR statements, its broker accounts, instruments and
// dates.
const (
	flexOpening   = "e2e-datapath-opening"
	flexWeek      = "e2e-datapath-week"
	flexCorrected = "e2e-datapath-week-corrected"
	// ibkrBroker is the IBKR account the statements report.
	ibkrBroker = "U0000002"
	// The synthetic instruments: TSTG in GBP and TSTH in CHF at IBKR, and
	// the XTB ticker TSTJ.NO in NOK.
	isinG     = "GBTEST000047"
	isinH     = "CHTEST000052"
	isinJ     = "NOTEST000068"
	xtbTicker = "TSTJ.NO"
	// goLive is the day the IBKR and XTB accounts are seeded on, and
	// weekEnd the last day of the IBKR week after it.
	goLive  = "2026-03-06"
	weekEnd = "2026-03-13"
	// correctedID is the deposit IBKR corrects: 1000 GBP, then 1100.
	correctedID = "8100201"
	// The owner's accounts, whose IDs the core makes from their names.
	ibkrAccountName  = "E2E Data Path IBKR"
	xtbAccountName   = "E2E Data Path XTB"
	bondsAccountName = "E2E Data Path Bonds"
	// dataPathTimeout bounds each wait for an import or a revaluation.
	dataPathTimeout = 5 * time.Minute
)

// TestDataPath runs the connectors' data path end to end against the
// institution fixture, as an owner does, and checks D-200's criteria 2, 3,
// 4, 5, 6 and 7 on it, criterion 8's charts and criterion 19. It deploys the fixture as an HTTPS Service with a
// certificate from a CA made for this run, and installs, with the openfief
// CLI built from this checkout, nbp-fx, eodhd, ibkr-flex and
// treasury-bonds, each with its base URL on the fixture, the fixture's
// Service as its in-cluster service (D-74) and the run's CA in its
// <instance>--ca-bundle, and the XTB parser, all signed with the run's key.
// The owner, signed in with a passkey in Chromium (owner.mjs), creates an
// IBKR account mapped to the statement's broker account, an XTB account and
// a treasury bond account. The IBKR account is seeded from the opening
// statement's lot-level positions and synced through the week after it and
// IBKR's correction of one deposit, each statement picked by the query ID
// in its Secret; the XTB account is seeded from an uploaded export of open
// positions, its ticker confirmed, then a later export uploaded twice; a
// TOS and an EDO lot are entered with their terms, the TOS lot twice, while
// the fixture withholds the Ministry of Finance's data, so both are booked
// under their provisional ISINs; then the fixture publishes it, a sync
// re-keys both to the registry's ISINs, and a COI lot is entered with its
// terms left to the connector, whose first coupon a sync books. Then:
//
//   - criterion 2 (SCEN-e2e-double-entry-whole-log): `core check ledger`
//     finds every version of every transaction balanced and the whole log
//     summing to zero per currency, and the read API's postings of each
//     data path account sum to zero per currency;
//   - criterion 3 (SCEN-e2e-reimport-idempotent, SCEN-e2e-correction-new-version):
//     syncing the same statement and uploading the same file and lot again
//     adds no transaction and no version, and the corrected deposit is a
//     second version, the first kept and superseded, which the read API
//     shows at its corrected amount;
//   - criterion 4 (SCEN-e2e-opening-lots-pln-cost): the opening lots keep
//     their open dates, units and costs, and the net values, whose tax is
//     on the gain over the PLN cost at NBP's rate of the day before each
//     open date, match the hand-computed ones;
//   - criterion 5 (SCEN-e2e-bond-values-reference, SCEN-e2e-bond-net-if-redeemed,
//     SCEN-e2e-bond-registry-isin-and-terms): each series' gross accrued
//     value on two days equals the reference valuer's and the hand
//     calculation's, its net and liquidation values and the account's
//     totals the hand calculation's (bonds_test.go), and the read API shows
//     each series under its registry ISIN with its terms, the published
//     ones with their attribution;
//   - criterion 6 (SCEN-e2e-nbp-rates-every-conversion, SCEN-e2e-eodhd-prices,
//     SCEN-e2e-cross-check-flagged): the PLN values and cash match the
//     fixture's NBP rates of their day, the prices are EODHD's from the
//     eodhd instance, and IBKR's mark 1.96% above it is flagged while one
//     0.48% above it matches;
//   - criterion 7 (SCEN-e2e-read-api-every-resource,
//     SCEN-e2e-read-api-openapi-conformance, SCEN-e2e-read-api-no-write,
//     SCEN-e2e-read-api-connect-equals-rest): every operation of the read
//     API's OpenAPI document, which lists exactly ReadService's routes,
//     answers the data path's record as the document describes, and 401
//     without a valid token; nothing but GET is served and the script
//     token reaches no owner service; and the SDK's Connect client answers
//     the same as REST (readapi_test.go);
//   - criterion 8's charts (SCEN-e2e-charts-match-read-api,
//     SCEN-e2e-charts-only-read-api): what the overview's value, return
//     and allocation charts, each data path account's charts and each
//     instrument's price chart draw in the owner's Chromium equals what
//     the read API answers the script token for the same days, and while
//     the UI drew them the core's log shows it calling nothing outside
//     the read API but the session check and two lookups without
//     numbers, and its pod reaches nothing but the core (charts_test.go);
//   - criterion 19 (SCEN-e2e-model-portfolio-off-shows-nothing,
//     SCEN-e2e-model-portfolio-backtest-hand-calculation): last, with the
//     model portfolio off the read API answers 404 and the overview draws
//     none of it; the owner saves two index targets, MODA 60% and MODB 40%,
//     and turns it on; the read API's backtest across a yearly rebalance
//     and its max drawdowns equal the hand calculation over the fixture's
//     synthetic index series, and the overview draws the same line and
//     lists the same drawdowns; turned off again, it is gone
//     (model_test.go).
func TestDataPath(t *testing.T) {
	t.Parallel()
	k := mustConnect(t)
	diagnoseOnFailure(t, k)
	if _, err := os.Stat(cliPath); err != nil {
		t.Fatalf("%v: %v", errNoCLI, err)
	}
	reg, err := registryFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	waitCoreAvailable(t, k)
	waitSeedComplete(t, k)
	waitPluginReady(t, k, uiPlugin)
	tokens := fixtureTokens{flex: syntheticCredential(t), eodhd: syntheticCredential(t)}
	ca := deployFixture(t, k, tokens)
	installDataPath(t, k, reg, dataPathInstances(ca, tokens))
	p := &dataPath{k: k, token: issueToken(t, k), owner: signInOwner(t, k), fixtureCA: ca}
	p.sleeve = firstSleeve(t, k, p.token)
	p.ibkr, p.xtb, p.bonds = accountID(t, ibkrAccountName), accountID(t, xtbAccountName), accountID(t, bondsAccountName)

	p.bookIBKR(t)
	p.bookXTB(t)
	p.bookBonds(t)
	p.rekeyBonds(t)
	p.waitPrices(t)

	p.checkDoubleEntry(t)
	p.checkOpeningLots(t)
	p.checkValues(t)
	p.checkBonds(t)
	p.checkPriceChecks(t)
	p.checkReadAPI(t)
	p.checkCharts(t)
	p.checkModelPortfolio(t)
}

// dataPath is the data path's state: the cluster, a script token for the
// read API, the signed-in owner, the sleeve the owner's accounts go in,
// their IDs, the run's CA the fixture's certificate is signed by, and what
// the bond account holds once its lots, re-key and coupons are booked.
type dataPath struct {
	k                *kube
	token            string
	owner            *owner
	sleeve           string
	ibkr, xtb, bonds string
	fixtureCA        []byte
	bondsBooked      ledgerCount
}

// bookIBKR seeds the IBKR account and books the weeks after it: nbp-fx and
// ibkr-flex sync, which reports the broker account; the owner creates the
// account mapped to it; the opening statement seeds three opening lots and
// cash in two currencies; the week books four transactions; the same week
// again changes nothing; IBKR's correction of the deposit is a second
// version, the first kept and superseded, and the read API shows it.
func (p *dataPath) bookIBKR(t *testing.T) {
	t.Helper()
	p.sync(t, dpNBP)
	p.sync(t, dpIBKR)
	p.owner.postForm(t, "/accounts", map[string]string{
		"name": ibkrAccountName, "sleeve": p.sleeve, "broker": dpIBKR + "/" + ibkrBroker,
	})
	p.syncUntil(t, dpIBKR, p.ibkr, ledgerCount{transactions: 5, versions: 5})
	p.setQuery(t, flexWeek)
	week := p.syncUntil(t, dpIBKR, p.ibkr, ledgerCount{transactions: 9, versions: 9})
	p.sync(t, dpIBKR)
	if again := p.ledger(t, p.ibkr); !cmp.Equal(week.Logged, again.Logged) {
		t.Errorf("syncing the same IBKR statement again changed the log (-before +after):\n%s", cmp.Diff(week.Logged, again.Logged))
	}
	p.setQuery(t, flexCorrected)
	corrected := p.syncUntil(t, dpIBKR, p.ibkr, ledgerCount{transactions: 9, versions: 10, superseded: 1})
	var deposit []loggedVersion
	for _, v := range corrected.Logged {
		if v.ExternalID == correctedID {
			deposit = append(deposit, loggedVersion{ExternalID: v.ExternalID, Version: v.Version, Kind: v.Kind, Superseded: v.Superseded})
		}
	}
	want := []loggedVersion{
		{ExternalID: correctedID, Version: 1, Kind: "deposit", Superseded: true},
		{ExternalID: correctedID, Version: 2, Kind: "deposit"},
	}
	if diff := cmp.Diff(want, deposit); diff != "" {
		t.Errorf("versions of the corrected deposit (-want +got):\n%s", diff)
	}
	p.checkCorrectedDeposit(t)
}

// checkCorrectedDeposit checks that the read API shows the corrected
// deposit's current version: 1100 GBP into cash.
func (p *dataPath) checkCorrectedDeposit(t *testing.T) {
	t.Helper()
	tx, err := findTransaction(t.Context(), p.k, p.token, p.ibkr, correctedID)
	if err != nil {
		t.Fatalf("the corrected deposit: %v", err)
	}
	var cash []string
	for _, posting := range tx.GetPostings() {
		if posting.GetBook().GetKind() == viewsv1.BookKind_BOOK_KIND_CASH {
			cash = append(cash, canonical(t, posting.GetAmount().GetAmount())+" "+posting.GetAmount().GetCurrency())
		}
	}
	if diff := cmp.Diff([]string{"1100 GBP"}, cash); diff != "" {
		t.Errorf("the corrected deposit's cash in the read API (-want +got):\n%s", diff)
	}
}

// bookXTB seeds the XTB account from an export of its open positions and
// books a later export: the owner creates the account bound to the XTB
// parser and uploads the opening export, which reports the broker account
// and books nothing; maps it; uploads the export again with the opening
// cash, which seeds the cash while the position waits for its ticker's
// ISIN; confirms the ticker, which books the opening lot; and uploads the
// later export twice, the second time changing nothing.
func (p *dataPath) bookXTB(t *testing.T) {
	t.Helper()
	p.owner.postForm(t, "/accounts", map[string]string{
		"name": xtbAccountName, "institution": xtbInstitution, "sleeve": p.sleeve, "instance": dpXTB,
	})
	opening := xtbOpening()
	p.owner.upload(t, p.xtb, opening)
	p.owner.postForm(t, "/accounts/bind", map[string]string{"account": p.xtb, "broker": dpXTB + "/" + xtbtest.Account})
	opening.openingCash, opening.currency = "500", "PLN"
	p.owner.upload(t, p.xtb, opening)
	p.owner.postForm(t, "/inbox/tickers/confirm", map[string]string{"instance": dpXTB, "symbol": xtbTicker, "isin": isinJ})
	p.waitLedger(t, p.xtb, ledgerCount{transactions: 2, versions: 2})
	week := xtbWeek()
	p.owner.upload(t, p.xtb, week)
	before := p.waitLedger(t, p.xtb, ledgerCount{transactions: 4, versions: 4})
	p.owner.upload(t, p.xtb, week)
	if again := p.ledger(t, p.xtb); !cmp.Equal(before.Logged, again.Logged) {
		t.Errorf("uploading the same XTB export again changed the log (-before +after):\n%s", cmp.Diff(before.Logged, again.Logged))
	}
}

// xtbOpening is a synthetic xStation export of the XTB account's open
// positions on the go-live day: 20 TSTJ.NO opened on Tuesday, February 3,
// 2026 at 50 NOK each, so its cost is 1000 NOK.
func xtbOpening() *statementFile {
	e := xtbtest.Export{
		Account: xtbtest.Account, Currency: "PLN",
		From: time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC), To: time.Date(2026, time.March, 6, 12, 0, 0, 0, time.UTC),
		Open: [][]xtbtest.Cell{xtbtest.OpenRow(xtbPosition())},
	}
	return &statementFile{name: e.FileName(), content: e.Bytes()}
}

// xtbWeek is a synthetic xStation export of the week after the go-live: a
// deposit of 1000 PLN and a withdrawal of 200 PLN, with the position still
// open.
func xtbWeek() *statementFile {
	at := func(day int) time.Time { return time.Date(2026, time.March, day, 9, 0, 0, 0, time.UTC) }
	e := xtbtest.Export{
		Account: xtbtest.Account, Currency: "PLN", From: at(9), To: at(13),
		Cash: [][]xtbtest.Cell{
			xtbtest.CashRow("Withdrawal", "", "", "", at(11), "-200", "9100000002", "Przelew na rachunek"),
			xtbtest.CashRow("Deposit", "", "", "", at(10), "1000", "9100000001", "Przelew z rachunku"),
		},
		Open: [][]xtbtest.Cell{xtbtest.OpenRow(xtbPosition())},
	}
	return &statementFile{name: e.FileName(), content: e.Bytes()}
}

// xtbPosition is the XTB account's one open position.
func xtbPosition() *xtbtest.Position {
	return &xtbtest.Position{
		Name: "Synthetic Test J", Ticker: xtbTicker, Category: "STOCK", Volume: "20", OpenPrice: "50",
		Opened: time.Date(2026, time.February, 3, 10, 0, 0, 0, time.UTC),
	}
}

// waitPrices syncs the eodhd instance, which prices what the accounts hold
// (D-93), until the read API serves its closes of the three instruments
// from the go-live week: EODHD's, of the eodhd instance, the source each
// listing is valued at.
func (p *dataPath) waitPrices(t *testing.T) {
	t.Helper()
	want := map[string]string{isinG: "105 GBP", isinH: "102 CHF", isinJ: "55 NOK"}
	err := waitFor(t.Context(), dataPathTimeout, func(ctx context.Context) (string, bool, error) {
		if _, err := p.syncOnce(ctx, dpEODHD); err != nil {
			return "", false, fmt.Errorf("%w: %w", errRetry, err)
		}
		var state []string
		for _, isin := range []string{isinG, isinH, isinJ} {
			var prices viewsv1.ListInstrumentPricesResponse
			if err := p.read(ctx, "/v1/instruments/"+isin+"/prices?from=2026-03-02&to="+goLive, &prices); err != nil {
				return "", false, err
			}
			got := describePrices(t, prices.GetPrices())
			if !slices.Equal(got, weekOf(want[isin])) {
				state = append(state, fmt.Sprintf("%s: %v", isin, got))
			}
		}
		return strings.Join(state, "; "), len(state) == 0, nil
	})
	if err != nil {
		t.Fatalf("waiting for EODHD's prices from %s: %v", dpEODHD, err)
	}
}

// weekOf is price on each weekday of the go-live week from the eodhd
// instance, as the primary source, as describePrices writes them.
func weekOf(price string) []string {
	days := []string{"2026-03-02", "2026-03-03", "2026-03-04", "2026-03-05", goLive}
	out := make([]string, 0, len(days))
	for _, d := range days {
		out = append(out, d+" "+price+" "+dpEODHD+" primary")
	}
	return out
}

// describePrices writes each price as its date, close, source and whether
// it is the primary one.
func describePrices(t *testing.T, prices []*viewsv1.InstrumentPrice) []string {
	t.Helper()
	out := make([]string, 0, len(prices))
	for _, pr := range prices {
		s := pr.GetDate() + " " + canonical(t, pr.GetClose().GetAmount()) + " " + pr.GetClose().GetCurrency() + " " + pr.GetSource()
		if pr.GetPrimary() {
			s += " primary"
		}
		out = append(out, s)
	}
	return out
}

// checkDoubleEntry checks D-200's criterion 2: `core check ledger` finds
// the whole log balanced, every version of every transaction of every
// account, the demo's included, summing to zero per currency, and holds the
// data path accounts' versions; and every current transaction of each data
// path account, as the read API serves it, sums to zero per currency.
func (p *dataPath) checkDoubleEntry(t *testing.T) {
	t.Helper()
	r := p.ledger(t, "")
	if !r.Balanced || len(r.Unbalanced) > 0 {
		t.Errorf("core check ledger: balanced %t, unbalanced versions %+v; want a balanced log", r.Balanced, r.Unbalanced)
	}
	for _, s := range r.Sums {
		if !kernelZero(t, s.Amount) {
			t.Errorf("core check ledger: the whole log sums to %s %s, want 0", s.Amount, s.Currency)
		}
	}
	want := map[string]int{p.ibkr: 10, p.xtb: 4, p.bonds: p.bondsBooked.versions}
	for _, a := range r.Accounts {
		if v, ok := want[a.Account]; ok && a.Versions != v {
			t.Errorf("core check ledger: %s holds %d versions, want %d", a.Account, a.Versions, v)
		}
		delete(want, a.Account)
	}
	if len(want) > 0 {
		t.Errorf("core check ledger left out the accounts %v", want)
	}
	for account, current := range map[string]int{p.ibkr: 9, p.xtb: 4, p.bonds: p.bondsBooked.transactions} {
		p.checkPostingsBalance(t, account, current)
	}
}

// checkPostingsBalance checks that account has current transactions in the
// read API, each with postings that sum to zero in every currency.
func (p *dataPath) checkPostingsBalance(t *testing.T, account string, current int) {
	t.Helper()
	txs, err := transactionsPage(t.Context(), p.k, p.token, account, "")
	if err != nil {
		t.Fatalf("transactions of %s: %v", account, err)
	}
	if got := len(txs.GetTransactions()); got != current {
		t.Errorf("the read API lists %d transactions of %s, want %d", got, account, current)
	}
	for _, tx := range txs.GetTransactions() {
		sums := map[string]kernel.Decimal{}
		for _, posting := range tx.GetPostings() {
			sums[posting.GetAmount().GetCurrency()] = sums[posting.GetAmount().GetCurrency()].Add(decimal(t, posting.GetAmount().GetAmount()))
		}
		for currency, sum := range sums {
			if !sum.IsZero() {
				t.Errorf("transaction %s of %s: postings sum to %s %s, want 0", tx.GetExternalId(), account, sum, currency)
			}
		}
	}
}

// checkOpeningLots checks D-200's criterion 4: the IBKR and XTB accounts'
// opening lots on the go-live day, with their open dates, units and costs,
// and that their PLN cost is at NBP's rate of the day before the open date,
// through the net values (checkValues).
func (p *dataPath) checkOpeningLots(t *testing.T) {
	t.Helper()
	tests := []struct {
		account string
		want    []string
	}{
		{account: p.ibkr, want: []string{
			isinH + "/TSTH opened 2026-01-20: 8 units, cost 800 CHF, opening",
			isinG + "/TSTG opened 2026-02-10: 10 units, cost 1000 GBP, opening",
			isinG + "/TSTG opened 2026-02-16: 5 units, cost 520 GBP, opening",
		}},
		{account: p.xtb, want: []string{isinJ + "/" + xtbTicker + " opened 2026-02-03: 20 units, cost 1000 NOK, opening"}},
	}
	for _, tt := range tests {
		var lots viewsv1.ListLotsResponse
		if err := p.read(t.Context(), "/v1/accounts/"+tt.account+"/lots?date="+goLive, &lots); err != nil {
			t.Fatalf("lots of %s: %v", tt.account, err)
		}
		got := make([]string, 0, len(lots.GetLots()))
		for _, l := range lots.GetLots() {
			s := fmt.Sprintf("%s/%s opened %s: %s units, cost %s %s", l.GetInstrument().GetIsin(), l.GetInstrument().GetSymbol(),
				l.GetOpened(), canonical(t, l.GetQuantity()), canonical(t, l.GetCost().GetAmount()), l.GetCost().GetCurrency())
			if l.GetOpening() {
				s += ", opening"
			}
			got = append(got, s)
		}
		slices.Sort(got)
		if diff := cmp.Diff(tt.want, got); diff != "" {
			t.Errorf("lots of %s on %s (-want +got):\n%s", tt.account, goLive, diff)
		}
	}
}

// accountValues are an account's stored values on a date in PLN, as the
// read API serves them: each holding's gross and, where checked, net and
// liquidation value by ISIN, and the day's total, cash and, where checked,
// net and liquidation totals.
type accountValues struct {
	gross, net, liquidation    map[string]string
	total, cash                string
	netTotal, liquidationTotal string
}

// checkValues checks D-200's criteria 4 and 6 on the stored valuations,
// each figure hand-computed from the fixture's synthetic data: closes TSTG
// 105 GBP, TSTH 102 CHF, TSTJ 55 NOK every day; NBP table A mid rates GBP
// 5.0000, CHF 4.5000 and NOK 0.3600 on Friday, January 2, 2026, and 0.0100,
// 0.0100 and 0.0010 more each weekday after it. Each PLN figure is an
// amount times its currency's rate of the valuation day, rounded half up to
// grosze (criterion 6). Net is gross less 19% Belka tax on the gain over
// the PLN cost, each lot's cost at the rate of the day before its open
// date, the latest on or before it (criterion 4):
//
//	IBKR, March 6, GBP 5.45 and CHF 4.95:
//	  TSTG 15 x 105 = 1575 GBP = 8583.75 PLN; cost 1000 GBP x 5.26 (Mon, Feb 9, for Feb 10)
//	    + 520 GBP x 5.30 (Fri, Feb 13, for Mon, Feb 16) = 8016.00; tax 0.19 x 567.75 = 107.87; net 8475.88
//	  TSTH 8 x 102 = 816 CHF = 4039.20 PLN; cost 800 CHF x 4.61 (Mon, Jan 19, for Jan 20) = 3688.00;
//	    tax 0.19 x 351.20 = 66.73; net 3972.47
//	  cash 250 GBP = 1362.50, 100 CHF = 495.00: 1857.50; total 14480.45
//	IBKR, March 13, GBP 5.50 and CHF 5.00, after the week and the correction:
//	  TSTG 17 x 105 = 1785 GBP = 9817.50; TSTH 816 CHF = 4080.00
//	  cash 1138 GBP = 6259.00, 107.80 CHF = 539.00: 6798.00; total 20695.50
//	XTB, March 6, NOK 0.4050:
//	  TSTJ 20 x 55 = 1100 NOK = 445.50 PLN; cost 1000 NOK x 0.3810 (Mon, Feb 2, for Feb 3) = 381.00;
//	    tax 0.19 x 64.50 = 12.26; net 433.24
//	  cash 500.00 PLN; total 945.50
func (p *dataPath) checkValues(t *testing.T) {
	t.Helper()
	tests := []struct {
		account, date string
		want          accountValues
	}{
		{account: p.ibkr, date: goLive, want: accountValues{
			gross: map[string]string{isinG: "8583.75", isinH: "4039.20"},
			net:   map[string]string{isinG: "8475.88", isinH: "3972.47"},
			total: "14480.45", cash: "1857.50",
		}},
		{account: p.ibkr, date: weekEnd, want: accountValues{
			gross: map[string]string{isinG: "9817.50", isinH: "4080.00"},
			total: "20695.50", cash: "6798.00",
		}},
		{account: p.xtb, date: goLive, want: accountValues{
			gross: map[string]string{isinJ: "445.50"},
			net:   map[string]string{isinJ: "433.24"},
			total: "945.50", cash: "500.00",
		}},
	}
	for _, tt := range tests {
		p.waitValues(t, tt.account, tt.date, &tt.want)
	}
}

// waitValues waits until account's stored values on date are want: they
// are revalued in the background after each import and each new price or
// rate (D-36).
func (p *dataPath) waitValues(t *testing.T, account, date string, want *accountValues) {
	t.Helper()
	err := waitFor(t.Context(), dataPathTimeout, func(ctx context.Context) (string, bool, error) {
		var v viewsv1.ListAccountValuationsResponse
		if err := p.read(ctx, "/v1/accounts/"+account+"/valuations?from="+date+"&to="+date, &v); err != nil {
			return "", false, err
		}
		got := valuesOf(t, &v, want)
		return fmt.Sprintf("%+v", got), sameValues(t, &got, want), nil
	})
	if err != nil {
		t.Errorf("the values of %s on %s: %v; want %+v", account, date, err, *want)
	}
}

// valuesOf reads from v the values want names.
func valuesOf(t *testing.T, v *viewsv1.ListAccountValuationsResponse, want *accountValues) accountValues {
	t.Helper()
	got := accountValues{gross: map[string]string{}, net: map[string]string{}, liquidation: map[string]string{}}
	for _, val := range v.GetValuations() {
		isin := val.GetInstrument().GetIsin()
		if _, ok := want.gross[isin]; ok {
			got.gross[isin] = val.GetValueReporting().GetAmount()
		}
		if _, ok := want.net[isin]; ok {
			got.net[isin] = val.GetNetReporting().GetAmount()
		}
		if _, ok := want.liquidation[isin]; ok {
			got.liquidation[isin] = val.GetLiquidationReporting().GetAmount()
		}
	}
	if totals := v.GetTotals(); len(totals) > 0 {
		got.total, got.cash = totals[0].GetValueReporting().GetAmount(), totals[0].GetCashReporting().GetAmount()
		got.netTotal, got.liquidationTotal = totals[0].GetNetReporting().GetAmount(), totals[0].GetLiquidationReporting().GetAmount()
	}
	return got
}

// sameValues reports whether got and want hold the same amounts; the net
// and liquidation totals only where want names them.
func sameValues(t *testing.T, got, want *accountValues) bool {
	t.Helper()
	maps := []struct{ got, want map[string]string }{{got.gross, want.gross}, {got.net, want.net}, {got.liquidation, want.liquidation}}
	for _, m := range maps {
		for isin, amount := range m.want {
			if !sameDecimal(t, m.got[isin], amount) {
				return false
			}
		}
	}
	for _, total := range []struct{ got, want string }{{got.netTotal, want.netTotal}, {got.liquidationTotal, want.liquidationTotal}} {
		if total.want != "" && !sameDecimal(t, total.got, total.want) {
			return false
		}
	}
	return sameDecimal(t, got.total, want.total) && sameDecimal(t, got.cash, want.cash)
}

// checkPriceChecks checks D-200's criterion 6 on the price checks: IBKR's
// marks on the go-live day are compared with EODHD's closes of the same
// day, from the eodhd instance, at the 1% threshold: TSTG's 105.5 GBP is
// 0.5/105 = 0.4762% above 105 GBP and matches, TSTH's 104 CHF is 2/102 =
// 1.9608% above 102 CHF and is flagged; the week's marks equal the closes.
func (p *dataPath) checkPriceChecks(t *testing.T) {
	t.Helper()
	tests := []struct {
		isin string
		want []string
	}{
		{isin: isinG, want: []string{
			"2026-03-06 TSTG: 105.5 GBP from dp-ibkr against 105 GBP of 2026-03-06 from dp-eodhd: 0.004762 over 0.01, matched",
			"2026-03-13 TSTG: 105 GBP from dp-ibkr against 105 GBP of 2026-03-13 from dp-eodhd: 0 over 0.01, matched",
		}},
		{isin: isinH, want: []string{
			"2026-03-06 TSTH: 104 CHF from dp-ibkr against 102 CHF of 2026-03-06 from dp-eodhd: 0.019608 over 0.01, flagged",
			"2026-03-13 TSTH: 102 CHF from dp-ibkr against 102 CHF of 2026-03-13 from dp-eodhd: 0 over 0.01, matched",
		}},
	}
	for _, tt := range tests {
		var checks viewsv1.ListPriceChecksResponse
		if err := p.read(t.Context(), "/v1/price-checks?isin="+tt.isin, &checks); err != nil {
			t.Fatalf("price checks of %s: %v", tt.isin, err)
		}
		got := make([]string, 0, len(checks.GetChecks()))
		for _, c := range checks.GetChecks() {
			status := strings.ToLower(strings.TrimPrefix(c.GetStatus().String(), "PRICE_CHECK_STATUS_"))
			got = append(got, fmt.Sprintf("%s %s: %s %s from %s against %s %s of %s from %s: %s over %s, %s",
				c.GetDate(), c.GetInstrument().GetSymbol(),
				canonical(t, c.GetCrossCheck().GetAmount()), c.GetCrossCheck().GetCurrency(), c.GetCrossCheckSource(),
				canonical(t, c.GetPrice().GetAmount()), c.GetPrice().GetCurrency(), c.GetPriceDate(), c.GetPriceSource(),
				canonical(t, c.GetRelativeDifference()), canonical(t, c.GetThreshold()), status))
		}
		if diff := cmp.Diff(tt.want, got); diff != "" {
			t.Errorf("price checks of %s (-want +got):\n%s", tt.isin, diff)
		}
	}
}

// ledgerCount is how many transactions, versions and superseded versions
// an account holds, as `core check ledger` counts them.
type ledgerCount struct {
	transactions, versions, superseded int
}

// ledgerReport is what `core check ledger` writes (cmd/core/checkledger.go).
type ledgerReport struct {
	Balanced bool `json:"balanced"`
	Accounts []struct {
		Account      string `json:"account"`
		Transactions int    `json:"transactions"`
		Versions     int    `json:"versions"`
		Superseded   int    `json:"superseded"`
	} `json:"accounts"`
	Unbalanced []struct {
		Account    string `json:"account"`
		ExternalID string `json:"externalId"`
		Version    int    `json:"version"`
		Amount     string `json:"amount"`
		Currency   string `json:"currency"`
	} `json:"unbalanced"`
	Sums []struct {
		Amount   string `json:"amount"`
		Currency string `json:"currency"`
	} `json:"sums"`
	Logged []loggedVersion `json:"logged"`
}

// loggedVersion is one stored version `core check ledger` lists.
type loggedVersion struct {
	Source     string `json:"source"`
	ExternalID string `json:"externalId"`
	Version    int    `json:"version"`
	Kind       string `json:"kind"`
	Superseded bool   `json:"superseded"`
}

// countOf returns account's counts in r.
func (r *ledgerReport) countOf(account string) ledgerCount {
	for _, a := range r.Accounts {
		if a.Account == account {
			return ledgerCount{transactions: a.Transactions, versions: a.Versions, superseded: a.Superseded}
		}
	}
	return ledgerCount{}
}

// ledger runs `core check ledger`, with --account when account isn't
// empty, in the core's pod, as the owner runs it with kubectl exec.
func (p *dataPath) ledger(t *testing.T, account string) ledgerReport {
	t.Helper()
	var r ledgerReport
	err := waitFor(t.Context(), syncTimeout, func(ctx context.Context) (string, bool, error) {
		var err error
		r, err = p.ledgerOnce(ctx, account)
		if err != nil {
			return "", false, fmt.Errorf("%w: %w", errRetry, err)
		}
		return "", true, nil
	})
	if err != nil {
		t.Fatalf("core check ledger: %v", err)
	}
	return r
}

// ledgerOnce runs `core check ledger` once.
func (p *dataPath) ledgerOnce(ctx context.Context, account string) (ledgerReport, error) {
	args := []string{"check", "ledger"}
	if account != "" {
		args = append(args, "--account", account)
	}
	out, err := p.core(ctx, args...)
	if err != nil {
		return ledgerReport{}, err
	}
	var r ledgerReport
	if err := json.Unmarshal(out, &r); err != nil {
		return ledgerReport{}, fmt.Errorf("decoding core check ledger's report: %w", err)
	}
	return r, nil
}

// waitLedger waits until account holds want, as `core check ledger`
// counts it, and returns the report.
func (p *dataPath) waitLedger(t *testing.T, account string, want ledgerCount) ledgerReport {
	t.Helper()
	return p.waitCount(t, account, want, func(context.Context) error { return nil })
}

// syncUntil syncs instance until account holds want, and returns the
// report. A sync may fail while the core's scheduler syncs the instance, or
// see the instance's previous statement until the kubelet updates its
// Secret's files; it is tried again.
func (p *dataPath) syncUntil(t *testing.T, instance, account string, want ledgerCount) ledgerReport {
	t.Helper()
	return p.waitCount(t, account, want, func(ctx context.Context) error {
		_, err := p.syncOnce(ctx, instance)
		return err
	})
}

// waitCount runs step, then checks account's counts, until they are want.
func (p *dataPath) waitCount(t *testing.T, account string, want ledgerCount, step func(context.Context) error) ledgerReport {
	t.Helper()
	var r ledgerReport
	err := waitFor(t.Context(), dataPathTimeout, func(ctx context.Context) (string, bool, error) {
		if err := step(ctx); err != nil {
			return "", false, fmt.Errorf("%w: %w", errRetry, err)
		}
		var err error
		if r, err = p.ledgerOnce(ctx, account); err != nil {
			return "", false, fmt.Errorf("%w: %w", errRetry, err)
		}
		got := r.countOf(account)
		return fmt.Sprintf("%s holds %+v", account, got), got == want, nil
	})
	if err != nil {
		t.Fatalf("waiting for %s to hold %+v: %v", account, want, err)
	}
	return r
}

// sync syncs instance with the core's CLI until a sync succeeds.
func (p *dataPath) sync(t *testing.T, instance string) {
	t.Helper()
	err := waitFor(t.Context(), syncTimeout, func(ctx context.Context) (string, bool, error) {
		out, err := p.syncOnce(ctx, instance)
		if err != nil {
			return "", false, fmt.Errorf("%w: %w", errRetry, err)
		}
		return out, true, nil
	})
	if err != nil {
		t.Fatalf("core connector sync --instance %s: %v", instance, err)
	}
}

// syncOnce runs `core connector sync --instance <instance>` once in the
// core's pod and returns what it printed: the batch's counts.
func (p *dataPath) syncOnce(ctx context.Context, instance string) (string, error) {
	out, err := p.core(ctx, "connector", "sync", "--instance", instance)
	return string(out), err
}

// core runs the core's own CLI with args in a ready core pod.
func (p *dataPath) core(ctx context.Context, args ...string) ([]byte, error) {
	pod, err := p.k.corePod(ctx)
	if err != nil {
		return nil, err
	}
	return p.k.exec(ctx, namespace, pod, "core", nil, append([]string{"/core"}, args...)...)
}

// setQuery points the IBKR instance's flex-token Secret at the Flex query
// query, as the owner's statement moves on; the plugin reads the Secret's
// files on each sync.
func (p *dataPath) setQuery(t *testing.T, query string) {
	t.Helper()
	secrets := p.k.clients.CoreV1().Secrets(namespace)
	s, err := secrets.Get(t.Context(), dpIBKR+"--"+flexRole, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("getting %s's Secret: %v", dpIBKR, err)
	}
	s.Data[flexQueryFile] = []byte(query)
	if _, err := secrets.Update(t.Context(), s, metav1.UpdateOptions{FieldManager: fieldManager}); err != nil {
		t.Fatalf("updating %s's Secret: %v", dpIBKR, err)
	}
}

// read GETs path from the read API with the script token and decodes it
// into m.
func (p *dataPath) read(ctx context.Context, path string, m proto.Message) error {
	r, err := call(ctx, p.k, clientPod, path, p.token)
	if err != nil {
		return fmt.Errorf("%w: %w", errRetry, err)
	}
	if r.status != http.StatusOK {
		return fmt.Errorf("%w: GET %s: status %d, body %q", errRetry, path, r.status, r.body)
	}
	if err := (protojson.UnmarshalOptions{DiscardUnknown: true}).Unmarshal(r.body, m); err != nil {
		return fmt.Errorf("decoding GET %s: %w", path, err)
	}
	return nil
}

// firstSleeve returns the ID of the first sleeve the read API lists, which
// the data path's accounts go in.
func firstSleeve(t *testing.T, k *kube, token string) string {
	t.Helper()
	p := &dataPath{k: k, token: token}
	var sleeves viewsv1.ListSleevesResponse
	err := waitFor(t.Context(), coreTimeout, func(ctx context.Context) (string, bool, error) {
		if err := p.read(ctx, "/v1/sleeves", &sleeves); err != nil {
			return "", false, err
		}
		return "no sleeve", len(sleeves.GetSleeves()) > 0, nil
	})
	if err != nil {
		t.Fatalf("listing sleeves: %v", err)
	}
	return sleeves.GetSleeves()[0].GetId()
}

// accountID is the ID the core gives an account named name.
func accountID(t *testing.T, name string) string {
	t.Helper()
	id, err := portfolio.KeyFromName(name)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

// decimal parses s, failing t on an invalid decimal.
func decimal(t *testing.T, s string) kernel.Decimal {
	t.Helper()
	d, err := kernel.ParseDecimal(s)
	if err != nil {
		t.Fatalf("decimal %q: %v", s, err)
	}
	return d
}

// canonical writes the decimal s without trailing zeros, so "105.00" and
// "105" read alike; anything else stays as it is.
func canonical(t *testing.T, s string) string {
	t.Helper()
	d, err := kernel.ParseDecimal(s)
	if err != nil {
		return s
	}
	return d.String()
}

// sameDecimal reports whether got and want are the same decimal; an empty
// or invalid got is not.
func sameDecimal(t *testing.T, got, want string) bool {
	t.Helper()
	g, err := kernel.ParseDecimal(got)
	return err == nil && g.Equal(decimal(t, want))
}

// kernelZero reports whether the decimal s is zero.
func kernelZero(t *testing.T, s string) bool {
	t.Helper()
	return decimal(t, s).IsZero()
}
