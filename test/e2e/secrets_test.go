//go:build e2e

package e2e

import (
	"bytes"
	"fmt"
	"maps"
	"slices"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/operator"
)

// The classes of pod the secrets audit tells apart, and the database
// cluster's name, which CloudNativePG gives every Secret and pod it makes.
const (
	podCore     = "app:openfief-core"
	podSeedJob  = "app:demo-seed"
	podDatabase = "cnpg"
	podSpiceDB  = "app:openfief-spicedb"
	// minSecretValue is the shortest Secret value the audit looks for
	// elsewhere: a shorter one would match by chance.
	minSecretValue = 8
	// appLabel names the pod's application, as every deploy/ manifest sets it.
	appLabel = "app.kubernetes.io/name"
	// databaseLabel is the label CloudNativePG puts on its pods.
	databaseLabel = "cnpg.io/cluster"
)

// secretRule declares one Secret of the openfief namespace: what each of its
// keys is, and which pods may and must use it. The audit fails on a Secret
// with no rule and on a key a rule doesn't name, so a new credential has to
// be declared here, with who may read it, before the suite passes.
type secretRule struct {
	// sensitive are the keys whose values must be nowhere but the Secret;
	// public are keys that hold nothing secret, such as a certificate, a
	// host name or the database's name.
	sensitive, public []string
	// may are the classes of pod that may use the Secret; must those that
	// have to (they are in may too).
	may, must []string
	// inDatabase is whether the audit also looks for the sensitive values in
	// a dump of the database, which the connectors' credentials must never
	// reach. The database's own credentials are no such case.
	inDatabase bool
}

// fixedSecretRules are the Secrets the install makes itself, by name:
// CloudNativePG's for the openfief-db cluster, and SpiceDB's two, which
// deploy/spicedb's Job generates. The operator, the egress proxy, the
// built-in UI and every plugin use none of them (D-200 criterion 14).
func fixedSecretRules() map[string]secretRule {
	return map[string]secretRule{
		// The application's credentials: the core's connection string (deploy/core)
		// and the demo-seed Job's (deploy/demo-data), and CloudNativePG's own
		// pods, which read the role's password.
		"openfief-db-app": {
			sensitive: []string{"password", "uri", "jdbc-uri", "fqdn-uri", "fqdn-jdbc-uri", "pgpass"},
			public:    []string{"username", "user", "dbname", "host", "port"},
			may:       []string{podCore, podSeedJob, podDatabase}, must: []string{podCore},
		},
		// The cluster's certificate authority, and the server's and the
		// replication client's certificates it signed: only its own pods.
		// SpiceDB's (D-164): the preshared key, which only SpiceDB and the core, its one caller, read, and
		// the password of SpiceDB's role in the cluster, which only SpiceDB reads; the audit also finds
		// neither in a ConfigMap, a pod spec or a Plugin. The Job that made them uses no Secret.
		"openfief-spicedb-key": {
			sensitive: []string{"key"}, may: []string{podSpiceDB, podCore}, must: []string{podSpiceDB, podCore},
		},
		"openfief-db-spicedb": {
			sensitive: []string{"password"}, public: []string{"username"}, may: []string{podSpiceDB}, must: []string{podSpiceDB},
		},
		databaseCluster + "-ca": {
			sensitive: []string{"ca.key"}, public: []string{"ca.crt"}, may: []string{podDatabase},
		},
		databaseCluster + "-server": {
			sensitive: []string{"tls.key"}, public: []string{"ca.crt", "tls.crt"}, may: []string{podDatabase},
		},
		databaseCluster + "-replication": {
			sensitive: []string{"tls.key"}, public: []string{"ca.crt", "tls.crt"}, may: []string{podDatabase},
		},
	}
}

// ruleFor returns the rule for s: a fixed Secret's, or, for a Secret that
// names its instance in plugins.OwnerInstanceLabel and is called
// <instance>--<role>, an instance Secret's: only that instance's pods may use
// it, and every file but a public one is sensitive. mounted
// holds the instance Secrets the caller knows an instance's pods mount.
func ruleFor(s *corev1.Secret, mounted map[string]string) (secretRule, bool) {
	if r, ok := fixedSecretRules()[s.Name]; ok {
		return r, true
	}
	instance, ok := s.Labels[plugins.OwnerInstanceLabel]
	if !ok || !strings.HasPrefix(s.Name, instance+plugins.SecretNameSeparator) {
		return secretRule{}, false
	}
	class := "plugin:" + instance
	r := secretRule{may: []string{class}, inDatabase: true}
	// The run's CA certificate, which the plugin trusts, and the Flex query
	// ID, which picks a statement and is no credential, are public; every
	// other file is a credential.
	r.public = []string{plugins.CABundleKey, flexQueryFile}
	for key := range s.Data {
		if !slices.Contains(r.public, key) {
			r.sensitive = append(r.sensitive, key)
		}
	}
	slices.Sort(r.sensitive)
	if mounted[s.Name] == class {
		r.must = []string{class}
	}
	return r, true
}

