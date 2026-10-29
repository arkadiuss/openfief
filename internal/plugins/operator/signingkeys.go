package operator

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/arkadiuss/openfief/internal/plugins"
)

// The API server's ServiceAccount issuer discovery endpoints. They are
// non-resource URLs, which no namespaced Role can grant: Kubernetes' default
// RBAC binds the ClusterRole system:service-account-issuer-discovery, which
// allows GET on both, to the group system:serviceaccounts, so the operator's
// ServiceAccount may read them without a ClusterRole of OpenFief's (D-15).
const (
	openIDConfigPath = "/.well-known/openid-configuration"
	jwksPath         = "/openid/v1/jwks"
	// DefaultSigningKeysRefresh is how often the operator reads the keys
	// again, so a rotated key reaches plugins.
	DefaultSigningKeysRefresh = 10 * time.Minute
	signingKeysFieldOwner     = FieldOwner + "/signing-keys"
)

// IssuerDiscovery reads the cluster's ServiceAccount token issuer and public
// signing keys from the API server's issuer discovery endpoints. The JWKS is
// always read from the API server itself: the jwks_uri the configuration
// names may be an address the operator can't reach.
type IssuerDiscovery struct {
	// Client is a REST client of the API server, such as a discovery
	// client's RESTClient().
	Client rest.Interface
}

// SigningKeys fetches and checks the issuer and the keys.
func (d IssuerDiscovery) SigningKeys(ctx context.Context) (plugins.SigningKeys, error) {
	cfg, err := d.Client.Get().AbsPath(openIDConfigPath).DoRaw(ctx)
	if err != nil {
		return plugins.SigningKeys{}, fmt.Errorf("read %s: %w", openIDConfigPath, err)
	}
	jwks, err := d.Client.Get().AbsPath(jwksPath).DoRaw(ctx)
	if err != nil {
		return plugins.SigningKeys{}, fmt.Errorf("read %s: %w", jwksPath, err)
	}
	keys, err := plugins.ParseSigningKeys(cfg, jwks)
	if err != nil {
		return plugins.SigningKeys{}, fmt.Errorf("signing keys: %w", err)
	}
	return keys, nil
}

// SigningKeySource returns the cluster's current signing keys.
type SigningKeySource interface {
	SigningKeys(ctx context.Context) (plugins.SigningKeys, error)
}

// SigningKeysPublisher publishes the cluster's ServiceAccount signing keys
// and issuer in the plugins.SigningKeysConfigMap ConfigMap, which every
// plugin mounts to verify the core's token (D-59). It publishes them when
// the operator starts, every Refresh, and whenever the ConfigMap changes, so
// an edited or deleted ConfigMap is put back. Keys it can't read or check
// leave the published ones in place.
type SigningKeysPublisher struct {
	Client  client.Client
	Source  SigningKeySource
	Refresh time.Duration
}

// Reconcile publishes the current keys and asks to run again after Refresh.
func (p *SigningKeysPublisher) Reconcile(ctx context.Context, _ ctrl.Request) (ctrl.Result, error) {
	if err := p.Publish(ctx); err != nil {
		return ctrl.Result{}, err
	}
	refresh := p.Refresh
	if refresh <= 0 {
		refresh = DefaultSigningKeysRefresh
	}
	return ctrl.Result{RequeueAfter: refresh}, nil
}

// Publish reads the keys from Source and applies the ConfigMap.
func (p *SigningKeysPublisher) Publish(ctx context.Context) error {
	keys, err := p.Source.SigningKeys(ctx)
	if err != nil {
		return fmt.Errorf("publish signing keys: %w", err)
	}
	return applyAs(ctx, p.Client, SigningKeysConfigMap(keys), signingKeysFieldOwner)
}

// SigningKeysConfigMap is the plugins.SigningKeysConfigMap ConfigMap holding
// keys.
func SigningKeysConfigMap(keys plugins.SigningKeys) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: kindConfigMap},
		ObjectMeta: metav1.ObjectMeta{
			Name: plugins.SigningKeysConfigMap, Namespace: plugins.Namespace,
			Labels: map[string]string{labelManagedBy: managerName},
		},
		Data: keys.Data(),
	}
}

// SetupWithManager registers the publisher: it runs once at start, then on
// every change to the ConfigMap and after each Refresh.
func (p *SigningKeysPublisher) SetupWithManager(mgr ctrl.Manager) error {
	start := make(chan event.GenericEvent, 1)
	start <- event.GenericEvent{Object: &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{
		Name: plugins.SigningKeysConfigMap, Namespace: plugins.Namespace,
	}}}
	isKeys := predicate.NewPredicateFuncs(func(o client.Object) bool {
		return o.GetNamespace() == plugins.Namespace && o.GetName() == plugins.SigningKeysConfigMap
	})
	err := ctrl.NewControllerManagedBy(mgr).
		Named("signing-keys").
		For(&corev1.ConfigMap{}, builder.WithPredicates(isKeys)).
		WatchesRawSource(source.Channel(start, &handler.EnqueueRequestForObject{})).
		Complete(p)
	if err != nil {
		return fmt.Errorf("set up signing keys publisher: %w", err)
	}
	return nil
}
