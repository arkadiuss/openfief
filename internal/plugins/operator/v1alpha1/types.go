package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// Condition types on a Plugin's status.
const (
	// ConditionReady is True when the plugin's workload is applied and its
	// Deployment is available.
	ConditionReady = "Ready"
	// ConditionApprovalRequired is True when the Plugin requests permissions
	// its PluginGrant does not approve; status.pendingApproval lists them.
	ConditionApprovalRequired = "ApprovalRequired"
	// ConditionSignatureVerified reports the cosign verification of the
	// Plugin's image (D-17).
	ConditionSignatureVerified = "SignatureVerified"
	// ConditionValid, on a PluginGrant, is True when the grant is well-formed.
	ConditionValid = "Valid"
)

// Plugin is what a plugin's vendor requests: a digest-pinned image, a
// manifest of permissions and a config (D-13, D-16). The operator runs it
// only with what its PluginGrant of the same name approves.
//
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories=openfief
// +kubebuilder:validation:XValidation:rule="self.metadata.name.size() <= 56 && self.metadata.name.matches('^[a-z]([-a-z0-9]*[a-z0-9])?$')",message="name must be a lower-case DNS label of at most 56 characters that starts with a letter"
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`
// +kubebuilder:printcolumn:name="Approval required",type=string,JSONPath=`.status.conditions[?(@.type=="ApprovalRequired")].status`
// +kubebuilder:printcolumn:name="Signature",type=string,JSONPath=`.status.conditions[?(@.type=="SignatureVerified")].status`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type Plugin struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PluginSpec   `json:"spec"`
	Status PluginStatus `json:"status,omitempty"`
}

// PluginSpec is one version of a plugin.
type PluginSpec struct {
	// Image is the plugin's container image, pinned by digest:
	// repository@sha256:<64 hex digits>. Tags are refused (D-17).
	// +kubebuilder:validation:MaxLength=320
	// +kubebuilder:validation:Pattern=`^[^@]+@sha256:[0-9a-f]{64}$`
	Image string `json:"image"`

	// Manifest is what the plugin requests.
	Manifest Manifest `json:"manifest"`

	// Config is the plugin's configuration, a JSON object checked against
	// manifest.configSchema. It must not hold credentials: those belong in
	// Secrets (D-29).
	// +optional
	// +kubebuilder:validation:Type=object
	// +kubebuilder:pruning:PreserveUnknownFields
	Config *runtime.RawExtension `json:"config,omitempty"`

	// ClusterService is the in-cluster service the owner picked at install
	// for a manifest that asks for one (D-74). The plugin reaches it
	// directly at its ClusterIP, not through the egress proxy.
	// +optional
	ClusterService *ClusterServiceRef `json:"clusterService,omitempty"`
}

// ClusterServiceRef names a Service port in another namespace than
// openfief (D-74).
type ClusterServiceRef struct {
	// Namespace is the Service's namespace; never openfief.
	// +kubebuilder:validation:MaxLength=63
	// +kubebuilder:validation:Pattern=`^[a-z]([-a-z0-9]*[a-z0-9])?$`
	Namespace string `json:"namespace"`
	// Service is the Service's name.
	// +kubebuilder:validation:MaxLength=63
	// +kubebuilder:validation:Pattern=`^[a-z]([-a-z0-9]*[a-z0-9])?$`
	Service string `json:"service"`
	// Port is the Service port the plugin dials.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	Port int32 `json:"port"`
}

// ClusterServiceRequest is a manifest's request for one in-cluster service
// that the owner picks at install (D-74).
type ClusterServiceRequest struct {
	// Description says what the service is for, such as "the owner's
	// Ghostfolio"; the CLI shows it when it asks for the service.
	// +optional
	// +kubebuilder:validation:MaxLength=200
	Description string `json:"description,omitempty"`
}

// Manifest is what a plugin declares (D-11).
type Manifest struct {
	// Capabilities are what the plugin acts as.
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:items:Enum=connector;view
	// +listType=set
	Capabilities []string `json:"capabilities"`

	// Egress lists the hosts the plugin reaches through the egress proxy, as
	// "host" (port 443) or "host:port". Never IP addresses or wildcards.
	// +optional
	// +kubebuilder:validation:MaxItems=64
	// +kubebuilder:validation:items:MaxLength=259
	// +listType=set
	Egress []string `json:"egress,omitempty"`

	// Secrets lists, by role, the Kubernetes Secrets the plugin needs
	// mounted read-only. Each instance mounts the Secret named
	// <instance>-<role> in the openfief namespace, so two instances never
	// share one (D-89). The owner creates them (D-29).
	// +optional
	// +kubebuilder:validation:MaxItems=32
	// +listType=map
	// +listMapKey=role
	Secrets []SecretMount `json:"secrets,omitempty"`

	// DataTiers lists the classes of data the plugin asks to see.
	// +optional
	// +kubebuilder:validation:MaxItems=32
	// +kubebuilder:validation:items:MaxLength=32
	// +listType=set
	DataTiers []string `json:"dataTiers,omitempty"`

	// ConfigSchema is the JSON Schema that spec.config must satisfy.
	// +optional
	// +kubebuilder:validation:Type=object
	// +kubebuilder:pruning:PreserveUnknownFields
	ConfigSchema *runtime.RawExtension `json:"configSchema,omitempty"`

	// BrowserPath is the path prefix a view plugin serves to browsers: the
	// core, the only entry point for browsers, proxies it to the plugin
	// (D-81). "/" means every path the core doesn't serve itself.
	// +optional
	// +kubebuilder:validation:MaxLength=256
	// +kubebuilder:validation:Pattern=`^/[-a-z0-9_./]*$`
	BrowserPath string `json:"browserPath,omitempty"`

	// EgressFromConfig names top-level config fields whose https URL's host
	// the plugin reaches through the egress proxy, instead of a fixed host
	// (D-75).
	// +optional
	// +kubebuilder:validation:MaxItems=8
	// +kubebuilder:validation:items:MaxLength=64
	// +listType=set
	EgressFromConfig []string `json:"egressFromConfig,omitempty"`

	// ClusterService, when set, asks for one in-cluster service that the
	// owner picks at install (D-74).
	// +optional
	ClusterService *ClusterServiceRequest `json:"clusterService,omitempty"`
}

// SecretMount asks for the instance's Secret of one role to be mounted
// read-only at MountPath.
type SecretMount struct {
	// Role names the Secret's purpose, such as flex-token; the instance's
	// Secret is <instance>-<role>.
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=63
	// +kubebuilder:validation:Pattern=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`
	Role string `json:"role"`
	// MountPath is an absolute path outside /var/run/secrets.
	// +kubebuilder:validation:MinLength=2
	// +kubebuilder:validation:MaxLength=1024
	MountPath string `json:"mountPath"`
}

