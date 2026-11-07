package operator_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/operator"
)

// TestBuildObjectsOptionalSecret checks that a Secret the manifest marks
// optional, such as a connector's extra CA bundle, is mounted as an optional
// volume, so the pod starts while the owner hasn't created it, and that a
// required one stays required.
func TestBuildObjectsOptionalSecret(t *testing.T) {
	t.Parallel()
	p := &plugins.Plugin{
		Name:  "sample",
		Image: sampleImage,
		Manifest: plugins.Manifest{
			Capabilities: []plugins.Capability{plugins.CapabilityConnector},
			Secrets: []plugins.SecretMount{
				{Role: "token", MountPath: "/etc/openfief/token"},
				{Role: "ca-bundle", MountPath: "/etc/openfief/ca-bundle", Optional: true},
			},
		},
	}
	g := &plugins.Grant{Plugin: "sample", Permissions: plugins.Permissions{
		Capabilities: []plugins.Capability{plugins.CapabilityConnector},
		Secrets:      []string{"sample--ca-bundle", "sample--token"},
	}}
	e, err := plugins.Effective(p, g)
	if err != nil {
		t.Fatalf("Effective: %v", err)
	}
	c := cluster()
	w, err := plugins.Compile(e, &c.Topology)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	objs, err := operator.BuildObjects(&w, &c)
	if err != nil {
		t.Fatalf("BuildObjects: %v", err)
	}
	var got []corev1.Volume
	for _, v := range objs.Deployment.Spec.Template.Spec.Volumes {
		if v.Secret != nil {
			got = append(got, v)
		}
	}
	want := []corev1.Volume{
		{Name: "secret-0", VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{
			SecretName: plugins.SecretName("sample", "ca-bundle"), Optional: ptrTo(true),
		}}},
		{Name: "secret-1", VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{SecretName: plugins.SecretName("sample", "token")}}},
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Secret volumes (-want +got):\n%s", diff)
	}
}
