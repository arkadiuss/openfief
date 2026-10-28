// Package v1alpha1 is the openfief.io/v1alpha1 API of the plugin operator:
// the Plugin and PluginGrant custom resources (D-13, D-16). Both are
// namespaced and live in the openfief namespace only (D-15).
//
// These types are the Kubernetes shape of the plugins context. The operator
// translates them into the domain's plugins.Plugin and plugins.Grant before
// any decision is made, so nothing outside the operator depends on them.
//
// +kubebuilder:object:generate=true
// +groupName=openfief.io
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// The API group and version of Plugin and PluginGrant.
const (
	Group   = "openfief.io"
	Version = "v1alpha1"
)

// GroupVersion returns openfief.io/v1alpha1.
func GroupVersion() schema.GroupVersion { return schema.GroupVersion{Group: Group, Version: Version} }

// AddToScheme registers Plugin, PluginGrant and their lists with s.
func AddToScheme(s *runtime.Scheme) error {
	s.AddKnownTypes(GroupVersion(), &Plugin{}, &PluginList{}, &PluginGrant{}, &PluginGrantList{})
	metav1.AddToGroupVersion(s, GroupVersion())
	return nil
}
