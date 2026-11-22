//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"slices"
	"strings"
	"testing"
	"time"

	"google.golang.org/protobuf/encoding/protojson"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/yaml"

	"github.com/arkadiuss/openfief/internal/plugins"
	viewsv1 "github.com/arkadiuss/openfief/sdk/go/gen/openfief/views/v1"
)

// The upgrade run (task e2e runs it between the bundle run and the demo run): TestMain installs the PUBLISHED
// release in publishedDir, the real artifacts scripts/release-fetch.sh downloaded and verified with cosign as the
// install guide tells an owner to, in the guide's order and with its commands (kubectl apply --server-side).
// TestUpgradeFromPublished then seeds that install with the synthetic demo dataset, loaded by that release's own
// demo-seed image, upgrades it in place to the bundle of this checkout (OPENFIEF_E2E_BUNDLE, the one the bundle run
// installs) the way the guide's "Upgrade" section says, and checks what D-161, D-209, D-212 and D-216 promise an
// owner who upgrades. TestMain removes the install again afterwards.
const (
	// upgradeEnv, set, makes TestMain run the upgrade run instead of the demo's. The published release is in
	// publishedDir, relative to this package, in the gitignored .bin, where task e2e:upgrade has
	// scripts/release-fetch.sh put what it downloaded and verified: the release's bundle, its CLI under bin/ and
	// the demo-seed image.
	upgradeEnv   = "OPENFIEF_E2E_UPGRADE"
	publishedDir = "../../.bin/e2e/published"
	// publishedCLI is the release's own openfief CLI, which installs the UI on the published release.
	publishedCLI = publishedDir + "/bin/openfief"
	// seedImageFile is the file in publishedDir that names the release's demo-seed image.
	seedImageFile = "demo-seed.image"
	// upgradeSeedFile is the loader's Job and NetworkPolicy; seedJob the Job's name.
	upgradeSeedFile = "testdata/upgrade-seed.yaml"
	seedJob         = "demo-seed"
	// kubectlTimeout bounds one kubectl call and one database call of the run.
	kubectlTimeout = 5 * time.Minute
	// settleTimeout bounds the wait for the published release's core to raise its attention items and for the
	// seeded data to stop changing, and for the new core to derive its valuations again; the core runs its checks
	// once a minute.
	settleTimeout = 6 * time.Minute
	// stableSnapshots is how many snapshots in a row must be equal before the data counts as at rest.
	stableSnapshots = 3
	// upgradeTimeout bounds the new release's rollout.
	upgradeTimeout = 10 * time.Minute
	// ownerDB is the database the install's openfief-db cluster creates.
	ownerDB = "openfief"
	// householdID is the one household target (identity.GrantTarget).
	householdID = "home"
	// syntheticTickers is a confirmed and an unconfirmed ticker with one entry, in the shape the published release
	// stores them: the demo loader writes no ingest row, and the migration that moves a ticker's confirmation to
	// its institution (ingest 0015, D-153) is the one that drops columns, so the seed gives it rows to move. The
	// ISINs and the instance are made up (GR-12).
	syntheticTickers = `
INSERT INTO ingest.tickers (instance, symbol, code, country, name, proposed_isin, proposed_name, isin, confirmed_at, proposed_currency, currency)
VALUES ('e2e-synthetic', 'SYNTH.PL', 'SYNTH', 'PL', 'Synthetic Plc', 'PLSYNTH00014', 'Synthetic Plc', 'PLSYNTH00014', '2026-01-02T10:00:00Z', 'PLN', 'PLN');
INSERT INTO ingest.tickers (instance, symbol, code, country, name, proposed_isin, proposed_name, proposed_currency)
VALUES ('e2e-synthetic', 'WAIT.PL', 'WAIT', 'PL', 'Waiting Plc', 'PLWAIT000019', 'Waiting Plc', 'PLN');
INSERT INTO ingest.ticker_entries (instance, key_source, external_id, symbol, account_id, kind, date, quantity, amount, currency, charge, charge_currency, description, ref)
VALUES ('e2e-synthetic', 'xtb', 'synthetic-1', 'SYNTH.PL', 'demo-xtb', 'buy', '2026-01-02', '10', '-1000.00', 'PLN', '0', '', 'synthetic buy', 'synthetic-ref-1');
`
	// confirmedTicker is what ingest 0015 makes of the confirmed ticker: one confirmation of the institution, which
	// is the instance's own, since it reported no broker account, in version 1 with the market's country, the
	// proposal's currency, as the confirmed ISIN is the proposed one, and the owner's pick.
	confirmedTicker = "instance:e2e-synthetic|SYNTH.PL|1|PLSYNTH00014|PL|PLN|PLN"
)

