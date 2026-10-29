package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/jackc/pgx/v5/pgxpool"
	authnv1 "k8s.io/api/authentication/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"

	"github.com/arkadiuss/openfief/internal/demo"
	"github.com/arkadiuss/openfief/internal/ingest"
	ingestkube "github.com/arkadiuss/openfief/internal/ingest/kube"
	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/valuation"
	connectorv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1"
	"github.com/arkadiuss/openfief/sdk/go/gen/openfief/connector/v1/connectorv1connect"
)

// coreToken is the synthetic token the fake API server issues to the core.
const coreToken = "synthetic-core-token"

// errNotCore is what the fake connector answers a call without the core's
// token.
var errNotCore = errors.New("not the core's token")

// depositConnector serves the connector contract: every Sync delivers one
// deposit into account, after checking the core's bearer token.
type depositConnector struct {
	connectorv1connect.UnimplementedConnectorServiceHandler

	account string
}

func (c depositConnector) Sync(_ context.Context, req *connect.Request[connectorv1.SyncRequest]) (*connect.Response[connectorv1.SyncResponse], error) {
	if req.Header().Get("Authorization") != "Bearer "+coreToken {
		return nil, fmt.Errorf("sync: %w", connect.NewError(connect.CodeUnauthenticated, errNotCore))
	}
	cash := &connectorv1.Book{Kind: connectorv1.BookKind_BOOK_KIND_CASH, Currency: "USD"}
	equity := &connectorv1.Book{Kind: connectorv1.BookKind_BOOK_KIND_EQUITY, Currency: "USD"}
	return connect.NewResponse(&connectorv1.SyncResponse{
		Watermark: "after-" + req.Msg.GetWatermark(),
		Delivery: &connectorv1.Delivery{
			BatchId: req.Msg.GetBatchId(), Institution: "ibkr",
			Transactions: []*connectorv1.Transaction{{
				Key:       &connectorv1.TransactionKey{Source: "cash", ExternalId: "D1"},
				AccountId: c.account, Kind: connectorv1.TransactionKind_TRANSACTION_KIND_DEPOSIT, Date: "2026-10-27",
				Postings: []*connectorv1.Posting{
					{Book: cash, Amount: &connectorv1.Money{Amount: "100.00", Currency: "USD"}},
					{Book: equity, Amount: &connectorv1.Money{Amount: "-100.00", Currency: "USD"}},
				},
			}},
			FxRates: []*connectorv1.FxRate{{Currency: "USD", Date: "2026-10-27", Rate: "4.0000"}},
		},
	}), nil
}

// serveConnector serves conn over HTTP/2 without TLS on 127.0.0.1 and
// returns its port.
func serveConnector(t *testing.T, conn connectorv1connect.ConnectorServiceHandler) int32 {
	t.Helper()
	mux := http.NewServeMux()
	mux.Handle(connectorv1connect.NewConnectorServiceHandler(conn))
	srv := httptest.NewUnstartedServer(mux)
	var p http.Protocols
	p.SetHTTP1(true)
	p.SetUnencryptedHTTP2(true)
	srv.Config.Protocols = &p
	srv.Start()
	t.Cleanup(srv.Close)
	_, port, err := net.SplitHostPort(srv.Listener.Addr().String())
	if err != nil {
		t.Fatalf("SplitHostPort: %v", err)
	}
	n, err := strconv.ParseInt(port, 10, 32)
	if err != nil {
		t.Fatalf("ParseInt: %v", err)
	}
	return int32(n)
}

