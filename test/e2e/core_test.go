//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/encoding/protojson"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arkadiuss/openfief/internal/demo"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

const (
	// clientNamespace holds the suite's client pods (testdata/clients.yaml).
	clientNamespace = "openfief-e2e"
	// clientPod is admitted to the core's http port, as the demo's Traefik is.
	clientPod = "client"
	// outsiderPod is admitted by no policy.
	outsiderPod = "outsider"
	// coreURL is the core's Service, port http, as the demo's Ingress names it.
	coreURL = "http://openfief-core.openfief.svc.cluster.local:8080"
	// coreTimeout bounds the wait for the core Deployment to be available.
	coreTimeout = 5 * time.Minute
	// seedTimeout bounds the wait for the demo-seed Job, which loads and
	// values the whole demo dataset.
	seedTimeout = 12 * time.Minute
	// valuedTimeout bounds the wait for every demo account to be valued.
	valuedTimeout = 3 * time.Minute
	// callTimeout bounds one HTTP call from a client pod, kubectl exec included.
	callTimeout = time.Minute
	// connectTimeout is how many seconds a client pod's curl tries to connect.
	connectTimeout = "5"
	// traefikURL is k3s's bundled Traefik, its plain-HTTP entry point.
	traefikURL = "http://traefik.kube-system.svc.cluster.local"
	// routeHost is the host overlays/routes gives the demo's Ingress.
	routeHost = "openfief.e2e.test"
	// traefikTimeout bounds the waits for Traefik and its route.
	traefikTimeout = 3 * time.Minute
)

var (
	// errJobFailed reports a Job that gave up.
	errJobFailed = errors.New("job failed")
	// errNoReply reports curl output without a status line.
	errNoReply = errors.New("no HTTP status in curl's output")
	// errNoSecret reports token issue output without a secret.
	errNoSecret = errors.New("token issue printed no secret")
)

// TestCoreReadyThroughService checks that the core's Deployment becomes
// available under its real NetworkPolicy and that /readyz answers 200
// through the Service from a pod admitted the way the demo admits Traefik
// (SCEN-e2e-core-ready).
func TestCoreReadyThroughService(t *testing.T) {
	t.Parallel()
	k := mustConnect(t)
	diagnoseOnFailure(t, k)
	waitCoreAvailable(t, k)
	r := mustCall(t, k, clientPod, "/readyz", "")
	if r.status != http.StatusOK {
		t.Fatalf("GET /readyz from %s: status %d, body %q; want 200", clientPod, r.status, r.body)
	}
}

// TestTraefikRoutesToCore checks the demo machine's own way in to the read
// API: k3s's bundled Traefik routes /v1 of the demo's Ingress
// (overlays/routes) to the core through the core-from-traefik policy, which
// answers a call without a token with its own 401 rather than Traefik's 503
// "no available server" or the UI's 404 (SCEN-e2e-traefik-route). Only k3s
// bundles Traefik, so on kind there is nothing to check.
func TestTraefikRoutesToCore(t *testing.T) {
	t.Parallel()
	k := mustConnect(t)
	diagnoseOnFailure(t, k)
	skipUnlessK3s(t, k)
	waitCoreAvailable(t, k)
	waitTraefik(t, k)
	// Traefik picks up the Ingress and the core's endpoints on its own
	// schedule; until then it answers 404 or 503.
	err := waitFor(t.Context(), traefikTimeout, func(ctx context.Context) (string, bool, error) {
		r, getErr := get(ctx, k, clientPod, traefikURL+"/v1/accounts", "Host: "+routeHost)
		if getErr != nil {
			return "", false, getErr
		}
		return fmt.Sprintf("status %d, body %q", r.status, r.body), r.status == http.StatusUnauthorized, nil
	})
	if err != nil {
		t.Fatalf("GET /v1/accounts through Traefik: %v", err)
	}
}