// podClass names the kind of pod p is: CloudNativePG's, one of the
// install's own by its application label, or a plugin instance's.
func podClass(p *corev1.Pod) string {
	switch {
	case p.Labels[databaseLabel] == databaseCluster:
		return podDatabase
	case p.Labels[operator.LabelPlugin] != "":
		return "plugin:" + p.Labels[operator.LabelPlugin]
	case p.Labels[appLabel] != "":
		return "app:" + p.Labels[appLabel]
	}
	return "pod:" + p.Name
}

// secretsReferencedBy returns the name of every Secret p refers to, from a
// volume (whole or projected), an environment variable, envFrom or an image
// pull Secret, in any of its containers.
func secretsReferencedBy(p *corev1.Pod) []string {
	names := map[string]bool{}
	for i := range p.Spec.Volumes {
		for _, n := range volumeSecretNames(&p.Spec.Volumes[i]) {
			names[n] = true
		}
	}
	containers := slices.Concat(p.Spec.InitContainers, p.Spec.Containers)
	for i := range containers {
		for _, n := range containerSecretNames(&containers[i]) {
			names[n] = true
		}
	}
	for _, r := range p.Spec.ImagePullSecrets {
		names[r.Name] = true
	}
	return slices.Sorted(maps.Keys(names))
}

// volumeSecretNames names the Secrets v mounts, whole or as projected sources.
func volumeSecretNames(v *corev1.Volume) []string {
	var names []string
	if v.Secret != nil {
		names = append(names, v.Secret.SecretName)
	}
	if v.Projected != nil {
		for _, s := range v.Projected.Sources {
			if s.Secret != nil {
				names = append(names, s.Secret.Name)
			}
		}
	}
	return names
}

// containerSecretNames names the Secrets c takes environment variables from.
func containerSecretNames(c *corev1.Container) []string {
	var names []string
	for _, e := range c.EnvFrom {
		if e.SecretRef != nil {
			names = append(names, e.SecretRef.Name)
		}
	}
	for _, e := range c.Env {
		if e.ValueFrom != nil && e.ValueFrom.SecretKeyRef != nil {
			names = append(names, e.ValueFrom.SecretKeyRef.Name)
		}
	}
	return names
}

// mountedSecrets maps each Secret the data path's instances declare to the
// class of the instance's pods, which must mount it.
func mountedSecrets(instances []dataPathInstance) map[string]string {
	m := map[string]string{}
	for i := range instances {
		in := &instances[i]
		for role := range in.secrets {
			m[plugins.SecretName(in.name, role)] = "plugin:" + in.name
		}
	}
	return m
}

// sensitiveValue is one sensitive value of a Secret, to look for elsewhere.
type sensitiveValue struct {
	name, key  string
	value      []byte
	inDatabase bool
}

// checkSecretsAudit checks D-200 criterion 14 over the whole openfief
// namespace, not one Secret: every Secret has a rule (fixedSecretRules and
// ruleFor), every key of it is classified, and only the pods its rule allows
// use it, with those it requires doing so, whether or not the Secret exists
// (an optional Secret a pod names but nobody created must still be that
// pod's own instance's). Every sensitive value of eight bytes or more is then
// absent from every ConfigMap of the cluster, every pod spec of the
// namespace (environment, arguments, command: the fixture's, in its own
// namespace, takes the synthetic Flex and EODHD tokens as arguments, being
// the institution), every Plugin and PluginGrant, and, for the connectors'
// credentials, a dump of the database. Failures name Secrets, keys, pods and
// places, never a value. mounted holds the instance Secrets (name to pod
// class) that must be mounted.
func checkSecretsAudit(t *testing.T, k *kube, mounted map[string]string) {
	t.Helper()
	secrets, err := k.clients.CoreV1().Secrets(namespace).List(t.Context(), metav1.ListOptions{})
	if err != nil {
		t.Fatalf("listing Secrets: %v", err)
	}
	pods, err := k.clients.CoreV1().Pods(namespace).List(t.Context(), metav1.ListOptions{})
	if err != nil {
		t.Fatalf("listing pods: %v", err)
	}
	users := podUsers(pods.Items)
	existing := map[string]bool{}
	var values []sensitiveValue
	for i := range secrets.Items {
		s := &secrets.Items[i]
		existing[s.Name] = true
		rule, ok := ruleFor(s, mounted)
		if !ok {
			t.Errorf("Secret %s has no rule: declare it in test/e2e/secrets_test.go (who may use it, which keys are sensitive), type %s, labels %v, keys %v",
				s.Name, s.Type, slices.Sorted(maps.Keys(s.Labels)), slices.Sorted(maps.Keys(s.Data)))
			continue
		}
		problems, found := auditSecret(s, &rule, users[s.Name])
		for _, p := range problems {
			t.Error(p)
		}
		values = append(values, found...)
	}
	for _, p := range missingSecretProblems(users, existing) {
		t.Error(p)
	}
	for _, p := range k.sensitiveElsewhere(t, values) {
		t.Error(p)
	}
	t.Logf("e2e: audited %d Secrets, %d sensitive values, %d pods", len(secrets.Items), len(values), len(pods.Items))
}

