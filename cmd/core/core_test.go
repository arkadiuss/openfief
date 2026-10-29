package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/arkadiuss/openfief/internal/demo"
	"github.com/arkadiuss/openfief/internal/jurisdiction"
	"github.com/arkadiuss/openfief/internal/kernel"
	ledgerapp "github.com/arkadiuss/openfief/internal/ledger/app"
	ledgerpg "github.com/arkadiuss/openfief/internal/ledger/postgres"
	"github.com/arkadiuss/openfief/internal/valuation"
	valuationapp "github.com/arkadiuss/openfief/internal/valuation/app"
	valuationpg "github.com/arkadiuss/openfief/internal/valuation/postgres"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// dsnVar names the server the tests create their databases on; task
// test:pg sets it to a disposable server, and without it the tests skip.
const dsnVar = "OPENFIEF_PG_DSN"

// demoSeed seeds the demo dataset the test loads.
const demoSeed = 20261028

// emptyDatabase returns the connection string of a fresh, empty database,
// dropped when the test ends.
func emptyDatabase(t *testing.T) string {
	t.Helper()
	dsn := os.Getenv(dsnVar)
	if dsn == "" {
		t.Skip(dsnVar + " is not set: task test:pg runs this test on a disposable PostgreSQL")
	}
	ctx := t.Context()
	sum := sha256.Sum256([]byte(t.Name()))
	name := "core_" + hex.EncodeToString(sum[:8]) + "_" + strconv.Itoa(os.Getpid())
	adminExec(ctx, t, dsn, "DROP DATABASE IF EXISTS "+name+" WITH (FORCE)")
	adminExec(ctx, t, dsn, "CREATE DATABASE "+name)
	t.Cleanup(func() { adminExec(context.WithoutCancel(ctx), t, dsn, "DROP DATABASE IF EXISTS "+name+" WITH (FORCE)") })
	return withDatabase(t, dsn, name)
}

// withDatabase returns dsn, a URL or key=value connection string, pointing
// at the database name instead.
func withDatabase(t *testing.T, dsn, name string) string {
	t.Helper()
	if !strings.HasPrefix(dsn, "postgres://") && !strings.HasPrefix(dsn, "postgresql://") {
		return dsn + " dbname=" + name // a later key wins
	}
	u, err := url.Parse(dsn)
	if err != nil {
		t.Fatalf("parsing %s: %v", dsnVar, err)
	}
	u.Path = "/" + name
	return u.String()
}

// adminExec runs sql on the server's own database.
func adminExec(ctx context.Context, t *testing.T, dsn, sql string) {
	t.Helper()
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatalf("connecting to %s: %v", dsnVar, err)
	}
	defer func() {
		if err := conn.Close(ctx); err != nil {
			t.Errorf("closing: %v", err)
		}
	}()
	if _, err := conn.Exec(ctx, sql); err != nil {
		t.Fatalf("%s: %v", sql, err)
	}
}

// quiet is a logger that drops everything.
func quiet() *slog.Logger { return slog.New(slog.DiscardHandler) }

// coreNow is the core's clock in the test: noon in Warsaw on the dataset's
// today.
func coreNow() time.Time { return time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC) }

// tokenCmd runs a token command as the CLI does and returns its output.
func tokenCmd(t *testing.T, dsn string, args ...string) string {
	t.Helper()
	cmd, err := parseArgs(append([]string{"token"}, args...), func(string) string { return dsn })
	if err != nil {
		t.Fatalf("parseArgs: %v", err)
	}
	var out bytes.Buffer
	if err := run(t.Context(), &cmd, &out, quiet()); err != nil {
		t.Fatalf("core token %s: %v", strings.Join(args, " "), err)
	}
	return out.String()
}

// startCore wires and starts the core over dsn, as serve does, and returns
// it with its HTTP server; both stop when the test ends.
func startCore(t *testing.T, dsn string) (*core, *httptest.Server, *pgxpool.Pool) {
	t.Helper()
	return startCoreWith(t, dsn, func(*serveConfig) {})
}

// startCoreWith is startCore with the serve configuration edited by edit.
func startCoreWith(t *testing.T, dsn string, edit func(cfg *serveConfig)) (*core, *httptest.Server, *pgxpool.Pool) {
	t.Helper()
	return startCoreIn(t, dsn, nil, edit)
}

// startCoreIn is startCoreWith in cluster cl, or outside a cluster when cl
// is nil.
func startCoreIn(t *testing.T, dsn string, cl *cluster, edit func(cfg *serveConfig)) (*core, *httptest.Server, *pgxpool.Pool) {
	t.Helper()
	ctx := t.Context()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("connecting: %v", err)
	}
	t.Cleanup(pool.Close)
	cfg := defaultServe(t)
	cfg.Workers = 3
	edit(&cfg)
	c, err := start(ctx, pool, &cfg, coreNow, quiet(), cl)
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	s := &server{pool: pool}
	s.api.Store(&c.api)
	srv := httptest.NewServer(s.handler())
	t.Cleanup(func() {
		srv.Close()
		if err := shutdown(context.WithoutCancel(ctx), s, srv.Config, c, 10*time.Second); err != nil {
			t.Errorf("shutdown: %v", err)
		}
	})
	return c, srv, pool
}

