//go:build e2e

package e2e

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arkadiuss/openfief/internal/ingest/treasurybonds"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// The data path's treasury bond series, as the fixture's synthetic Ministry
// of Finance spreadsheet lists them (test/e2e/fixture/bonds.go), with their
// registry ISINs (D-214), and the two days their values are checked on.
const (
	seriesTOS = "TOS0329"
	seriesEDO = "EDO0534"
	seriesCOI = "COI1029"
	isinTOS   = "PL9999001016"
	isinEDO   = "PL9999001024"
	isinCOI   = "PL9999001032"
	// bondsSpring is in every series' first period but EDO's second, the
	// first set by CPI; bondsAutumn in EDO's third and COI's second, both
	// set by CPI, after COI's first coupon.
	bondsSpring = "2026-03-31"
	bondsAutumn = "2026-11-02"
	// ministryControl publishes the fixture's ministry data (POST).
	ministryControl = "/control/ministry"
	// warsawZone dates the treasury-bonds connector's today.
	warsawZone = "Europe/Warsaw"
)

// bondLotForm is a bond lot as the owner enters it in the UI's form, its
// rates in percent; empty rates and fee leave them to the connector (D-97).
func bondLotForm(series, quantity, purchased, first, margin, fee string) map[string]string {
	return map[string]string{
		"series": series, "quantity": quantity, "purchase_date": purchased, "purchase": "cash",
		"first_rate": first, "margin": margin, "fee": fee,
	}
}

// bookBonds has the owner create a treasury bond account bound to the
// treasury-bonds connector and, while the fixture withholds the ministry's
// data, enter a lot of TOS0329 and one of EDO0534 with their terms, each
// booking a deposit and a buy on its purchase date (D-98) under its series'
// provisional ISIN (D-214); entering the TOS lot again changes nothing.
func (p *dataPath) bookBonds(t *testing.T) {
	t.Helper()
	p.owner.postForm(t, "/accounts", map[string]string{
		"name": bondsAccountName, "institution": "treasury-bonds", "sleeve": p.sleeve, "instance": dpBonds,
	})
	tos := bondLotForm(seriesTOS, "10", "2026-03-10", "5.65", "0", "0.7")
	p.owner.postForm(t, "/accounts/"+p.bonds+"/bond-lots", tos)
	before := p.waitLedger(t, p.bonds, ledgerCount{transactions: 2, versions: 2})
	p.owner.postForm(t, "/accounts/"+p.bonds+"/bond-lots", tos)
	if again := p.ledger(t, p.bonds); !cmp.Equal(before.Logged, again.Logged) {
		t.Errorf("entering the same bond lot again changed the log (-before +after):\n%s", cmp.Diff(before.Logged, again.Logged))
	}
	p.owner.postForm(t, "/accounts/"+p.bonds+"/bond-lots", bondLotForm(seriesEDO, "20", "2024-05-14", "6.5", "1.75", "2"))
	p.waitLedger(t, p.bonds, ledgerCount{transactions: 4, versions: 4})
	p.checkBondISINs(t, map[string]bool{seriesTOS: false, seriesEDO: false})
}

// rekeyBonds publishes the fixture's ministry data and syncs the
// treasury-bonds instance until the core re-keyed both series to the
// registry's ISINs the data lists (D-214): each buy gets a second version
// naming the registry's ISIN, which supersedes the first. Then the owner
// enters a lot of COI1029 leaving its terms to the connector, which books it
// under its registry ISIN with the published rates and the fee table's fee
// (D-97), and syncs until COI's coupons due by today are booked as interest
// (D-101, D-143).
func (p *dataPath) rekeyBonds(t *testing.T) {
	t.Helper()
	p.publishMinistry(t)
	p.syncUntil(t, dpBonds, p.bonds, ledgerCount{transactions: 4, versions: 6, superseded: 2})
	p.checkBondISINs(t, map[string]bool{seriesTOS: true, seriesEDO: true})
	p.owner.postForm(t, "/accounts/"+p.bonds+"/bond-lots", bondLotForm(seriesCOI, "30", "2025-10-06", "", "", ""))
	coupons := coiCouponsDue(t)
	p.bondsBooked = ledgerCount{transactions: 6 + coupons, versions: 8 + coupons, superseded: 2}
	p.syncUntil(t, dpBonds, p.bonds, p.bondsBooked)
	p.checkBondISINs(t, map[string]bool{seriesTOS: true, seriesEDO: true, seriesCOI: true})
}

// coiCouponsDue is how many of COI1029's coupons the connector books by
// today in Warsaw: the first period's, on 2026-10-06, and the second's, on
// 2027-10-06, whose rate takes the August 2026 CPI the fixture serves. The
// third's takes August 2027's, which neither the fixture nor the demo
// holds, so it waits (treasurybonds.BookCoupons), as maturity does.
func coiCouponsDue(t *testing.T) int {
	t.Helper()
	loc, err := time.LoadLocation(warsawZone)
	if err != nil {
		t.Fatal(err)
	}
	today := time.Now().In(loc).Format(time.DateOnly)
	n := 0
	for _, due := range []string{"2026-10-06", "2027-10-06"} {
		if due <= today {
			n++
		}
	}
	return n
}