var (
	// errNoBundle reports an upgrade run without the bundle to upgrade to.
	errNoBundle = errors.New(bundleEnv + " is not set; run the suite with task e2e")
	// errNoKubectl reports a machine without kubectl, which the install guide's commands need.
	errNoKubectl = errors.New("kubectl is not on PATH; the upgrade run applies the release as the install guide does")
	// errNoSeedImage reports a published release without the demo-seed image the fetch script verified.
	errNoSeedImage = errors.New("the published release has no verified demo-seed image")
	// errBadSeedImage reports a seed image reference that isn't by digest.
	errBadSeedImage = errors.New("the demo-seed image isn't pinned by digest")
	// errNoPrimary reports a database cluster without a ready primary.
	errNoPrimary = errors.New("no ready primary in the openfief-db cluster")
	// errQuery reports a database answer that isn't in the shape the test asked for.
	errQuery = errors.New("unexpected database answer")
	// errCoreArgs reports an install without the core container whose args the test extends.
	errCoreArgs = errors.New("the install has no core container to start without default sleeves")
	// errSeedJob reports a loader Job that isn't the one the test expects.
	errSeedJob = errors.New("the loader's Job has not exactly one container")
)

// publishedRelease is the published release the upgrade run installs first.
type publishedRelease struct {
	bundle bundle
	// seedImage is the release's demo-seed image, by the digest cosign verified.
	seedImage string
}

// readRelease reads what scripts/release-fetch.sh left in publishedDir.
func readRelease() (*publishedRelease, error) {
	b, err := readBundle(publishedDir)
	if err != nil {
		return nil, err
	}
	root, err := os.OpenRoot(publishedDir)
	if err != nil {
		return nil, fmt.Errorf("opening the release's directory: %w", err)
	}
	image, err := root.ReadFile(seedImageFile)
	if closeErr := root.Close(); err == nil && closeErr != nil {
		err = closeErr
	}
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errNoSeedImage, err)
	}
	ref := strings.TrimSpace(string(image))
	if !strings.Contains(ref, "@sha256:") {
		return nil, fmt.Errorf("%w: %q", errBadSeedImage, ref)
	}
	if _, err := os.Stat(publishedCLI); err != nil {
		return nil, fmt.Errorf("the release's CLI: %w", err)
	}
	return &publishedRelease{bundle: b, seedImage: ref}, nil
}

// runUpgrade is TestMain's upgrade run: it installs the published release, runs the tests, removes the install
// whatever happened, and returns the exit code.
func runUpgrade(m *testing.M) int {
	k, err := connect()
	if err == nil {
		ctx, cancel := context.WithTimeout(context.Background(), installTimeout)
		err = installPublished(ctx, k, logStderrf)
		cancel()
	}
	code := 1
	if err != nil {
		logStderrf("e2e: installing the published release failed: %v", err)
		if k != nil {
			dctx, dcancel := context.WithTimeout(context.Background(), diagnosticsTimeout)
			k.dumpDiagnostics(dctx, logStderrf)
			dcancel()
		}
	} else {
		code = m.Run()
	}
	if k != nil {
		ctx, cancel := context.WithTimeout(context.Background(), removeTimeout)
		// The demo run that follows installs its own CRDs on a clean cluster, as the bundle run's end leaves it.
		if rmErr := errors.Join(k.removeNamespace(ctx, logStderrf), k.removeClusterScoped(ctx, logStderrf)); rmErr != nil {
			logStderrf("e2e: removing the published release's install: %v", rmErr)
			code = 1
		}
		cancel()
	}
	return code
}

// removeClusterScoped deletes what an install of OpenFief leaves outside its namespace: the ClusterRoleBindings
// and ClusterRoles of the core's and the egress proxy's token reviews, and the Plugin and PluginGrant CRDs, then
// waits until the CRDs are gone. Its Custom Resources went with the namespace, so none is left to hold them. What
// isn't there is fine.
func (k *kube) removeClusterScoped(ctx context.Context, logf func(string, ...any)) error {
	rbac := func(resource string) schema.GroupVersionResource {
		return schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: resource}
	}
	crd := schema.GroupVersionResource{Group: "apiextensions.k8s.io", Version: "v1", Resource: "customresourcedefinitions"}
	objects := []struct {
		resource schema.GroupVersionResource
		name     string
	}{
		{rbac("clusterrolebindings"), "openfief-core-tokenreview"},
		{rbac("clusterrolebindings"), "openfief-egress-proxy-tokenreview"},
		{rbac("clusterroles"), "openfief-core-tokenreview"},
		{rbac("clusterroles"), "openfief-egress-proxy-tokenreview"},
		{crd, "plugins.openfief.io"},
		{crd, "plugingrants.openfief.io"},
	}
	logf("e2e: removing OpenFief's cluster-scoped objects: its CRDs and token-review RBAC")
	for _, o := range objects {
		if err := k.dynamic.Resource(o.resource).Delete(ctx, o.name, metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
			return fmt.Errorf("deleting %s %s: %w", o.resource.Resource, o.name, err)
		}
	}
	for _, o := range objects {
		err := waitFor(ctx, removeTimeout, func(ctx context.Context) (string, bool, error) {
			_, getErr := k.dynamic.Resource(o.resource).Get(ctx, o.name, metav1.GetOptions{})
			switch {
			case apierrors.IsNotFound(getErr):
				return "gone", true, nil
			case getErr != nil:
				return "", false, fmt.Errorf("%w: %w", errRetry, getErr)
			}
			return "still there", false, nil
		})
		if err != nil {
			return fmt.Errorf("waiting for %s %s to be removed: %w", o.resource.Resource, o.name, err)
		}
	}
	return nil
}

