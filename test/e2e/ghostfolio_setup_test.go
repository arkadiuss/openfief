//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	utilyaml "k8s.io/apimachinery/pkg/util/yaml"

	"github.com/arkadiuss/openfief/internal/kernel"
	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/operator"
	"github.com/arkadiuss/openfief/internal/views/ghostfolio"
	"github.com/arkadiuss/openfief/test/e2e/ghostfoliocheck"
)

// The suite's shards: task e2e runs every test but the Ghostfolio one,
// task e2e:ghostfolio only that one, against a real Ghostfolio, so the main
// runs don't wait for it. TestMain signs only the plugin images the shard's
// task e2e:images built (pluginImages).
const (
	shardEnv        = "OPENFIEF_E2E_SHARD"
	shardGhostfolio = "ghostfolio"
)

// The real Ghostfolio the shard runs (testdata/ghostfolio.yaml), in a
// namespace of its own outside openfief, as an owner's in-cluster service
// is (D-74).
const (
	gfManifests  = "testdata/ghostfolio.yaml"
	gfNamespace  = "openfief-e2e-ghostfolio"
	gfDeployment = "ghostfolio"
	gfService    = "ghostfolio"
	gfPort       = 3333
	gfDatabase   = "ghostfolio-db"
	// gfRunEnv is the Secret of the run's random JWT key, access token
	// salt and Redis password.
	gfRunEnv = "ghostfolio-env"
	// gfSecretBytes is how many random bytes each of them has.
	gfSecretBytes = 32
	// gfTimeout bounds the waits for Ghostfolio's database and API, the
	// first pull of their images included.
	gfTimeout = 10 * time.Minute
	// gfRequestTimeout bounds one request of the owner's to Ghostfolio.
	gfRequestTimeout = 2 * time.Minute
	// gfBodyLimit bounds a response of Ghostfolio's the owner reads.
	gfBodyLimit = 1 << 20
	// gfBaseCurrency is the Ghostfolio user's base currency, the
	// currency of the accounts the view creates.
	gfBaseCurrency = "PLN"
)

// The view plugin instance the shard installs, from plugins/ghostfolio's
// image in the suite's registry, signed with the run's key.
const (
	ghostfolioImage = "ghostfolio"
	gfInstance      = "e2e-ghostfolio"
	// gfTokenRole is the manifest's Secret role: the instance's Secret
	// <instance>--security-token holds the Ghostfolio user's security
	// token in the file security-token (D-89).
	gfTokenRole = "security-token"
	gfTokenFile = "security-token"
	// gfInterval is the view's sync interval, longer than the test: every
	// pass the test checks is the one a new pod runs at start.
	gfInterval = "1h"
)

var (
	// errNoShard reports TestGhostfolio run outside its shard.
	errNoShard = errors.New("TestGhostfolio runs in its own shard: run task e2e:ghostfolio")
	// errGhostfolioStatus reports a request of the owner's that Ghostfolio
	// refused.
	errGhostfolioStatus = errors.New("ghostfolio refused the request")
	// errNoAccessToken reports a sign-up without a security token.
	errNoAccessToken = errors.New("ghostfolio's sign-up gave no security token")
)

// deployGhostfolio applies Ghostfolio's namespace, its run's Secret, its
// database, Redis and Services, everything but Ghostfolio's own
// Deployment, which startGhostfolio applies once the database is ready, so
// the entrypoint's migration doesn't crash-loop. The namespace goes when t
// ends.
func deployGhostfolio(t *testing.T, k *kube) {
	t.Helper()
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), diagnosticsTimeout)
		defer cancel()
		err := k.clients.CoreV1().Namespaces().Delete(ctx, gfNamespace, metav1.DeleteOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			t.Errorf("deleting namespace %s: %v", gfNamespace, err)
		}
	})
	err := retry(t.Context(), webhookTimeout, func() error {
		return applyGhostfolio(t.Context(), k, func(obj *unstructured.Unstructured) bool { return !isGhostfolioAPI(obj) })
	})
	if err != nil {
		t.Fatalf("deploying Ghostfolio's database and cache: %v", err)
	}
	env := map[string]string{}
	for _, key := range []string{"ACCESS_TOKEN_SALT", "JWT_SECRET_KEY", "REDIS_PASSWORD"} {
		env[key] = randomHex(t)
	}
	s := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: gfRunEnv, Namespace: gfNamespace}, StringData: env}
	if _, err := k.clients.CoreV1().Secrets(gfNamespace).Create(t.Context(), s, metav1.CreateOptions{FieldManager: fieldManager}); err != nil {
		t.Fatalf("creating Secret %s: %v", gfRunEnv, err)
	}
}