// loadDemo loads part of the demo dataset as a separate process would, the
// demo loader (cmd/demo-seed) say: the registry, the XTB and bond accounts
// through a ledger of its own whose import hook enqueues revaluations
// (D-36), and their valuation inputs. The IBKR account stays empty. FX rates
// are synthetic, one flat rate per currency from before go-live. The core's
// workers then value the accounts.
func loadDemo(t *testing.T, pool *pgxpool.Pool) *demo.Dataset {
	t.Helper()
	ctx := t.Context()
	today, err := kernel.ParseDate("2026-10-28")
	if err != nil {
		t.Fatalf("ParseDate: %v", err)
	}
	ds, err := demo.Generate(demoSeed, today)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	vstore, err := valuationpg.NewStore(pool)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	store := ledgerpg.New(pool)
	l := ledgerapp.New(store, vstore.ImportHook())
	saveRegistry(t, l, &ds)
	for _, b := range []*demo.Batch{&ds.XTBOpening, &ds.BondsOpening} {
		if _, err := l.Seed(ctx, b.Transactions, b.Assertions); err != nil {
			t.Fatalf("Seed: %v", err)
		}
	}
	for _, b := range []*demo.Batch{&ds.XTBActivity, &ds.BondsActivity} {
		if _, err := l.Import(ctx, b.Transactions, b.Assertions); err != nil {
			t.Fatalf("Import: %v", err)
		}
	}
	recordInputs(t, valuationapp.NewInputRecorder(valuationapp.NewPlanner(l, store), vstore, jurisdictionPack()), &ds)
	return &ds
}

// saveRegistry saves ds's sleeves and accounts.
func saveRegistry(t *testing.T, l *ledgerapp.Ledger, ds *demo.Dataset) {
	t.Helper()
	for _, sl := range ds.Sleeves {
		if err := l.SaveSleeve(t.Context(), sl); err != nil {
			t.Fatalf("SaveSleeve: %v", err)
		}
	}
	for _, a := range ds.Accounts {
		if err := l.SaveAccount(t.Context(), a); err != nil {
			t.Fatalf("SaveAccount: %v", err)
		}
	}
}

// recordInputs records ds's valuation inputs and one synthetic FX rate per
// foreign currency, dated before go-live.
func recordInputs(t *testing.T, r *valuationapp.InputRecorder, ds *demo.Dataset) {
	t.Helper()
	ctx := t.Context()
	if err := r.RecordTaxWrapper(ctx, demo.BondsAccount, jurisdiction.Wrapper(ds.TaxWrapper.String())); err != nil {
		t.Fatalf("RecordTaxWrapper: %v", err)
	}
	for i := range ds.Bonds {
		if err := r.RecordBondTerms(ctx, ds.Bonds[i].Instrument.ISIN, &ds.Bonds[i].Terms); err != nil {
			t.Fatalf("RecordBondTerms: %v", err)
		}
	}
	rates := make([]valuation.FXRate, 0, 2)
	for _, code := range []string{"EUR", "USD"} {
		c, err := kernel.ParseCurrency(code)
		if err != nil {
			t.Fatalf("ParseCurrency: %v", err)
		}
		rate, err := kernel.ParseDecimal("4.0000")
		if err != nil {
			t.Fatalf("ParseDecimal: %v", err)
		}
		rates = append(rates, valuation.FXRate{Currency: c, Date: ds.GoLive.AddDays(-14), Rate: rate, Source: "synthetic"})
	}
	if err := errors.Join(
		r.RecordIndexReadings(ctx, ds.CPI), r.RecordPrices(ctx, ds.Prices), r.RecordFXRates(ctx, rates),
	); err != nil {
		t.Fatalf("recording inputs: %v", err)
	}
}

// reply is an HTTP response, read.
type reply struct {
	status int
	header http.Header
	body   []byte
}

// get sends GET path with token, if any, and returns the response.
func get(t *testing.T, srv *httptest.Server, path, token string) reply {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, srv.URL+path, http.NoBody)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	body, readErr := io.ReadAll(resp.Body)
	if closeErr := resp.Body.Close(); readErr != nil || closeErr != nil {
		t.Fatalf("reading %s: %v %v", path, readErr, closeErr)
	}
	return reply{status: resp.StatusCode, header: resp.Header, body: body}
}

// getJSON GETs path with token, wants 200, and decodes the body into msg,
// refusing unknown fields: the REST shape must be the protobuf message's.
func getJSON(t *testing.T, srv *httptest.Server, path, token string, msg proto.Message) {
	t.Helper()
	r := get(t, srv, path, token)
	if r.status != http.StatusOK {
		t.Fatalf("GET %s = %d, want 200", path, r.status)
	}
	if err := protojson.Unmarshal(r.body, msg); err != nil {
		t.Fatalf("GET %s: decoding %T: %v", path, msg, err)
	}
}