// publishMinistry publishes the fixture's ministry data, which it withholds
// until then, through a port-forward to the fixture's pod over HTTPS,
// trusting the run's CA.
func (p *dataPath) publishMinistry(t *testing.T) {
	t.Helper()
	addr := p.k.forwardPod(t, fixtureNamespace, podPort{pod: fixturePod(t, p.k), port: fixturePodPort})
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(p.fixtureCA) {
		t.Fatal("the run's CA is no PEM certificate")
	}
	client := &http.Client{Timeout: callTimeout, Transport: &http.Transport{
		TLSClientConfig: &tls.Config{RootCAs: roots, ServerName: fixtureHost, MinVersion: tls.VersionTLS12},
	}}
	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, "https://"+addr+ministryControl, http.NoBody)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("publishing the fixture's ministry data: %v", err)
	}
	if err := resp.Body.Close(); err != nil || resp.StatusCode != http.StatusNoContent {
		t.Fatalf("publishing the fixture's ministry data: status %d (%v), want 204", resp.StatusCode, err)
	}
}

// fixturePod returns the name of the fixture's running pod: it has one
// replica.
func fixturePod(t *testing.T, k *kube) string {
	t.Helper()
	pods, err := k.clients.CoreV1().Pods(fixtureNamespace).List(t.Context(), metav1.ListOptions{LabelSelector: "app.kubernetes.io/name=" + fixtureService})
	if err != nil {
		t.Fatalf("listing the fixture's pods: %v", err)
	}
	for i := range pods.Items {
		if pods.Items[i].Status.Phase == corev1.PodRunning && pods.Items[i].DeletionTimestamp == nil {
			return pods.Items[i].Name
		}
	}
	t.Fatalf("no running fixture pod among %d", len(pods.Items))
	return ""
}

// checkBondISINs checks the bond account's open lots on bondsAutumn: one
// per series of want, under its registry ISIN when want says so, else under its
// provisional one, which is never the registry's.
func (p *dataPath) checkBondISINs(t *testing.T, want map[string]bool) {
	t.Helper()
	registry := map[string]string{seriesTOS: isinTOS, seriesEDO: isinEDO, seriesCOI: isinCOI}
	var lots viewsv1.ListLotsResponse
	if err := p.read(t.Context(), "/v1/accounts/"+p.bonds+"/lots?date="+bondsAutumn, &lots); err != nil {
		t.Fatalf("lots of %s: %v", p.bonds, err)
	}
	got := map[string]bool{}
	for _, l := range lots.GetLots() {
		in := l.GetInstrument()
		got[in.GetSymbol()] = in.GetIsin() == registry[in.GetSymbol()]
		if !got[in.GetSymbol()] && !strings.HasPrefix(in.GetIsin(), "PL"+in.GetSymbol()+"00") {
			t.Errorf("lot of %s under %s: neither its registry ISIN nor its provisional one", in.GetSymbol(), in.GetIsin())
		}
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("bond lots under their registry ISIN (true) or provisional one (false) (-want +got):\n%s", diff)
	}
}

// checkBonds checks D-200's criterion 5 (SCEN-e2e-bond-values-reference,
// SCEN-e2e-bond-net-if-redeemed, SCEN-e2e-bond-registry-isin-and-terms):
// the stored values of each series on two days equal the reference values,
// and the series' terms are shown with where they came from.
func (p *dataPath) checkBonds(t *testing.T) {
	t.Helper()
	p.checkBondValues(t)
	p.checkBondTerms(t)
}