// TestCorePolicyRefusesOtherPods checks that the core's NetworkPolicies are
// enforced: a pod no policy admits cannot connect to the core while an
// admitted one can (SCEN-e2e-core-policy-enforced, D-12).
func TestCorePolicyRefusesOtherPods(t *testing.T) {
	t.Parallel()
	k := mustConnect(t)
	diagnoseOnFailure(t, k)
	waitCoreAvailable(t, k)
	if r := answerFrom(t, k, clientPod, "/healthz"); r.status != http.StatusOK {
		t.Fatalf("GET /healthz from %s: status %d, body %q; want 200", clientPod, r.status, r.body)
	}
	if r := answerFrom(t, k, outsiderPod, "/healthz"); r.status != 0 {
		t.Fatalf("GET /healthz from %s: status %d; want no connection (is NetworkPolicy enforced?)", outsiderPod, r.status)
	}
}

// answerFrom calls path on the core from pod with answerAt.
func answerFrom(t *testing.T, k *kube, pod, path string) reply {
	t.Helper()
	return answerAt(t, k, pod, coreURL+path)
}

// answerAt sends GET url from pod until curl reports a result, retrying
// only when the exec stream itself fails (as it can on kind while the API
// server is busy), and returns that result: an HTTP status, or 0 when curl
// got no connection. It never retries on an HTTP answer, so a refusal or a
// wrong status still fails the test.
func answerAt(t *testing.T, k *kube, pod, url string) reply {
	t.Helper()
	var got reply
	err := waitFor(t.Context(), coreTimeout, func(ctx context.Context) (string, bool, error) {
		r, callErr := get(ctx, k, pod, url)
		if callErr != nil {
			return "", false, fmt.Errorf("%w: %w", errRetry, callErr)
		}
		got = r
		return "", true, nil
	})
	if err != nil {
		t.Fatalf("GET %s from %s: %v", url, pod, err)
	}
	return got
}

// TestDemoSeedCompletes checks that the demo-seed Job loads and values the
// demo dataset next to the running core (SCEN-e2e-demo-seed-completes).
func TestDemoSeedCompletes(t *testing.T) {
	t.Parallel()
	k := mustConnect(t)
	diagnoseOnFailure(t, k)
	waitSeedComplete(t, k)
}

// TestReadAPIWithScriptToken checks that, with a script token issued by the
// core's own CLI, the read API lists the three demo accounts and reports each
// valued through today (SCEN-e2e-read-api-demo-accounts).
func TestReadAPIWithScriptToken(t *testing.T) {
	t.Parallel()
	k := mustConnect(t)
	diagnoseOnFailure(t, k)
	waitCoreAvailable(t, k)
	waitSeedComplete(t, k)
	today := warsawToday(t)
	token := issueToken(t, k)

	r := mustCall(t, k, clientPod, "/v1/accounts", token)
	if r.status != http.StatusOK {
		t.Fatalf("GET /v1/accounts: status %d, body %q; want 200", r.status, r.body)
	}
	var accounts viewsv1.ListAccountsResponse
	if err := protojson.Unmarshal(r.body, &accounts); err != nil {
		t.Fatalf("decoding the accounts: %v", err)
	}
	got := make([]string, 0, len(accounts.GetAccounts()))
	for _, a := range accounts.GetAccounts() {
		got = append(got, a.GetId())
	}
	want := demoAccountIDs()
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("GET /v1/accounts: account IDs (-want +got):\n%s", diff)
	}

	err := waitFor(t.Context(), valuedTimeout, func(ctx context.Context) (string, bool, error) {
		r, err := call(ctx, k, clientPod, "/v1/valuation-status", token)
		if err != nil {
			return "", false, err
		}
		if r.status != http.StatusOK {
			return fmt.Sprintf("status %d, body %q", r.status, r.body), false, nil
		}
		var statuses viewsv1.ListValuationStatusesResponse
		if err := protojson.Unmarshal(r.body, &statuses); err != nil {
			return "", false, fmt.Errorf("decoding the valuation statuses: %w", err)
		}
		return valuationState(&statuses), allValued(&statuses, want, today()), nil
	})
	if err != nil {
		t.Fatalf("waiting for every demo account to be valued through today: %v", err)
	}
}