// podUsers maps each Secret name a live pod refers to the classes of those pods.
func podUsers(pods []corev1.Pod) map[string][]string {
	users := map[string][]string{}
	for i := range pods {
		p := &pods[i]
		if p.DeletionTimestamp != nil {
			continue
		}
		class := podClass(p)
		for _, name := range secretsReferencedBy(p) {
			if !slices.Contains(users[name], class) {
				users[name] = append(users[name], class)
			}
		}
	}
	return users
}

// auditSecret checks s against its rule: every key classified, its users
// allowed and the required ones among them. It returns the problems and the
// sensitive values worth looking for elsewhere.
func auditSecret(s *corev1.Secret, rule *secretRule, used []string) ([]string, []sensitiveValue) {
	var problems []string
	var values []sensitiveValue
	for _, key := range slices.Sorted(maps.Keys(s.Data)) {
		switch {
		case slices.Contains(rule.sensitive, key):
			if v := s.Data[key]; len(v) >= minSecretValue {
				values = append(values, sensitiveValue{s.Name, key, v, rule.inDatabase})
			}
		case !slices.Contains(rule.public, key):
			problems = append(problems, fmt.Sprintf("Secret %s has key %q that its rule classifies as neither sensitive nor public", s.Name, key))
		}
	}
	got := slices.Sorted(slices.Values(used))
	for _, class := range got {
		if !slices.Contains(rule.may, class) {
			problems = append(problems, fmt.Sprintf("Secret %s is used by %s pods; its rule allows only %v", s.Name, class, rule.may))
		}
	}
	for _, class := range rule.must {
		if !slices.Contains(got, class) {
			problems = append(problems, fmt.Sprintf("Secret %s is not used by %s pods, which must use it; used by %v", s.Name, class, got))
		}
	}
	return problems, values
}

// missingSecretProblems reports pods that use a Secret that doesn't exist
// yet or at all, unless it is their own instance's, <instance>--<role>: an
// optional Secret nobody created.
func missingSecretProblems(users map[string][]string, existing map[string]bool) []string {
	var problems []string
	for _, name := range slices.Sorted(maps.Keys(users)) {
		if existing[name] {
			continue
		}
		for _, class := range users[name] {
			if instance, isPlugin := strings.CutPrefix(class, "plugin:"); !isPlugin || !strings.HasPrefix(name, instance+plugins.SecretNameSeparator) {
				problems = append(problems, fmt.Sprintf("%s pods use Secret %s, which doesn't exist and isn't their instance's", class, name))
			}
		}
	}
	return problems
}

// sensitiveElsewhere returns a problem for each value found in a ConfigMap,
// a pod spec of the namespace, a Plugin or PluginGrant or, when its rule says
// so, the database.
func (k *kube) sensitiveElsewhere(t *testing.T, values []sensitiveValue) []string {
	t.Helper()
	var problems []string
	var dump []byte
	for _, v := range values {
		var found []string
		found = append(found, k.configMapsHolding(t, v.value)...)
		found = append(found, k.podsHoldingIn(t, namespace, v.value)...)
		found = append(found, k.pluginsHolding(t, v.value)...)
		if v.inDatabase {
			if dump == nil {
				dump = dumpDatabase(t, k)
			}
			if bytes.Contains(dump, v.value) {
				found = append(found, "the database")
			}
		}
		if len(found) > 0 {
			problems = append(problems, fmt.Sprintf("the value of Secret %s, key %s, is outside its Secret, in: %s", v.name, v.key, strings.Join(found, "; ")))
		}
	}
	return problems
}
