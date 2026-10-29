package kube_test

import (
	"errors"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/arkadiuss/openfief/internal/identity/kube"
)

// corePod is a core pod named name in phase, ready or not, deleted or not.
func corePod(name string, phase corev1.PodPhase, ready, deleting bool) *corev1.Pod {
	status := corev1.ConditionFalse
	if ready {
		status = corev1.ConditionTrue
	}
	p := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: name, Namespace: "openfief",
			Labels: map[string]string{"app.kubernetes.io/name": "openfief-core"},
		},
		Status: corev1.PodStatus{Phase: phase, Conditions: []corev1.PodCondition{{Type: corev1.PodReady, Status: status}}},
	}
	if deleting {
		now := metav1.NewTime(time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC))
		p.DeletionTimestamp = &now
	}
	return p
}

func TestCorePod(t *testing.T) {
	t.Parallel()
	other := corePod("other-ready", corev1.PodRunning, true, false)
	other.Labels = map[string]string{"app.kubernetes.io/name": "openfief-operator"}
	tests := []struct {
		name    string
		pods    []runtime.Object
		want    string
		wantErr error
	}{
		{
			name: "the first ready, running core pod",
			pods: []runtime.Object{
				corePod("core-c", corev1.PodRunning, true, false), corePod("core-b", corev1.PodRunning, true, false),
				corePod("core-a", corev1.PodRunning, false, false), corePod("core-0", corev1.PodRunning, true, true),
				corePod("core-1", corev1.PodPending, true, false), other,
			},
			want: "core-b",
		},
		{
			name:    "none ready",
			pods:    []runtime.Object{corePod("core-a", corev1.PodRunning, false, false), other},
			wantErr: kube.ErrNoCorePod,
		},
		{name: "no core at all", wantErr: kube.ErrNoCorePod},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a := &kube.CoreAccess{Client: fake.NewClientset(tt.pods...), Namespace: "openfief"}
			got, err := a.CorePod(t.Context())
			if got != tt.want || !errors.Is(err, tt.wantErr) || (err == nil) != (tt.wantErr == nil) {
				t.Errorf("CorePod = %q, %v; want %q, %v", got, err, tt.want, tt.wantErr)
			}
		})
	}
}