// TestReadAPIRefusesWithoutToken checks that the read API answers 401 to a
// call without a script token (SCEN-e2e-read-api-needs-token).
func TestReadAPIRefusesWithoutToken(t *testing.T) {
	t.Parallel()
	k := mustConnect(t)
	diagnoseOnFailure(t, k)
	waitCoreAvailable(t, k)
	r := mustCall(t, k, clientPod, "/v1/accounts", "")
	if r.status != http.StatusUnauthorized {
		t.Fatalf("GET /v1/accounts without a token: status %d, body %q; want 401", r.status, r.body)
	}
}

// waitCoreAvailable waits until the core Deployment is available, failing
// at once if its pod can't start.
func waitCoreAvailable(t *testing.T, k *kube) {
	t.Helper()
	available := k.deploymentAvailable(namespace, "openfief-core")
	err := waitFor(t.Context(), coreTimeout, func(ctx context.Context) (string, bool, error) {
		if state, done, err := available(ctx); done || (err != nil && !errors.Is(err, errRetry)) {
			return state, done, err
		}
		state, err := k.podsState(ctx, namespace, "openfief-core")
		return state, false, err
	})
	if err != nil {
		t.Fatalf("waiting for the core Deployment to be available: %v", err)
	}
}

// waitSeedComplete waits until the demo-seed Job completes, failing at once
// if it fails or its pod can't start.
func waitSeedComplete(t *testing.T, k *kube) {
	t.Helper()
	err := waitFor(t.Context(), seedTimeout, func(ctx context.Context) (string, bool, error) {
		j, err := k.clients.BatchV1().Jobs(namespace).Get(ctx, "demo-seed", metav1.GetOptions{})
		if err != nil {
			return "", false, fmt.Errorf("%w: %w", errRetry, err)
		}
		if done, jobErr := jobFinished(j); done || jobErr != nil {
			return "complete", done, jobErr
		}
		state, err := k.podsState(ctx, namespace, "demo-seed")
		return fmt.Sprintf("%d failed attempts; %s", j.Status.Failed, state), false, err
	})
	if err != nil {
		t.Fatalf("waiting for the demo-seed Job to complete: %v", err)
	}
}

// jobFinished reports whether j completed, or an error if it failed.
func jobFinished(j *batchv1.Job) (bool, error) {
	for _, c := range j.Status.Conditions {
		if c.Status != corev1.ConditionTrue {
			continue
		}
		switch c.Type {
		case batchv1.JobComplete:
			return true, nil
		case batchv1.JobFailed:
			return false, fmt.Errorf("%w: %s: %s: %s", errJobFailed, j.Name, c.Reason, c.Message)
		case batchv1.JobSuspended, batchv1.JobFailureTarget, batchv1.JobSuccessCriteriaMet:
		}
	}
	return false, nil
}

// reply is an HTTP answer seen from a client pod; status 0 means no
// connection.
type reply struct {
	status int
	body   []byte
}

// call sends GET path to the core's Service from pod with curl, with token
// as the bearer token unless it is empty. The token goes in on stdin, so it
// is never in a command line or an error message. A reply with status 0 is
// no connection, which is not an error.
func call(ctx context.Context, k *kube, pod, path, token string) (reply, error) {
	var headers []string
	if token != "" {
		headers = append(headers, "Authorization: Bearer "+token)
	}
	return get(ctx, k, pod, coreURL+path, headers...)
}