// Permissions is a set of plugin permissions: what a PluginGrant approves,
// what a plugin runs with, or what still waits for approval.
type Permissions struct {
	// Capabilities are what the plugin may act as.
	// +optional
	// +kubebuilder:validation:items:Enum=connector;view
	// +listType=set
	Capabilities []string `json:"capabilities,omitempty"`
	// Egress lists hosts as "host" (port 443) or "host:port".
	// +optional
	// +kubebuilder:validation:MaxItems=64
	// +kubebuilder:validation:items:MaxLength=259
	// +listType=set
	Egress []string `json:"egress,omitempty"`
	// Secrets lists Secret names: <instance>-<role> for each role the
	// plugin's manifest names (D-89).
	// +optional
	// +kubebuilder:validation:MaxItems=32
	// +kubebuilder:validation:items:MaxLength=253
	// +listType=set
	Secrets []string `json:"secrets,omitempty"`
	// DataTiers lists data tier names.
	// +optional
	// +kubebuilder:validation:MaxItems=32
	// +kubebuilder:validation:items:MaxLength=32
	// +listType=set
	DataTiers []string `json:"dataTiers,omitempty"`
	// BrowserPaths lists path prefixes a view plugin serves to browsers
	// through the core (D-81).
	// +optional
	// +kubebuilder:validation:MaxItems=8
	// +kubebuilder:validation:items:MaxLength=256
	// +listType=set
	BrowserPaths []string `json:"browserPaths,omitempty"`
	// ClusterServices lists the in-cluster service a plugin reaches
	// directly, outside the egress proxy (D-74): at most one.
	// +optional
	// +kubebuilder:validation:MaxItems=1
	// +listType=atomic
	ClusterServices []ClusterServiceRef `json:"clusterServices,omitempty"`
}

