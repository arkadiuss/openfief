package operator

import (
	"fmt"
	"slices"
	"strings"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/operator/v1alpha1"
)

// Reasons of the Ready and ApprovalRequired conditions.
const (
	ReasonRunning              = "Running"
	ReasonRunningPrevious      = "RunningPreviousVersion"
	ReasonProgressing          = "Progressing"
	ReasonNoGrant              = "NoGrant"
	ReasonInvalidGrant         = "InvalidGrant"
	ReasonInvalidSpec          = "InvalidSpec"
	ReasonInvalidConfig        = "InvalidConfig"
	ReasonInvalidPrevious      = "InvalidPreviousVersion"
	ReasonSignatureNotVerified = "SignatureNotVerified"
	ReasonAwaitingApproval     = "AwaitingApproval"
	ReasonNothingGranted       = "NothingGranted"
	ReasonClusterUnavailable   = "ClusterUnavailable"
	ReasonPermissionsPending   = "PermissionsPending"
	ReasonApproved             = "Approved"
	ReasonInvalidImage         = "InvalidImage"
	ReasonBrowserPathHeld      = "BrowserPathHeld"
)

// maxMessageLen keeps condition messages short; the full error is logged by
// ID only, never with data.
const maxMessageLen = 1024

// observation is everything one reconcile found and decided for a Plugin.
type observation struct {
	generation int64
	// Problems with the current version, the grant, the plan and the cluster.
	specErr, configErr, grantErr, planErr, clusterErr error
	// browserErr says why a version that would run doesn't: another
	// instance holds a browser path (D-114).
	browserErr error
	// signature is the verification of spec.image.
	signature Verification
	// previousSignature is the verification of the kept previous version's
	// image when it was refused; nil otherwise.
	previousSignature *Verification
	rollout           plugins.Rollout
	// runSpec is the version that runs, nil when nothing does.
	runSpec         *v1alpha1.PluginSpec
	deploymentReady bool
}

// running reports whether a plugin version runs after this reconcile.
func (o *observation) running() bool { return o.runSpec != nil }

// pluginStatus is the status the operator writes after a reconcile. It keeps
// the transition times of conditions that did not change.
func pluginStatus(o *observation, old []metav1.Condition) v1alpha1.PluginStatus {
	st := v1alpha1.PluginStatus{
		ObservedGeneration: o.generation,
		Conditions:         slices.Clone(old),
		PendingApproval:    permissionsToAPI(&o.rollout.Pending),
	}
	if o.running() {
		st.Deployed = o.runSpec.DeepCopy()
		st.Effective = effectiveToAPI(o.rollout.Run)
	}
	set := func(t string, s metav1.ConditionStatus, reason, msg string) {
		meta.SetStatusCondition(&st.Conditions, metav1.Condition{
			Type: t, Status: s, Reason: reason, Message: truncate(msg), ObservedGeneration: o.generation,
		})
	}
	ready, approval := readyCondition(o), approvalCondition(o)
	set(v1alpha1.ConditionReady, ready.status, ready.reason, ready.message)
	set(v1alpha1.ConditionApprovalRequired, approval.status, approval.reason, approval.message)
	set(v1alpha1.ConditionSignatureVerified, o.signature.Status, o.signature.Reason, o.signature.Message)
	if o.signature.Verified() && o.signature.Signature != nil {
		st.Signature = o.signature.Signature.DeepCopy()
	}
	return st
}

// condition is the status, reason and message of one condition.
type condition struct {
	status  metav1.ConditionStatus
	reason  string
	message string
}