// waitValued polls GET /v1/valuation-status until the XTB and bond accounts
// are valued through today with nothing pending or running.
func waitValued(t *testing.T, srv *httptest.Server, token string, ds *demo.Dataset) *viewsv1.ListValuationStatusesResponse {
	t.Helper()
	deadline := time.Now().Add(2 * time.Minute)
	loaded := map[string]bool{string(demo.XTBAccount): true, string(demo.BondsAccount): true}
	for {
		var got viewsv1.ListValuationStatusesResponse
		getJSON(t, srv, "/v1/valuation-status", token, &got)
		valued := 0
		for _, s := range got.GetStatuses() {
			if loaded[s.GetAccountId()] && !s.GetPending() && !s.GetRunning() && s.GetValuedThrough() == ds.Today.String() {
				valued++
			}
		}
		if valued == len(loaded) {
			return &got
		}
		if time.Now().After(deadline) {
			t.Fatalf("not valued through %s in time: %v", ds.Today, &got)
		}
		time.Sleep(200 * time.Millisecond)
	}
}

// checkPerformance checks that GET /v1/performance over the valued demo
// data is a series in PLN that starts at 100 and reaches today with no gap.
func checkPerformance(t *testing.T, srv *httptest.Server, token string, ds *demo.Dataset) {
	t.Helper()
	var perf viewsv1.GetPerformanceResponse
	getJSON(t, srv, "/v1/performance", token, &perf)
	days := perf.GetDays()
	if len(days) == 0 || days[0].GetIndex() != "100" || days[len(days)-1].GetDate() != ds.Today.String() {
		t.Errorf("GET /v1/performance days = %v, want a series from 100 through %s", days, ds.Today)
	}
	if perf.GetGap() != viewsv1.PerformanceGap_PERFORMANCE_GAP_UNSPECIFIED || perf.GetReportingCurrency() != "PLN" {
		t.Errorf("GET /v1/performance = gap %v in %q, want no gap, in PLN", perf.GetGap(), perf.GetReportingCurrency())
	}
}

func TestCoreServesTheReadAPI(t *testing.T) {
	t.Parallel()
	dsn := emptyDatabase(t)
	_, srv, pool := startCore(t, dsn)
	ds := loadDemo(t, pool)
	issued := tokenCmd(t, dsn, "issue", "--name", "integration-test", "--ttl", "1h")
	secret, _, _ := strings.Cut(issued, "\n")
	if !strings.HasPrefix(secret, "ofs_") {
		t.Fatalf("token issue printed %q first, want the secret", secret)
	}

	for _, path := range []string{"/healthz", "/readyz"} {
		if r := get(t, srv, path, ""); r.status != http.StatusOK {
			t.Errorf("GET %s = %d, want 200", path, r.status)
		}
	}
	var accounts viewsv1.ListAccountsResponse
	getJSON(t, srv, "/v1/accounts", secret, &accounts)
	ids := make([]string, 0, len(accounts.GetAccounts()))
	for _, a := range accounts.GetAccounts() {
		ids = append(ids, a.GetId())
	}
	if want := "demo-bonds-ike demo-ibkr demo-xtb"; strings.Join(ids, " ") != want {
		t.Errorf("GET /v1/accounts ids = %v, want %s", ids, want)
	}
	var holdings viewsv1.ListHoldingsResponse
	getJSON(t, srv, "/v1/accounts/"+string(demo.XTBAccount)+"/holdings?date="+ds.Today.String(), secret, &holdings)
	if len(holdings.GetHoldings()) == 0 || holdings.GetDate() != ds.Today.String() {
		t.Errorf("GET holdings = %v, want the XTB account's holdings on %s", &holdings, ds.Today)
	}
	waitValued(t, srv, secret, ds)
	checkPerformance(t, srv, secret, ds)

	checkRefused(t, srv, "", "no token")
	checkRefused(t, srv, "ofs_0123456789abcdef_"+strings.Repeat("A", 43), "an unknown token")
	listed := tokenCmd(t, dsn, "list")
	id := strings.Fields(strings.Split(listed, "\n")[1])[0]
	tokenCmd(t, dsn, "revoke", id)
	checkRefused(t, srv, secret, "a revoked token")
	if strings.Contains(listed, secret) {
		t.Error("token list shows the secret")
	}
}

// checkRefused checks that GET /v1/accounts with token is 401 with a Bearer
// challenge and no data.
func checkRefused(t *testing.T, srv *httptest.Server, token, what string) {
	t.Helper()
	r := get(t, srv, "/v1/accounts", token)
	if r.status != http.StatusUnauthorized || !strings.HasPrefix(r.header.Get("WWW-Authenticate"), "Bearer") {
		t.Errorf("GET /v1/accounts with %s = %d %q, want 401 with a Bearer challenge", what, r.status, r.header.Get("WWW-Authenticate"))
	}
	if bytes.Contains(r.body, []byte("demo-")) {
		t.Errorf("GET /v1/accounts with %s returned data: %s", what, r.body)
	}
}