// fakeCluster is a cluster where each named plugin is a Ready connector
// whose Service points at 127.0.0.1:port, and the API server issues
// coreToken for the core's ServiceAccount.
func fakeCluster(t *testing.T, port int32, names ...string) *cluster {
	t.Helper()
	plugins := make([]runtime.Object, 0, len(names))
	client := fake.NewClientset()
	for _, name := range names {
		plugins = append(plugins, &unstructured.Unstructured{Object: map[string]any{
			"apiVersion": "openfief.io/v1alpha1", "kind": "Plugin",
			"metadata": map[string]any{"name": name, "namespace": "openfief"},
			"status": map[string]any{
				"conditions": []any{map[string]any{"type": "Ready", "status": "True"}},
				"effective":  map[string]any{"capabilities": []any{"connector"}},
			},
		}})
		svc := &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name: "plugin-" + name, Namespace: "openfief",
				Labels: map[string]string{ingestkube.LabelName: ingestkube.PluginAppName, ingestkube.LabelPlugin: name},
			},
			Spec: corev1.ServiceSpec{ClusterIP: "127.0.0.1", Ports: []corev1.ServicePort{{Name: ingestkube.PortName, Port: port}}},
		}
		if _, err := client.CoreV1().Services("openfief").Create(t.Context(), svc, metav1.CreateOptions{}); err != nil {
			t.Fatalf("creating a Service: %v", err)
		}
	}
	client.PrependReactor("create", "serviceaccounts", func(k8stesting.Action) (bool, runtime.Object, error) {
		return true, &authnv1.TokenRequest{Status: authnv1.TokenRequestStatus{
			Token: coreToken, ExpirationTimestamp: metav1.NewTime(time.Now().Add(10 * time.Minute)),
		}}, nil
	})
	dyn := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(),
		map[schema.GroupVersionResource]string{ingestkube.PluginResource(): "PluginList"}, plugins...)
	return &cluster{client: client, dynamic: dyn, namespace: "openfief"}
}

// connectorCmd runs a connector command as the CLI does, over cl, and
// returns its output.
func connectorCmd(t *testing.T, dsn string, cl *cluster, args ...string) (string, error) {
	t.Helper()
	cmd, err := parseArgs(append([]string{"connector"}, args...), func(k string) string {
		if k == dsnEnv {
			return dsn
		}
		return ""
	})
	if err != nil {
		t.Fatalf("parseArgs: %v", err)
	}
	var out bytes.Buffer
	err = withConnectorDeps(t.Context(), cmd.DSN, quiet(), func(d *connectorDeps) error {
		return connectorCommand(t.Context(), &cmd, d, cl, &out)
	})
	return out.String(), err
}

func TestConnectorBindAndSync(t *testing.T) {
	t.Parallel()
	dsn := emptyDatabase(t)
	port := serveConnector(t, depositConnector{account: string(demo.IBKRAccount)})
	cl := fakeCluster(t, port, "ibkr-main", "rogue")
	saveDemoRegistry(t, dsn)
	if _, bindErr := connectorCmd(t, dsn, cl, "bind", "--instance=ibkr-main", "--account=unknown-account"); bindErr == nil {
		t.Error("bound an instance to an account that isn't stored")
	}
	out, err := connectorCmd(t, dsn, cl, "bind", "--instance=ibkr-main", "--account="+string(demo.IBKRAccount))
	if err != nil || !strings.Contains(out, "bound connector instance ibkr-main") {
		t.Fatalf("connector bind = %q, %v", out, err)
	}

	// The first sync is an opening batch; a deposit seeds nothing, so a
	// transactions batch is asked for.
	out, err = connectorCmd(t, dsn, cl, "sync", "--instance=ibkr-main", "--mode=transactions")
	if err != nil || !strings.Contains(out, "(transactions): 1 transactions, 1 new") {
		t.Fatalf("connector sync = %q, %v", out, err)
	}
	out, err = connectorCmd(t, dsn, cl, "sync", "--instance=ibkr-main")
	if err != nil || !strings.Contains(out, "1 transactions, 0 new, 0 new versions, 1 unchanged") {
		t.Fatalf("second connector sync = %q, %v", out, err)
	}
	if _, syncErr := connectorCmd(t, dsn, cl, "sync", "--instance=rogue"); !errors.Is(syncErr, ingest.ErrAccountNotBound) {
		t.Errorf("sync of an unbound instance = %v, want ErrAccountNotBound", syncErr)
	}
	if _, syncErr := connectorCmd(t, dsn, nil, "sync", "--instance=ibkr-main"); !errors.Is(syncErr, errNoCluster) {
		t.Errorf("sync outside a cluster = %v, want errNoCluster", syncErr)
	}
	checkSynced(t, dsn)
}