// installPublished installs the published release on a cluster without OpenFief, in the install guide's order and
// with its commands: CloudNativePG, the CRDs, the database, the install, and a wait for every rollout. As in the
// bundle run, the suite's key is added to the owner's trusted signers between the CRDs and the install (D-45),
// which the run's CLI, built from this checkout, needs to upgrade the UI to this checkout's image.
func installPublished(ctx context.Context, k *kube, logf func(string, ...any)) error {
	rel, err := readRelease()
	if err != nil {
		return err
	}
	if _, lookErr := exec.LookPath("kubectl"); lookErr != nil {
		return errNoKubectl
	}
	reg, err := registryFromEnv()
	if err != nil {
		return err
	}
	b := &rel.bundle
	steps := []func() error{
		func() error { return k.removeNamespace(ctx, logf) },
		// The bundle run before this one leaves the checkout's CRDs and cluster-scoped RBAC, which belong to the
		// field manager of the suite: the published release's older CRDs must not be applied over them, nor forced.
		func() error { return k.removeClusterScoped(ctx, logf) },
		func() error {
			logf("e2e: installing CloudNativePG from the published release")
			return applyCNPG(ctx, k, b.cnpg)
		},
		func() error { return kubectlApply(ctx, logf, "the published release's CRDs", b.crds) },
		func() error {
			signer, trustErr := trustSigner(ctx, k, reg, logf)
			if trustErr != nil {
				return trustErr
			}
			return buildCLI(ctx, reg, signer, logf)
		},
		func() error { return kubectlApply(ctx, logf, "the published release's database", b.database) },
		func() error { return waitDatabase(ctx, k, logf) },
		func() error {
			install, editErr := withoutDefaultSleeves(b.install)
			if editErr != nil {
				return editErr
			}
			return kubectlApply(ctx, logf, "the published release's install, its core without default sleeves", install)
		},
		func() error { return waitDeployments(ctx, k, b) },
	}
	for _, step := range steps {
		if stepErr := step(); stepErr != nil {
			return stepErr
		}
	}
	return nil
}

// withoutDefaultSleeves returns install with the core started with --ensure-defaults=false, as the published
// release's own demo layer does (deploy/demo/openfief): the core stores the three default sleeves at its first
// start, and the demo loader refuses a database that holds data and no demo marker ("3 rows of ledger and valuation
// data"), which is the guard a real portfolio relies on, so it stays on. The same kubectl applies the candidate's
// install afterwards, whose args replace this list, and the candidate's core finds the demo's sleeves and adds none.
func withoutDefaultSleeves(install []byte) ([]byte, error) {
	objs, err := decodeEdited(install, noDefaultSleeves)
	if err != nil {
		return nil, err
	}
	var out bytes.Buffer
	for _, obj := range objs {
		doc, marshalErr := yaml.Marshal(obj.Object)
		if marshalErr != nil {
			return nil, fmt.Errorf("encoding %s %s: %w", obj.GetKind(), obj.GetName(), marshalErr)
		}
		out.WriteString("---\n")
		out.Write(doc)
	}
	return out.Bytes(), nil
}

// noDefaultSleeves adds --ensure-defaults=false to the args of the core's container, and changes no other object.
func noDefaultSleeves(obj *unstructured.Unstructured) error {
	if obj.GetKind() != "Deployment" || obj.GetName() != coreDeployment {
		return nil
	}
	containers, found, err := unstructured.NestedSlice(obj.Object, "spec", "template", "spec", "containers")
	if err != nil || !found {
		return fmt.Errorf("%w: the core's containers: %w", errCoreArgs, err)
	}
	for i, raw := range containers {
		c, ok := raw.(map[string]any)
		if !ok || c["name"] != "core" {
			continue
		}
		args, ok := c["args"].([]any)
		if !ok {
			return errCoreArgs
		}
		c["args"] = append(args, "--ensure-defaults=false")
		containers[i] = c
		if err := unstructured.SetNestedSlice(obj.Object, containers, "spec", "template", "spec", "containers"); err != nil {
			return fmt.Errorf("%w: %w", errCoreArgs, err)
		}
		return nil
	}
	return errCoreArgs
}