// startGhostfolio waits for Ghostfolio's database, applies Ghostfolio's
// Deployment and waits until its API is available.
func startGhostfolio(t *testing.T, k *kube) {
	t.Helper()
	dbErr := waitFor(t.Context(), gfTimeout, func(ctx context.Context) (string, bool, error) {
		ready, err := k.conditionTrue(ctx, cnpgCluster(), gfNamespace, gfDatabase, "Ready")
		return "database not ready", ready, err
	})
	if dbErr != nil {
		t.Fatalf("waiting for Ghostfolio's database: %v", dbErr)
	}
	if applyErr := applyGhostfolio(t.Context(), k, isGhostfolioAPI); applyErr != nil {
		t.Fatalf("deploying Ghostfolio: %v", applyErr)
	}
	available := k.deploymentAvailable(gfNamespace, gfDeployment)
	upErr := waitFor(t.Context(), gfTimeout, func(ctx context.Context) (string, bool, error) {
		if state, done, err := available(ctx); done || (err != nil && !errors.Is(err, errRetry)) {
			return state, done, err
		}
		state, stateErr := k.podsState(ctx, gfNamespace, gfDeployment)
		return state, false, stateErr
	})
	if upErr != nil {
		t.Fatalf("waiting for Ghostfolio: %v", upErr)
	}
}

// isGhostfolioAPI reports whether obj is Ghostfolio's own Deployment.
func isGhostfolioAPI(obj *unstructured.Unstructured) bool {
	return obj.GetKind() == "Deployment" && obj.GetName() == gfDeployment
}

// applyGhostfolio applies the objects of testdata/ghostfolio.yaml that
// keep selects.
func applyGhostfolio(ctx context.Context, k *kube, keep func(*unstructured.Unstructured) bool) error {
	manifests, err := os.ReadFile(gfManifests)
	if err != nil {
		return fmt.Errorf("reading %s: %w", gfManifests, err)
	}
	dec := utilyaml.NewYAMLOrJSONDecoder(bytes.NewReader(manifests), yamlBuffer)
	for {
		var obj unstructured.Unstructured
		err := dec.Decode(&obj.Object)
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("decoding %s: %w", gfManifests, err)
		}
		if len(obj.Object) == 0 || !keep(&obj) {
			continue
		}
		if err := k.applyOne(ctx, &obj); err != nil {
			return err
		}
	}
}

// randomHex is a random secret of the run's, as hex.
func randomHex(t *testing.T) string {
	t.Helper()
	b := make([]byte, gfSecretBytes)
	if _, err := rand.Read(b); err != nil {
		t.Fatal(err)
	}
	return hex.EncodeToString(b)
}

// gfOwner is the owner's side of Ghostfolio: its address through a
// port-forward, the user the owner signed up, and a client of the view's,
// logged in as that user, through which the test reads what Ghostfolio
// holds.
type gfOwner struct {
	base string
	// securityToken is what the owner gives the view (its Secret); auth
	// is the bearer token of the owner's own requests.
	securityToken, auth string
	client              *ghostfolio.Client
	user                ghostfolio.User
	hc                  *http.Client
}

// signUpOwner forwards a local port to Ghostfolio's pod, signs a new
// anonymous user up, as an owner does on a fresh Ghostfolio, sets the
// user's base currency, and logs the view's client in with the user's
// security token. The first user of a fresh Ghostfolio is its
// administrator; no credential is made up.
func signUpOwner(t *testing.T, k *kube) *gfOwner {
	t.Helper()
	pod, err := k.runningPod(t.Context(), gfNamespace, gfDeployment)
	if err != nil {
		t.Fatal(err)
	}
	o := &gfOwner{
		base: "http://" + k.forwardPod(t, gfNamespace, podPort{pod: pod, port: gfPort}),
		hc:   &http.Client{Timeout: gfRequestTimeout},
	}
	var signed struct {
		AccessToken string `json:"accessToken"`
		AuthToken   string `json:"authToken"`
	}
	o.send(t, http.MethodPost, "/api/v1/user", nil, &signed)
	if signed.AccessToken == "" || signed.AuthToken == "" {
		t.Fatal(errNoAccessToken)
	}
	o.securityToken, o.auth = signed.AccessToken, signed.AuthToken
	o.send(t, http.MethodPut, "/api/v1/user/setting", map[string]any{"baseCurrency": gfBaseCurrency}, nil)
	o.client = ghostfolio.NewClient(o.hc, o.base, func() (string, error) { return o.securityToken, nil })
	if loginErr := o.client.Login(t.Context()); loginErr != nil {
		t.Fatalf("logging in to Ghostfolio: %v", loginErr)
	}
	if o.user, err = o.client.User(t.Context()); err != nil {
		t.Fatalf("reading the Ghostfolio user: %v", err)
	}
	if o.user.BaseCurrency.String() != gfBaseCurrency {
		t.Fatalf("the Ghostfolio user's base currency is %s, want %s", o.user.BaseCurrency, gfBaseCurrency)
	}
	return o
}