// PluginStatus is what the operator observed and did.
type PluginStatus struct {
	// ObservedGeneration is the spec generation the status describes.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions are Ready, ApprovalRequired and SignatureVerified.
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Deployed is the plugin version that runs now; absent when nothing runs.
	// It is the version kept while an upgrade waits for approval.
	// +optional
	Deployed *PluginSpec `json:"deployed,omitempty"`

	// Effective is what the running version is allowed: only what it
	// requests and the PluginGrant approves (D-16). Its egress hosts are the
	// plugin's egress allow-list. Absent when nothing runs.
	// +optional
	Effective *Permissions `json:"effective,omitempty"`

	// PendingApproval is what spec requests and the PluginGrant does not
	// approve. The owner approves it by adding it to the PluginGrant.
	// +optional
	PendingApproval *Permissions `json:"pendingApproval,omitempty"`

	// Signature says who signed spec.image and how far that is trusted
	// (D-45, D-56), while the SignatureVerified condition is True; absent
	// otherwise. The condition's reason tells verified from only signed;
	// this also names the signer and its kind, for the UI to show.
	// +optional
	Signature *SignatureStatus `json:"signature,omitempty"`
}

// SignatureStatus is the trusted signature on a plugin's image.
type SignatureStatus struct {
	// Trust is Verified when the catalog pins the image to the publisher
	// that signed it (D-44), Unverified when another trusted signer did.
	// +kubebuilder:validation:Enum=Verified;Unverified
	Trust string `json:"trust"`
	// SignerKind is the kind of trusted signer (D-45).
	// +kubebuilder:validation:Enum=OpenFiefCI;Publisher;Owner
	SignerKind string `json:"signerKind"`
	// Signer is the trusted signer's name: openfief-ci, a catalog
	// publisher's or one the owner added.
	// +kubebuilder:validation:MaxLength=253
	Signer string `json:"signer"`
}

// PluginList is a list of Plugins.
//
// +kubebuilder:object:root=true
type PluginList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Plugin `json:"items"`
}

// PluginGrant is what the owner approved for the Plugin of the same name
// (D-16). The operator enforces only what both hold.
//
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,categories=openfief
// +kubebuilder:validation:XValidation:rule="self.metadata.name.size() <= 56 && self.metadata.name.matches('^[a-z]([-a-z0-9]*[a-z0-9])?$')",message="name must be the plugin's name"
// +kubebuilder:printcolumn:name="Valid",type=string,JSONPath=`.status.conditions[?(@.type=="Valid")].status`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type PluginGrant struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec is the approved permissions.
	Spec   Permissions       `json:"spec"`
	Status PluginGrantStatus `json:"status,omitempty"`
}

// PluginGrantStatus says whether the grant is well-formed.
type PluginGrantStatus struct {
	// ObservedGeneration is the spec generation the status describes.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions holds Valid.
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// PluginGrantList is a list of PluginGrants.
//
// +kubebuilder:object:root=true
type PluginGrantList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []PluginGrant `json:"items"`
}