// waitDeployments waits until each Deployment of the bundle's install has rolled out its latest template on all
// its replicas, with no old pod left.
func waitDeployments(ctx context.Context, k *kube, b *bundle) error {
	deployments, err := deploymentsIn(b.install)
	if err != nil {
		return err
	}
	for i := range deployments {
		if err := waitRolledOut(ctx, k, deployments[i].Name); err != nil {
			return err
		}
	}
	return nil
}

// kubectlApply runs `kubectl apply --server-side -f -` over manifest, the install guide's command for one file of
// the bundle, with the kubeconfig the suite uses.
func kubectlApply(ctx context.Context, logf func(string, ...any), what string, manifest []byte) error {
	logf("e2e: kubectl apply --server-side: %s", what)
	ctx, cancel := context.WithTimeout(ctx, kubectlTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "kubectl", "apply", "--server-side", "-f", "-")
	cmd.Stdin = bytes.NewReader(manifest)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("kubectl apply --server-side (%s): %w\n%s", what, err, out)
	}
	return nil
}

// cliRun runs cmd, a CLI of the suite's own, and fails t when it does.
func cliRun(t *testing.T, cmd *exec.Cmd) {
	t.Helper()
	var out bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &out
	if err := cmd.Run(); err != nil {
		t.Fatalf("%s: %v\n%s", strings.Join(cmd.Args, " "), err, out.String())
	}
}

// TestUpgradeFromPublished upgrades a published release in place to this checkout's bundle and checks that nothing
// the owner stored is lost or changed, that the household-grant cutover ran exactly once, that the SpiceDB
// credentials Job made its Secrets, that the upgraded core answers from SpiceDB, the default, with the cutover's
// household grants written into it for the script token and the view plugin that existed, that the agentic switch
// stays off, and that the core is ready and the UI serves (SCEN-e2e-upgrade-keeps-stored-data).
func TestUpgradeFromPublished(t *testing.T) {
	t.Parallel()
	k := mustConnect(t)
	diagnoseOnFailure(t, k)
	rel, err := readRelease()
	if err != nil {
		t.Fatal(err)
	}
	if os.Getenv(bundleEnv) == "" {
		t.Fatal(errNoBundle)
	}
	candidate, err := readBundle(os.Getenv(bundleEnv))
	if err != nil {
		t.Fatal(err)
	}

	seeded := seedPublished(t, k, rel)
	before := settledSnapshot(t, k)
	checkSeeded(t, k, before)

	upgradeAsGuided(t, k, &rel.bundle, &candidate)

	checkStoredData(t, k, before)
	checkTickerConfirmations(t, k)
	checkGrantCutover(t, k, seeded)
	checkSpiceDBCredentials(t, k)
	checkCutoverInSpiceDB(t, k, seeded)
	checkAgenticSwitchOff(t, k, seeded.kept)
	checkServing(t, k, seeded)
	checkCutoverRanOnce(t, k)
}

// seededInstall is what the test did to the published release's install: kept is a script token that stays valid,
// revoked one that was revoked before the upgrade.
type seededInstall struct {
	kept, revoked issuedToken
}

// seedPublished seeds the published release's install: the built-in UI, installed with the release's own CLI as
// the guide does; the synthetic demo dataset, which the release's own demo-seed image loads; and two script
// tokens, one then revoked, from the release's own core.
func seedPublished(t *testing.T, k *kube, rel *publishedRelease) seededInstall {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), cliTimeout)
	install := exec.CommandContext(ctx, publishedCLI, "plugin", "install", uiPlugin, "--yes")
	cliRun(t, install)
	cancel()
	waitPluginReady(t, k, uiPlugin)

	loadDataset(t, k, rel.seedImage)
	if _, err := k.psql(t.Context(), syntheticTickers); err != nil {
		t.Fatalf("seeding the synthetic tickers: %v", err)
	}

	kept := issueTokenWith(t, k)
	revoked := issueTokenWith(t, k)
	pod, err := k.runningPod(t.Context(), namespace, "openfief-core")
	if err != nil {
		t.Fatal(err)
	}
	if _, execErr := k.exec(t.Context(), namespace, pod, "core", nil, "/core", "token", "revoke", revoked.id); execErr != nil {
		t.Fatalf("revoking a script token: %v", execErr)
	}
	return seededInstall{kept: kept, revoked: revoked}
}