// send makes one request of the owner's to Ghostfolio, with body as JSON
// unless it is nil, and decodes the answer into out unless it is nil;
// any status but 2xx fails t. Tokens never appear in a failure.
func (o *gfOwner) send(t *testing.T, method, path string, body, out any) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), gfRequestTimeout)
	defer cancel()
	resp, err := o.hc.Do(o.request(t, method, path, body).WithContext(ctx))
	if err != nil {
		t.Fatalf("%s %s: %v", method, path, err)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, gfBodyLimit))
	if closeErr := resp.Body.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		t.Fatalf("%s %s: reading the answer: %v", method, path, err)
	}
	if resp.StatusCode/100 != 2 {
		t.Fatalf("%s %s: %v: HTTP %d", method, path, errGhostfolioStatus, resp.StatusCode)
	}
	if out != nil {
		if err := json.Unmarshal(data, out); err != nil {
			t.Fatalf("%s %s: decoding the answer: %v", method, path, err)
		}
	}
}

// request is the owner's request to Ghostfolio, with body as JSON unless
// it is nil, and the owner's bearer token once signed up; the caller gives
// it its context.
func (o *gfOwner) request(t *testing.T, method, path string, body any) *http.Request {
	t.Helper()
	var reader io.Reader = http.NoBody
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		reader = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(t.Context(), method, o.base+path, reader)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	if o.auth != "" {
		req.Header.Set("Authorization", "Bearer "+o.auth)
	}
	return req
}

// gfOwnAsset is the symbol of a MANUAL asset of the owner's own.
const gfOwnAsset = "6c1f0d2e-8b4a-4e57-9a3d-2f6b7c8d9e0a"

// gfOwn is what the owner keeps in the account Ghostfolio made at sign-up,
// which OpenFief didn't create: an activity and a cash balance of their
// own, each as a line with its ID.
type gfOwn struct {
	account              string
	activities, balances []string
}

// addOwn gives the account Ghostfolio made for the new user an activity on
// an asset of the owner's own and a cash balance, which no sync may touch,
// and returns them as Ghostfolio holds them.
func (o *gfOwner) addOwn(t *testing.T) gfOwn {
	t.Helper()
	ctx := t.Context()
	ids, err := o.client.Accounts(ctx)
	if err != nil || len(ids) != 1 {
		t.Fatalf("Ghostfolio's accounts after sign-up = %v, %v; want the account it made for the user", ids, err)
	}
	pln, err := kernel.ParseCurrency(gfBaseCurrency)
	if err != nil {
		t.Fatal(err)
	}
	day, err := kernel.ParseDate("2025-01-02")
	if err != nil {
		t.Fatal(err)
	}
	fee := ghostfolio.Activity{
		AccountID: ids[0], Comment: "a fee of my own", Type: ghostfolio.TypeFee, Date: day, Currency: pln,
		DataSource: ghostfolio.DataSourceManual, Symbol: gfOwnAsset, Fee: kernel.NewDecimal(3, 0),
	}
	asset := ghostfolio.AssetProfile{Symbol: gfOwnAsset, Name: "My own fees", Currency: pln}
	if _, err := o.client.Import(ctx, &ghostfolio.Batch{Activities: []ghostfolio.Activity{fee}, Assets: []ghostfolio.AssetProfile{asset}}, false); err != nil {
		t.Fatalf("importing the owner's own activity: %v", err)
	}
	if err := o.client.SetBalance(ctx, &ghostfolio.Balance{AccountID: ids[0], Date: day, Value: kernel.NewDecimal(42, 0)}); err != nil {
		t.Fatalf("setting the owner's own balance: %v", err)
	}
	own := o.ownNow(t, ids[0])
	if len(own.activities) != 1 || len(own.balances) != 1 {
		t.Fatalf("the owner's own account holds %v and %v; want their one activity and balance", own.activities, own.balances)
	}
	return own
}

// ownNow returns what account, the owner's own, holds now.
func (o *gfOwner) ownNow(t *testing.T, account string) gfOwn {
	t.Helper()
	acts, err := o.client.Activities(t.Context(), []string{account})
	if err != nil {
		t.Fatalf("reading the owner's own activities: %v", err)
	}
	bs, err := o.client.Balances(t.Context(), account)
	if err != nil {
		t.Fatalf("reading the owner's own balances: %v", err)
	}
	return gfOwn{account: account, activities: ghostfoliocheck.Describe(checkActivities(acts)), balances: ghostfoliocheck.DescribeBalances(checkBalances(bs))}
}

