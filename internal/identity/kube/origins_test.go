package kube_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/arkadiuss/openfief/internal/identity/kube"
)

func TestAuthOrigins(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		args []string
		want []string
	}{
		{name: "none: the core's default", args: []string{"serve", "--listen=:8080"}, want: []string{kube.DefaultAuthOrigin}},
		{
			name: "every form, in order",
			args: []string{"serve", "--auth-origin=http://localhost:8080", "-auth-origin", "https://fief.example.org", "--auth-origin", "https://demo.example.com"},
			want: []string{"http://localhost:8080", "https://fief.example.org", "https://demo.example.com"},
		},
		{name: "other flags that start alike", args: []string{"--auth-origins=x", "--session-idle=12h"}, want: []string{kube.DefaultAuthOrigin}},
		{name: "a flag with its value missing", args: []string{"--auth-origin"}, want: []string{kube.DefaultAuthOrigin}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if diff := cmp.Diff(tt.want, kube.AuthOrigins(tt.args)); diff != "" {
				t.Errorf("AuthOrigins (-want +got):\n%s", diff)
			}
		})
	}
}

func TestAllowedOrigins(t *testing.T) {
	t.Parallel()
	pod := corePod("core-a", corev1.PodRunning, true, false)
	pod.Spec.Containers = []corev1.Container{
		{Name: "sidecar", Args: []string{"--auth-origin=https://wrong.example.com"}},
		{Name: kube.CoreContainer, Args: []string{"serve", "--auth-origin=http://localhost:8080", "--auth-origin=https://fief.example.org"}},
	}
	a := &kube.CoreAccess{Client: fake.NewClientset(pod), Namespace: "openfief"}
	got, err := a.AllowedOrigins(t.Context())
	if err != nil {
		t.Fatalf("AllowedOrigins: %v", err)
	}
	if diff := cmp.Diff([]string{"http://localhost:8080", "https://fief.example.org"}, got); diff != "" {
		t.Errorf("AllowedOrigins (-want +got):\n%s", diff)
	}
	empty := &kube.CoreAccess{Client: fake.NewClientset(), Namespace: "openfief"}
	if _, err := empty.AllowedOrigins(t.Context()); err == nil {
		t.Error("AllowedOrigins with no core pod succeeded")
	}
}