// loadDataset runs the published release's own demo-seed image, named by ref, in the loader's Job and waits until it
// completes: it loads the synthetic demo dataset into the release's database through the release's own code.
func loadDataset(t *testing.T, k *kube, ref string) {
	t.Helper()
	t.Cleanup(func() {
		if t.Failed() {
			logGuardTables(t, k)
		}
	})
	seed, err := os.ReadFile(upgradeSeedFile)
	if err != nil {
		t.Fatalf("reading the loader's Job: %v", err)
	}
	// An unstructured Job holds its containers as a list, which SetNestedField can't index, so the edit sets the list.
	setSeedImage := func(obj *unstructured.Unstructured) error {
		if obj.GetKind() != "Job" {
			return nil
		}
		containers, found, nestedErr := unstructured.NestedSlice(obj.Object, "spec", "template", "spec", "containers")
		if nestedErr != nil || !found || len(containers) != 1 {
			return fmt.Errorf("%w: %d found: %w", errSeedJob, len(containers), nestedErr)
		}
		c, ok := containers[0].(map[string]any)
		if !ok {
			return errSeedJob
		}
		c["image"] = ref
		return unstructured.SetNestedSlice(obj.Object, []any{c}, "spec", "template", "spec", "containers")
	}
	if applyErr := k.applyEdited(t.Context(), seed, setSeedImage, nil); applyErr != nil {
		t.Fatalf("applying the loader's Job: %v", applyErr)
	}
	waitJobComplete(t, k, seedJob)
}

// guardTables are the tables whose rows make the demo loader refuse a database that has no demo marker
// (cmd/demo-seed's dataTables, as the published release has them).
func guardTables() []string {
	return []string{
		"portfolio.sleeves", "ledger.accounts", "ledger.transaction_versions", "valuation.prices", "valuation.fx_rates",
		"valuation.index_readings", "valuation.bond_terms", "valuation.tax_wrappers",
	}
}

// logGuardTables logs how many rows each of the loader's guard tables holds, and whether the demo marker is there, so
// a refusal explains itself. It logs counts only.
func logGuardTables(t *testing.T, k *kube) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), diagnosticsTimeout)
	defer cancel()
	var query strings.Builder
	for _, table := range guardTables() {
		fmt.Fprintf(&query, "SELECT '%s' || '|' || count(*) FROM %s UNION ALL ", table, table)
	}
	query.WriteString("SELECT 'demo marker|' || coalesce((SELECT count(*)::text FROM demo.marker), 'no table')")
	lines, err := queryLinesCtx(ctx, k, query.String())
	if err != nil {
		t.Logf("the loader's guard tables: %v", err)
		return
	}
	t.Logf("the loader's guard tables (table|rows): %s", strings.Join(lines, ", "))
}

// checkSeeded checks that the seed put something in every kind of record the test promises to keep, so a loader
// that stored nothing can't make the upgrade's checks pass: accounts, transactions, lots, valuations and attention
// items, the script tokens, and that the published release keeps no audit log yet (it begins with the new one).
func checkSeeded(t *testing.T, k *kube, s *snapshot) {
	t.Helper()
	for _, table := range []string{
		"ledger.accounts", "ledger.transaction_versions", "ledger.postings", "valuation.holding_valuations",
		"attention.items", "attention.changes", "identity.script_tokens", "portfolio.sleeves", "ingest.tickers", "ingest.ticker_entries",
	} {
		if s.count(table) == 0 {
			t.Errorf("the seeded install holds no rows in %s", table)
		}
	}
	lots := queryOne(t, k, `SELECT (SELECT count(*) FROM ledger.postings WHERE book_kind = 'holding')
		+ (SELECT count(*) FROM ledger.transaction_versions WHERE kind = 'opening-lot')`)
	if lots == "0" {
		t.Error("the seeded install holds no lots (holding postings or opening lots)")
	}
	if got := queryOne(t, k, `SELECT to_regclass('identity.audit_log') IS NULL`); got != "t" {
		t.Errorf("the published release already has an audit log (identity.audit_log); the test expects none until the upgrade, got %q", got)
	}
	t.Logf("seeded: %d accounts, %d transaction versions, %d postings (%s lots), %d valuations, %d attention items (%d changes), %d script tokens",
		s.count("ledger.accounts"), s.count("ledger.transaction_versions"), s.count("ledger.postings"), lots,
		s.count("valuation.holding_valuations"), s.count("attention.items"), s.count("attention.changes"), s.count("identity.script_tokens"))
}