// checkBondValues checks each series' gross accrued value against the
// reference valuer, github.com/maciekmm/obligacje (MIT) at commit 8e1c495,
// run offline on the same terms, and against the hand calculation from the
// issue letters' formulas (pl.ValueLot's doc); and each series' net and
// liquidation values, and the account's three totals, against
// the hand calculation. The account is taxable; every rate is the fixture's
// synthetic one, every CPI the demo dataset's (GUS, March 2025 104.9, March
// 2026 103.0, August 2026 103.4). Per bond, N = 100:
//
//	capitalized (TOS, EDO): WP = N × Π(1 + r_i) × (1 + r_k × a / ACT), rounded half up to grosze
//	coupon (COI): WP = N + N × r_k × a / ACT, the interest rounded half up to grosze
//	net = gross − 19% of the interest (rounded half up to grosze)
//	liquidation = gross − fee − 19% of (interest − fee), the fee b per bond
//	  capped at the accrued interest, except COI after its first period
//
//	2026-03-31:
//	  TOS0329 10 × 100 × (1 + 0.0565 × 21/365) = 100.33 → 1003.30; tax 0.63, net 1002.67;
//	    fee 0.70 capped at 0.33 → 3.30, liquidation 1000.00
//	  EDO0534 20 × 100 × 1.065 × (1 + (0.049 + 0.0175) × 321/365) = 112.73 → 2254.60; tax 48.37,
//	    net 2206.23; fee 40.00, tax 0.19 × 214.60 = 40.77, liquidation 2173.83
//	  COI1029 30 × (100 + 100 × 0.056 × 176/365 = 2.70) → 3081.00; tax 15.39, net 3065.61;
//	    fee 60.00, tax 3.99, liquidation 3017.01
//	  cash 0; totals 6338.90, net 6274.51, liquidation 6190.84
//	2026-11-02:
//	  TOS0329 100 × (1 + 0.0565 × 237/365) = 103.67 → 1036.70; tax 6.97, net 1029.73;
//	    fee 7.00, tax 5.64, liquidation 1024.06
//	  EDO0534 100 × 1.065 × 1.0665 × (1 + (0.030 + 0.0175) × 172/365) = 116.12 → 2322.40;
//	    tax 61.26, net 2261.14; fee 40.00, tax 53.66, liquidation 2228.74
//	  COI1029 100 + 100 × (0.034 + 0.015) × 27/365 = 100.36 → 3010.80; tax 2.05, net 3008.75;
//	    full fee 60.00, no tax, liquidation 2950.80
//	  cash: COI's first coupon 30 × 5.60 = 168.00 less 31.92 tax = 136.08;
//	    totals 6505.98, net 6435.70, liquidation 6339.68
func (p *dataPath) checkBondValues(t *testing.T) {
	t.Helper()
	tests := []struct {
		date string
		want accountValues
	}{
		{date: bondsSpring, want: accountValues{
			gross:       map[string]string{isinTOS: "1003.30", isinEDO: "2254.60", isinCOI: "3081.00"},
			net:         map[string]string{isinTOS: "1002.67", isinEDO: "2206.23", isinCOI: "3065.61"},
			liquidation: map[string]string{isinTOS: "1000.00", isinEDO: "2173.83", isinCOI: "3017.01"},
			total:       "6338.90", netTotal: "6274.51", liquidationTotal: "6190.84", cash: "0",
		}},
		{date: bondsAutumn, want: accountValues{
			gross:       map[string]string{isinTOS: "1036.70", isinEDO: "2322.40", isinCOI: "3010.80"},
			net:         map[string]string{isinTOS: "1029.73", isinEDO: "2261.14", isinCOI: "3008.75"},
			liquidation: map[string]string{isinTOS: "1024.06", isinEDO: "2228.74", isinCOI: "2950.80"},
			total:       "6505.98", netTotal: "6435.70", liquidationTotal: "6339.68", cash: "136.08",
		}},
	}
	for _, tt := range tests {
		p.waitValues(t, p.bonds, tt.date, &tt.want)
	}
}

// checkBondTerms checks the series' terms the read API shows under their
// registry ISINs: COI1029's published rates, with the attribution of the
// fixture's gov.pl file, and the fee table's fee; TOS0329's and EDO0534's
// as the owner entered them, moved over by the re-key.
func (p *dataPath) checkBondTerms(t *testing.T) {
	t.Helper()
	file := "https://" + fixtureHost + "/gov/attachment/0e2e0000-0000-4000-8000-000000000805"
	want := []string{
		seriesTOS + ": 0.0565 + 0, entered; fee 0.7 PLN, entered",
		seriesEDO + ": 0.065 + 0.0175, entered; fee 2 PLN, entered",
		seriesCOI + ": 0.056 + 0.015, published; fee 2 PLN, table; " + treasurybonds.Publisher + ", " +
			treasurybonds.DatasetTitle + ", " + treasurybonds.License + ", " + file + ", made 2026-11-02",
	}
	isins := []string{isinTOS, isinEDO, isinCOI}
	got := make([]string, 0, len(isins))
	for _, isin := range isins {
		var r viewsv1.GetInstrumentResponse
		if err := p.read(t.Context(), "/v1/instruments/"+isin, &r); err != nil {
			t.Fatalf("instrument %s: %v", isin, err)
		}
		bt := r.GetInstrument().GetBondTerms()
		s := fmt.Sprintf("%s: %s + %s, %s; fee %s %s, %s", bt.GetSeries(), canonical(t, bt.GetFirstPeriodRate()),
			canonical(t, bt.GetMargin()), bt.GetRatesSource(), canonical(t, bt.GetEarlyRedemptionFee().GetAmount()),
			bt.GetEarlyRedemptionFee().GetCurrency(), bt.GetFeeSource())
		if a := bt.GetAttribution(); a != nil {
			s += fmt.Sprintf("; %s, %s, %s, %s, made %s", a.GetPublisher(), a.GetDataset(), a.GetLicense(), a.GetUrl(), a.GetMade())
			if !strings.HasPrefix(a.GetProcessing(), treasurybonds.PublishedProcessing) || a.GetFetched() == "" {
				t.Errorf("%s's attribution: processing %q, fetched %q; want the reader's processing and a fetch time", isin, a.GetProcessing(), a.GetFetched())
			}
		}
		got = append(got, s)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("bond series terms (-want +got):\n%s", diff)
	}
}