// readyCondition explains whether the plugin runs and, if not, the first
// reason it does not.
func readyCondition(o *observation) condition {
	if o.running() {
		switch {
		case !o.deploymentReady:
			return condition{metav1.ConditionFalse, ReasonProgressing, "the plugin's Deployment is not available yet"}
		case o.rollout.Outcome == plugins.RolloutPrevious:
			return condition{metav1.ConditionTrue, ReasonRunningPrevious, "the previously deployed version keeps running: " + o.runSpec.Image}
		default:
			return condition{metav1.ConditionTrue, ReasonRunning, "running " + o.runSpec.Image}
		}
	}
	for _, p := range []struct {
		err    error
		reason string
	}{
		{o.grantErr, ReasonInvalidGrant},
		{o.planErr, ReasonInvalidPrevious},
		{o.browserErr, ReasonBrowserPathHeld},
		{o.clusterErr, ReasonClusterUnavailable},
	} {
		if p.err != nil {
			return condition{metav1.ConditionFalse, p.reason, p.err.Error()}
		}
	}
	if o.previousSignature != nil {
		return condition{metav1.ConditionFalse, ReasonSignatureNotVerified, "previous version: " + o.previousSignature.Message}
	}
	return condition{metav1.ConditionFalse, outcomeReason(o), outcomeMessage(o)}
}

// outcomeReason names why the rollout runs nothing.
func outcomeReason(o *observation) string {
	switch o.rollout.Outcome {
	case plugins.RolloutNoGrant:
		return ReasonNoGrant
	case plugins.RolloutAwaitingApproval:
		return ReasonAwaitingApproval
	case plugins.RolloutNothingGranted:
		return ReasonNothingGranted
	case plugins.RolloutNotDeployable, plugins.RolloutCurrent, plugins.RolloutPrevious:
	}
	switch {
	case o.specErr != nil:
		return ReasonInvalidSpec
	case o.configErr != nil:
		return ReasonInvalidConfig
	default:
		return ReasonSignatureNotVerified
	}
}

func outcomeMessage(o *observation) string {
	switch outcomeReason(o) {
	case ReasonNoGrant:
		return "no PluginGrant of the same name approves this plugin"
	case ReasonAwaitingApproval:
		return "the plugin requests permissions its PluginGrant does not approve; see status.pendingApproval"
	case ReasonNothingGranted:
		return "the PluginGrant approves none of the plugin's capabilities"
	case ReasonInvalidSpec:
		return o.specErr.Error()
	case ReasonInvalidConfig:
		return o.configErr.Error()
	default:
		return o.signature.Message
	}
}

// approvalCondition says whether the current version waits for approval.
func approvalCondition(o *observation) condition {
	switch {
	case !o.rollout.Pending.IsEmpty():
		return condition{metav1.ConditionTrue, ReasonPermissionsPending, "add to the PluginGrant to approve: " + describe(&o.rollout.Pending)}
	case o.grantErr != nil:
		return condition{metav1.ConditionUnknown, ReasonInvalidGrant, o.grantErr.Error()}
	case o.specErr != nil:
		return condition{metav1.ConditionUnknown, ReasonInvalidSpec, o.specErr.Error()}
	default:
		return condition{metav1.ConditionFalse, ReasonApproved, "the PluginGrant approves everything the plugin requests"}
	}
}

// describe lists permissions for a condition message, such as
// "egress: api.example.com:443; secrets: sample-token".
func describe(p *plugins.Permissions) string {
	api := permissionsToAPI(p)
	if api == nil {
		return "nothing"
	}
	var parts []string
	for _, f := range []struct {
		name  string
		items []string
	}{
		{"capabilities", api.Capabilities},
		{"egress", api.Egress},
		{"secrets", api.Secrets},
		{"dataTiers", api.DataTiers},
		{"browserPaths", api.BrowserPaths},
		{"clusterServices", clusterServiceNames(api.ClusterServices)},
	} {
		if len(f.items) > 0 {
			parts = append(parts, f.name+": "+strings.Join(f.items, ", "))
		}
	}
	return strings.Join(parts, "; ")
}

// clusterServiceNames writes in-cluster services as namespace/service:port.
func clusterServiceNames(refs []v1alpha1.ClusterServiceRef) []string {
	out := make([]string, 0, len(refs))
	for _, r := range refs {
		out = append(out, fmt.Sprintf("%s/%s:%d", r.Namespace, r.Service, r.Port))
	}
	return out
}

func truncate(s string) string {
	if len(s) <= maxMessageLen {
		return s
	}
	return s[:maxMessageLen-3] + "..."
}