// upgradeAsGuided upgrades the published release to the candidate bundle with the install guide's "Upgrade"
// section: CloudNativePG's manifest when the candidate ships another, the new CRDs, the new database manifest, the
// new install, a wait for every rollout, and the built-in UI upgraded with the CLI built from this checkout.
func upgradeAsGuided(t *testing.T, k *kube, published, candidate *bundle) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), upgradeTimeout)
	defer cancel()
	for _, step := range []func() error{
		func() error {
			if bytes.Equal(published.cnpg, candidate.cnpg) {
				return nil
			}
			if err := kubectlApply(ctx, t.Logf, "the candidate's CloudNativePG", candidate.cnpg); err != nil {
				return err
			}
			return waitFor(ctx, rolloutTimeout, k.deploymentAvailable("cnpg-system", "cnpg-controller-manager"))
		},
		func() error { return kubectlApply(ctx, t.Logf, "the candidate's CRDs", candidate.crds) },
		func() error { return kubectlApply(ctx, t.Logf, "the candidate's database", candidate.database) },
		func() error { return waitDatabase(ctx, k, t.Logf) },
		func() error { return kubectlApply(ctx, t.Logf, "the candidate's install", candidate.install) },
		func() error { return waitDeployments(ctx, k, candidate) },
	} {
		if err := step(); err != nil {
			t.Fatalf("upgrading: %v", err)
		}
	}
	waitCoreAvailable(t, k)
	// The UI still runs the published release's image until the owner upgrades it; the core and the operator of the
	// new release serve it meanwhile.
	waitPluginReady(t, k, uiPlugin)
	reg, err := registryFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	ui, err := reg.image(t.Context(), uiImage)
	if err != nil {
		t.Fatal(err)
	}
	upgrade := exec.CommandContext(ctx, cliPath, "plugin", "upgrade", uiPlugin)
	// The version is the suite's own image, by digest: the run's catalog pins nothing of the published release.
	upgrade.Args = append(upgrade.Args, ui.cluster.String(), "--yes")
	cliRun(t, upgrade)
	p := waitPluginReady(t, k, uiPlugin)
	if p.Spec.Image != ui.cluster.String() {
		t.Errorf("plugin %s runs %s after its upgrade, want %s", uiPlugin, p.Spec.Image, ui.cluster)
	}
}

// checkStoredData checks D-161's promise: after the upgrade every row the published release stored is still there
// and unchanged, in every table but the few that are derived or ephemeral (snapshot_test.go says which, and why),
// and the append-only attention log only grew. The new core derives its valuations again on its first start, so the
// check repeats until it holds or the deadline passes.
func checkStoredData(t *testing.T, k *kube, before *snapshot) {
	t.Helper()
	var diffs []string
	err := waitFor(t.Context(), settleTimeout, func(ctx context.Context) (string, bool, error) {
		after, err := takeSnapshot(ctx, k, before.tables)
		if err != nil {
			return "", false, fmt.Errorf("%w: %w", errRetry, err)
		}
		diffs = before.compare(after)
		return fmt.Sprintf("%d differences", len(diffs)), len(diffs) == 0, nil
	})
	if err != nil {
		t.Fatalf("the upgrade changed or lost stored data: %v\n%s", err, strings.Join(diffs, "\n"))
	}
}

// checkTickerConfirmations checks the one stored change that moves data between tables: the confirmation of the
// seeded ticker, which ingest 0015 moves from ingest.tickers to ingest.ticker_confirmations (D-153), is there once
// and as the seed made it, and the unconfirmed ticker has none.
func checkTickerConfirmations(t *testing.T, k *kube) {
	t.Helper()
	got := queryLines(t, k, `SELECT institution || '|' || symbol || '|' || version || '|' || isin || '|' || country || '|' || proposal_currency || '|' || picked
		FROM ingest.ticker_confirmations ORDER BY 1`)
	if want := []string{confirmedTicker}; !slices.Equal(got, want) {
		t.Errorf("ingest.ticker_confirmations holds %v after the upgrade, want %v", got, want)
	}
}