// installView creates the view instance's Secret with the owner's
// security token, labelled as the owner labels it (D-182), installs the
// instance with the openfief CLI built from this checkout, `openfief
// plugin install <image> --name e2e-ghostfolio --config <file>
// --cluster-service openfief-e2e-ghostfolio/ghostfolio:3333 --yes`, as an
// owner whose Ghostfolio runs in the same cluster does (D-74), and waits
// until it is ready. It is uninstalled with the CLI when t ends, which
// deletes its Secret.
func installView(t *testing.T, k *kube, reg suiteRegistry, securityToken string) {
	t.Helper()
	if _, err := os.Stat(cliPath); err != nil {
		t.Fatalf("%v: %v", errNoCLI, err)
	}
	img, err := reg.image(t.Context(), ghostfolioImage)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { uninstallWithCLI(t, gfInstance) })
	createInstanceSecret(t, k, gfInstance, gfTokenRole, map[string]string{gfTokenFile: securityToken})
	config, err := json.Marshal(map[string]any{"interval": gfInterval})
	if err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(t.TempDir(), gfInstance+".json")
	if writeErr := os.WriteFile(file, config, filePerm); writeErr != nil {
		t.Fatal(writeErr)
	}
	ctx, cancel := context.WithTimeout(t.Context(), cliTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, cliPath)
	// The arguments are the suite's own: names, a digest and a file it wrote.
	cmd.Args = append(cmd.Args, viewInstallArgs(img.cluster, file)...)
	out, err := cmd.CombinedOutput()
	if err != nil || !strings.Contains(string(out), "Installed "+gfInstance+".") {
		t.Fatalf("openfief plugin install %s: %v\n%s\nwant it installed", gfInstance, err, out)
	}
	waitPluginReady(t, k, gfInstance)
}

// viewInstallArgs are the arguments of `openfief plugin install` for the
// view instance from image, with its config from the file config and
// Ghostfolio's Service as its in-cluster service.
func viewInstallArgs(image plugins.ImageRef, config string) []string {
	return []string{
		"plugin", "install", image.String(), "--name", gfInstance, "--yes", "--config", config,
		"--cluster-service", fmt.Sprintf("%s/%s:%d", gfNamespace, gfService, gfPort),
	}
}

// waitPass waits until the view's pod named pod logged its first pass, and
// returns it; a failed pass fails t with its error.
func waitPass(t *testing.T, k *kube, pod string) ghostfoliocheck.Sync {
	t.Helper()
	var pass ghostfoliocheck.Sync
	err := waitFor(t.Context(), gfTimeout, func(ctx context.Context) (string, bool, error) {
		log, err := k.clients.CoreV1().Pods(namespace).GetLogs(pod, &corev1.PodLogOptions{Container: operator.ContainerName}).DoRaw(ctx)
		if err != nil {
			return "", false, fmt.Errorf("%w: the log of %s: %w", errRetry, pod, err)
		}
		passes, err := ghostfoliocheck.Syncs(log)
		if err != nil {
			return "", false, fmt.Errorf("the log of %s: %w", pod, err)
		}
		if len(passes) == 0 {
			return "no pass logged yet", false, nil
		}
		// The view tries a failed sync again within seconds (a pod can start
		// before its network rules allow the core); the latest pass counts,
		// and one that keeps failing fails the wait with its error.
		pass = passes[len(passes)-1]
		if pass.Err != "" {
			return "the latest sync failed: " + pass.Err, false, nil
		}
		return "a pass logged", true, nil
	})
	if err != nil {
		t.Fatalf("waiting for a successful sync of %s: %v", pod, err)
	}
	return pass
}

// restartView deletes the view's running pod, waits for the one that
// replaces it, and returns its name: the new pod syncs at start, which is
// how the test asks for another pass without waiting out the interval.
func restartView(t *testing.T, k *kube) string {
	t.Helper()
	old, err := k.runningPluginPod(t.Context(), gfInstance)
	if err != nil {
		t.Fatal(err)
	}
	if delErr := k.clients.CoreV1().Pods(namespace).Delete(t.Context(), old.Name, metav1.DeleteOptions{}); delErr != nil {
		t.Fatalf("deleting the view's pod %s: %v", old.Name, delErr)
	}
	var pod string
	err = waitFor(t.Context(), pluginTimeout, func(ctx context.Context) (string, bool, error) {
		p, podErr := k.runningPluginPod(ctx, gfInstance)
		if podErr != nil {
			return "", false, fmt.Errorf("%w: %w", errRetry, podErr)
		}
		if p.UID == old.UID {
			return "the old pod still runs", false, nil
		}
		pod = p.Name
		return "running", true, nil
	})
	if err != nil {
		t.Fatalf("waiting for the view's new pod: %v", err)
	}
	return pod
}