// get sends GET url from pod with curl and the given header lines, which go
// in on stdin.
func get(ctx context.Context, k *kube, pod, url string, headers ...string) (reply, error) {
	ctx, cancel := context.WithTimeout(ctx, callTimeout)
	defer cancel()
	var stdin bytes.Buffer
	for _, h := range headers {
		stdin.WriteString(h + "\n")
	}
	out, err := k.exec(ctx, clientNamespace, pod, "curl", append([]byte{}, stdin.Bytes()...),
		"curl", "-sS", "--connect-timeout", connectTimeout, "-H", "@-", "-w", "\n%{http_code}", url)
	r, perr := parseReply(out)
	switch {
	case perr != nil:
		return reply{}, fmt.Errorf("GET %s from %s: %w (%w)", url, pod, perr, err)
	case err != nil && r.status != 0:
		return reply{}, fmt.Errorf("GET %s from %s: status %d, then %w", url, pod, r.status, err)
	}
	return r, nil
}

// mustCall is call that fails t on an error.
func mustCall(t *testing.T, k *kube, pod, path, token string) reply {
	t.Helper()
	r, err := call(t.Context(), k, pod, path, token)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

// parseReply splits curl's output into the body and the status code that
// -w wrote on the last line; curl writes 000 when it got no answer.
func parseReply(out []byte) (reply, error) {
	i := bytes.LastIndexByte(out, '\n')
	if i < 0 {
		return reply{}, errNoReply
	}
	status, err := strconv.Atoi(string(bytes.TrimSpace(out[i+1:])))
	if err != nil {
		return reply{}, fmt.Errorf("%w: %w", errNoReply, err)
	}
	return reply{status: status, body: out[:i]}, nil
}

// issueToken issues a script token with the core's own CLI inside its pod,
// as an owner would with kubectl exec, and returns its secret, the first
// line of the output. The output is never logged.
func issueToken(t *testing.T, k *kube) string {
	t.Helper()
	pod, err := k.runningPod(t.Context(), namespace, "openfief-core")
	if err != nil {
		t.Fatal(err)
	}
	out, err := k.exec(t.Context(), namespace, pod, "core", nil, "/core", "token", "issue", "--name", "e2e", "--ttl", "1h")
	if err != nil {
		t.Fatalf("issuing a script token: %v", err)
	}
	line, _, _ := bytes.Cut(out, []byte("\n"))
	if len(line) == 0 {
		t.Fatal(errNoSecret)
	}
	return string(line)
}

// demoAccountIDs returns the demo dataset's account IDs in ID order, the
// order the read API lists them in.
func demoAccountIDs() []string {
	accounts := demo.Accounts()
	ids := make([]string, 0, len(accounts))
	for _, a := range accounts {
		ids = append(ids, string(a.ID))
	}
	slices.Sort(ids)
	return ids
}

// warsawToday returns a function that reports the dates that count as today
// in Europe/Warsaw, the demo's time zone: the date when the test started and
// the date now, so a run across midnight still passes.
func warsawToday(t *testing.T) func() []string {
	t.Helper()
	loc, err := time.LoadLocation("Europe/Warsaw")
	if err != nil {
		t.Fatalf("time zone: %v", err)
	}
	start := time.Now().In(loc).Format(time.DateOnly)
	return func() []string {
		return []string{start, time.Now().In(loc).Format(time.DateOnly)}
	}
}

// allValued reports whether each of ids is valued through one of today with
// no revaluation pending or running.
func allValued(statuses *viewsv1.ListValuationStatusesResponse, ids, today []string) bool {
	valued := map[string]bool{}
	for _, s := range statuses.GetStatuses() {
		valued[s.GetAccountId()] = !s.GetPending() && !s.GetRunning() && slices.Contains(today, s.GetValuedThrough())
	}
	for _, id := range ids {
		if !valued[id] {
			return false
		}
	}
	return true
}

// valuationState describes statuses in a line for a timeout message.
func valuationState(statuses *viewsv1.ListValuationStatusesResponse) string {
	var b bytes.Buffer
	for _, s := range statuses.GetStatuses() {
		fmt.Fprintf(&b, "%s through %q pending=%t running=%t; ", s.GetAccountId(), s.GetValuedThrough(), s.GetPending(), s.GetRunning())
	}
	return b.String()
}