// saveDemoRegistry saves the demo dataset's sleeves and accounts in dsn.
func saveDemoRegistry(t *testing.T, dsn string) {
	t.Helper()
	err := withConnectorDeps(t.Context(), dsn, quiet(), func(d *connectorDeps) error {
		today, err := kernel.ParseDate("2026-10-28")
		if err != nil {
			return fmt.Errorf("setup: %w", err)
		}
		ds, err := demo.Generate(demoSeed, today)
		if err != nil {
			return fmt.Errorf("setup: %w", err)
		}
		saveRegistry(t, d.ledger, &ds)
		return nil
	})
	if err != nil {
		t.Fatalf("saving the registry: %v", err)
	}
}

// checkSynced checks that the deposit is in the ledger once, keyed to the
// instance, and that the second sync's watermark is stored.
func checkSynced(t *testing.T, dsn string) {
	t.Helper()
	err := withConnectorDeps(t.Context(), dsn, quiet(), func(d *connectorDeps) error {
		txs, err := d.ledger.CurrentTransactions(t.Context(), demo.IBKRAccount)
		if err != nil {
			return fmt.Errorf("check: %w", err)
		}
		if len(txs) != 1 || txs[0].Key.Source != "plugin:ibkr-main:cash" {
			t.Errorf("ledger holds %+v, want one deposit keyed plugin:ibkr-main:cash", txs)
		}
		var watermark string
		if err := d.pool.QueryRow(t.Context(), "SELECT watermark FROM ingest.connector_instances WHERE instance = 'ibkr-main'").Scan(&watermark); err != nil {
			return fmt.Errorf("check: %w", err)
		}
		if watermark != "after-after-" {
			t.Errorf("stored watermark = %q, want after-after-", watermark)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("checking: %v", err)
	}
}

func TestServeSchedulesConnectors(t *testing.T) {
	t.Parallel()
	dsn := emptyDatabase(t)
	port := serveConnector(t, depositConnector{account: string(demo.IBKRAccount)})
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("connecting: %v", err)
	}
	defer pool.Close()
	cfg := defaultServe(t)
	cfg.Connectors = true
	// The instance is bound to nothing, so its first sync fails and is
	// recorded as a failed attempt of an open batch.
	c, err := start(ctx, pool, &cfg, coreNow, quiet(), fakeCluster(t, port, "ibkr-main"))
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	err = wait.PollUntilContextTimeout(ctx, 10*time.Millisecond, 30*time.Second, true, func(ctx context.Context) (bool, error) {
		var attempts int
		qErr := pool.QueryRow(ctx, "SELECT COALESCE(sum(attempts), 0) FROM ingest.import_batches WHERE instance = 'ibkr-main'").Scan(&attempts)
		if qErr != nil {
			return false, fmt.Errorf("counting attempts: %w", qErr)
		}
		return attempts > 0, nil
	})
	if err != nil {
		t.Fatalf("waiting for the first sync attempt: %v", err)
	}
	if c.deps.entries.client.Load() == nil {
		t.Error("the scheduler runs, but the owner's manual entries have no connector client")
	}
	cancel()
	if err := shutdown(context.WithoutCancel(ctx), &server{pool: pool}, &http.Server{ReadHeaderTimeout: time.Second}, c, 10*time.Second); err != nil {
		t.Errorf("shutdown: %v", err)
	}
	select {
	case <-c.connectors:
	default:
		t.Error("shutdown returned before the scheduler stopped")
	}
}

func TestMonthlyReading(t *testing.T) {
	t.Parallel()
	day, err := kernel.ParseDate("2026-03-05")
	if err != nil {
		t.Fatalf("ParseDate: %v", err)
	}
	r := ingest.IndexReading{Index: ingest.IndexReferenceRate, Date: day, Value: kernel.NewDecimal(575, -4), Source: "nbp"}
	got, err := monthlyReading(&r)
	if err != nil {
		t.Fatalf("monthlyReading: %v", err)
	}
	if got.Kind != valuation.ReferenceRate || got.Year != 2026 || got.Month != time.March || got.Source != "nbp" || !got.Value.Equal(r.Value) {
		t.Errorf("monthlyReading = %+v", got)
	}
	r.Index = 0
	if _, err := monthlyReading(&r); !errors.Is(err, errUnknownIndex) {
		t.Errorf("unknown index = %v, want errUnknownIndex", err)
	}
}