// checkGrantCutover checks D-209: the cutover ran once, at the new core's first start, and left the household
// grant to the script token still in use and to the installed view plugin, each once, and none to the token that
// was revoked; each grant has its audit entry, made by the core.
func checkGrantCutover(t *testing.T, k *kube, s seededInstall) {
	t.Helper()
	want := []string{
		"script_token|" + s.kept.id + "|household|" + householdID + "|t",
		"view_plugin|" + uiPlugin + "|household|" + householdID + "|t",
	}
	slices.Sort(want)
	if got := queryOne(t, k, `SELECT count(*) FROM identity.grant_cutover`); got != "1" {
		t.Errorf("identity.grant_cutover holds %s rows, want the cutover's one", got)
	}
	got := queryLines(t, k, `SELECT holder_kind || '|' || holder_id || '|' || target_kind || '|' || target_id || '|'
		|| CASE WHEN revoked_at IS NULL THEN 't' ELSE 'f' END
		FROM identity.grants ORDER BY 1`)
	if !slices.Equal(got, want) {
		t.Errorf("the grants after the upgrade are\n%s\nwant exactly\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
	audit := queryLines(t, k, `SELECT action || '|' || actor_kind || '|' || actor_id || '|' || outcome || '|' || count(*)
		FROM identity.audit_log WHERE action IN ('grant_created', 'grant_revoked') GROUP BY action, actor_kind, actor_id, outcome ORDER BY 1`)
	if wantAudit := []string{"grant_created|system|core|ok|2"}; !slices.Equal(audit, wantAudit) {
		t.Errorf("the audit log's grant entries are %v, want %v", audit, wantAudit)
	}
}

// checkSpiceDBCredentials checks D-216: the Job of the new release completed, and the two Secrets it makes exist,
// with generated values of a safe length, one apart from the other. No value is logged.
func checkSpiceDBCredentials(t *testing.T, k *kube) {
	t.Helper()
	waitSpiceDBAvailable(t, k)
	waitJobComplete(t, k, credentialsJobName(t, k))
	key := secretValue(t, k, spicedbPresharedName, "key")
	password := secretValue(t, k, spicedbRoleName, "password")
	for name, value := range map[string][]byte{spicedbPresharedName: key, spicedbRoleName: password} {
		if len(value) < minGenerated {
			t.Errorf("Secret %s: the generated value has %d bytes, want at least %d", name, len(value), minGenerated)
		}
	}
	if bytes.Equal(key, password) {
		t.Errorf("Secrets %s and %s hold the same value", spicedbPresharedName, spicedbRoleName)
	}
}

// checkCutoverInSpiceDB checks that after the upgrade SpiceDB answers, by default (D-164), and holds what the cutover
// gave (D-209): the core runs SpiceDB and has reconciled with it, and the script token issued before the upgrade and
// the UI plugin installed before it, each given the household grant by the cutover, list every account the published
// release stored, which only SpiceDB's household relationships for them can answer: under SpiceDB a holder with no
// grant there views nothing.
func checkCutoverInSpiceDB(t *testing.T, k *kube, s seededInstall) {
	t.Helper()
	d, err := k.clients.AppsV1().Deployments(namespace).Get(t.Context(), coreDeployment, metav1.GetOptions{})
	if err != nil {
		t.Fatalf("getting Deployment %s: %v", coreDeployment, err)
	}
	if !coreRunsSpiceDB(d) {
		t.Fatalf("the upgraded core runs without SpiceDB: the bundle's core must run %s (D-164)", authzSpiceDBFlag)
	}
	waitCoreAvailable(t, k)
	waitReconciled(t, k)
	stored := queryLines(t, k, `SELECT id FROM ledger.accounts ORDER BY 1`)
	if len(stored) == 0 {
		t.Fatal("the published release stored no account to check the grants against")
	}
	slices.Sort(stored) // as the lists below are, whatever the database's collation
	f := forwardCoreUntilStopped(t, k)
	defer f.stop()
	if got := sortedAccounts(t, bundleGet(t, f.base+"/v1/accounts", s.kept.secret)); !slices.Equal(got, stored) {
		t.Errorf("the script token issued before the upgrade lists %v under SpiceDB, want every stored account %v", got, stored)
	}
	owner := newOwnerAPI(t, k)
	ui := k.token(t, plugins.ServiceAccountName(uiPlugin), plugins.CoreAudience)
	err = waitFor(t.Context(), spicedbTimeout, func(ctx context.Context) (string, bool, error) {
		ids, listErr := owner.accountsOf(ctx, ui)
		if listErr != nil {
			return "", false, listErr
		}
		return fmt.Sprintf("%d accounts", len(ids)), slices.Equal(ids, stored), nil
	})
	if err != nil {
		t.Errorf("waiting for the UI plugin, installed before the upgrade, to list every stored account under SpiceDB: %v", err)
	}
}

// sortedAccounts returns the IDs of the accounts in a 200 answer of GET /v1/accounts, sorted, or fails t.
func sortedAccounts(t *testing.T, r reply) []string {
	t.Helper()
	if r.status != http.StatusOK {
		t.Fatalf("GET /v1/accounts: status %d, body %q; want 200", r.status, r.body)
	}
	var accounts viewsv1.ListAccountsResponse
	if err := protojson.Unmarshal(r.body, &accounts); err != nil {
		t.Fatalf("decoding the accounts: %v", err)
	}
	ids := make([]string, 0, len(accounts.GetAccounts()))
	for _, a := range accounts.GetAccounts() {
		ids = append(ids, a.GetId())
	}
	slices.Sort(ids)
	return ids
}

// checkAgenticSwitchOff checks D-212: agents and models are off, on a fresh install and after an upgrade: the switch
// has no row that turns it on, no feature toggle is set, and the MCP endpoint doesn't answer even to a valid script
// token.
func checkAgenticSwitchOff(t *testing.T, k *kube, token issuedToken) {
	t.Helper()
	if got := queryOne(t, k, `SELECT count(*) FROM modelrouter.switch WHERE enabled`); got != "0" {
		t.Errorf("the agentic switch is on: %s rows, want none", got)
	}
	if got := queryOne(t, k, `SELECT count(*) FROM modelrouter.feature_toggles`); got != "0" {
		t.Errorf("%s feature toggles are on, want none", got)
	}
	f := forwardCoreUntilStopped(t, k)
	defer f.stop()
	r := postMCP(t, f.base, token.secret, `{"jsonrpc":"2.0","id":1,"method":"tools/list"}`)
	if r.status != http.StatusNotFound {
		t.Errorf("POST /mcp with the switch off: status %d, want %d (no MCP endpoint answers)", r.status, http.StatusNotFound)
	}
}

// checkServing checks that the upgraded core is ready, that the script token issued before the upgrade still lists
// accounts that were seeded, that the revoked one is still refused, and that the UI, upgraded with the CLI, serves
// its login page through the core.
func checkServing(t *testing.T, k *kube, s seededInstall) {
	t.Helper()
	waitCoreAvailable(t, k)
	f := forwardCoreUntilStopped(t, k)
	defer f.stop()
	base := f.base
	if r := bundleGet(t, base+"/readyz", ""); r.status != http.StatusOK {
		t.Fatalf("GET /readyz after the upgrade: status %d, body %q; want 200", r.status, r.body)
	}
	checkAccountsListed(t, k, base, s.kept.secret)
	if r := bundleGet(t, base+"/v1/accounts", s.revoked.secret); r.status != http.StatusUnauthorized {
		t.Errorf("GET /v1/accounts with the revoked token: status %d, want 401", r.status)
	}
	err := waitFor(t.Context(), traefikTimeout, func(ctx context.Context) (string, bool, error) {
		login, getErr := fetch(ctx, base+"/login")
		if getErr != nil {
			return "", false, fmt.Errorf("%w: %w", errRetry, getErr)
		}
		return fmt.Sprintf("status %d", login.status), login.status == http.StatusOK && bytes.Contains(login.body, []byte(loginButton)), nil
	})
	if err != nil {
		t.Fatalf("GET /login through the core after the upgrade: %v; want 200 with %q", err, loginButton)
	}
}

// checkAccountsListed checks that token, issued before the upgrade, lists accounts, and only accounts that were
// seeded.
func checkAccountsListed(t *testing.T, k *kube, base, token string) {
	t.Helper()
	r := bundleGet(t, base+"/v1/accounts", token)
	if r.status != http.StatusOK {
		t.Fatalf("GET /v1/accounts with the token issued before the upgrade: status %d, body %q; want 200", r.status, r.body)
	}
	var accounts viewsv1.ListAccountsResponse
	if err := protojson.Unmarshal(r.body, &accounts); err != nil {
		t.Fatalf("decoding the accounts: %v", err)
	}
	stored := queryLines(t, k, `SELECT id FROM ledger.accounts ORDER BY 1`)
	if len(accounts.GetAccounts()) == 0 {
		t.Error("the read API lists no account after the upgrade")
	}
	for _, a := range accounts.GetAccounts() {
		if !slices.Contains(stored, a.GetId()) {
			t.Errorf("the read API lists an account that was not seeded: %s", a.GetId())
		}
	}
}

// fetch sends GET url without a token and returns the status and the body.
func fetch(ctx context.Context, url string) (reply, error) {
	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return reply{}, fmt.Errorf("GET %s: %w", url, err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return reply{}, fmt.Errorf("GET %s: %w", url, err)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, bundleBodyLimit))
	if closeErr := resp.Body.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return reply{}, fmt.Errorf("GET %s: reading the body: %w", url, err)
	}
	return reply{status: resp.StatusCode, body: body}, nil
}

