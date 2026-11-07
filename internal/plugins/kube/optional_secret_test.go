package kube_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/arkadiuss/openfief/internal/plugins"
	"github.com/arkadiuss/openfief/internal/plugins/kube"
)

// TestInstancesKeepOptionalSecret checks that an optional Secret, such as a
// connector's extra CA bundle, stays optional through the Plugin resource
// the CLI writes and reads back for an upgrade.
func TestInstancesKeepOptionalSecret(t *testing.T) {
	t.Parallel()
	img, err := plugins.ParseImageRef(sampleImage)
	if err != nil {
		t.Fatalf("image: %v", err)
	}
	mounts := []plugins.SecretMount{
		{Role: "ca-bundle", MountPath: "/etc/openfief/ca-bundle", Optional: true},
		{Role: "credentials", MountPath: "/etc/openfief/sample"},
	}
	signed := plugins.SignedImage{Image: img, Manifest: plugins.Manifest{
		Capabilities: []plugins.Capability{plugins.CapabilityConnector},
		Secrets:      mounts,
	}}
	p, g, err := plugins.PlanInstall("sample", &signed, nil, nil)
	if err != nil {
		t.Fatalf("plan: %v", err)
	}
	plugin, grant := kube.Resources(&p, &g)
	got, _, err := fakeCluster(t, plugin, grant).Get(context.Background(), "sample")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if diff := cmp.Diff(mounts, got.Manifest.Secrets); diff != "" {
		t.Errorf("Secrets read back (-want +got):\n%s", diff)
	}
}

// TestInstancesKeepExtraCA checks that whether an instance expects an extra
// CA bundle survives the Plugin resource the CLI writes, as spec.extraCa,
// and reads back for an upgrade.
func TestInstancesKeepExtraCA(t *testing.T) {
	t.Parallel()
	img, err := plugins.ParseImageRef(sampleImage)
	if err != nil {
		t.Fatalf("image: %v", err)
	}
	signed := plugins.SignedImage{Image: img, Manifest: plugins.Manifest{
		Capabilities: []plugins.Capability{plugins.CapabilityConnector},
		Secrets:      []plugins.SecretMount{{Role: plugins.CABundleRole, MountPath: "/etc/openfief/ca-bundle", Optional: true}},
	}}
	for _, on := range []bool{true, false} {
		p, g, err := plugins.PlanInstall("sample", &signed, nil, nil)
		if err != nil {
			t.Fatalf("plan: %v", err)
		}
		if setErr := p.SetExtraCA(on); setErr != nil {
			t.Fatalf("SetExtraCA(%v): %v", on, setErr)
		}
		plugin, grant := kube.Resources(&p, &g)
		if plugin.Spec.ExtraCA != on {
			t.Errorf("spec.extraCa = %v, want %v", plugin.Spec.ExtraCA, on)
		}
		got, _, err := fakeCluster(t, plugin, grant).Get(context.Background(), "sample")
		if err != nil {
			t.Fatalf("Get: %v", err)
		}
		if got.ExtraCA != on {
			t.Errorf("ExtraCA read back = %v, want %v", got.ExtraCA, on)
		}
	}
}