// checkCutoverRanOnce checks that the cutover runs once only: a script token issued after it holds no grant, and a
// restart of the core, whose start would run the cutover again if it could, changes no grant, no cutover record
// and no audit entry of a grant.
func checkCutoverRanOnce(t *testing.T, k *kube) {
	t.Helper()
	later := issueTokenWith(t, k)
	state := func() string {
		return strings.Join(queryLines(t, k, `SELECT 'cutover ' || done_at FROM identity.grant_cutover
			UNION ALL SELECT 'grant ' || id || ' ' || holder_kind || ' ' || holder_id || ' ' || target_kind || ' ' || created_at || ' ' || coalesce(revoked_at::text, '')
				FROM identity.grants
			UNION ALL SELECT 'audit ' || id || ' ' || action FROM identity.audit_log WHERE action IN ('grant_created', 'grant_revoked')
			ORDER BY 1`), "\n")
	}
	before := state()
	for _, replicas := range []int32{0, 1} {
		if err := scaleDeployment(t.Context(), k, coreDeployment, replicas); err != nil {
			t.Fatalf("restarting the core: %v", err)
		}
	}
	waitCoreAvailable(t, k)
	if after := state(); after != before {
		t.Errorf("a restart changed the grants or their audit entries:\nbefore\n%s\nafter\n%s", before, after)
	}
	if got := queryOne(t, k, `SELECT count(*) FROM identity.grants WHERE holder_kind = 'script_token' AND holder_id = '`+later.id+`'`); got != "0" {
		t.Errorf("the script token issued after the cutover holds %s grants, want none", got)
	}
}
