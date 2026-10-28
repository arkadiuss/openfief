package plugins_test

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/plugins"
)

func TestParseImageRef(t *testing.T) {
	t.Parallel()
	type want struct{ Repository, Digest, String string }
	tests := []struct {
		name    string
		in      string
		want    want
		wantErr error
	}{
		{
			name: "registry, path and digest", in: imageA,
			want: want{"registry.example.com/openfief/sample-connector", digestA, imageA},
		},
		{
			name: "registry with port", in: "registry.example.com:5000/sample@" + digestB,
			want: want{"registry.example.com:5000/sample", digestB, "registry.example.com:5000/sample@" + digestB},
		},
		{
			name: "no registry", in: "openfief/sample@" + digestA,
			want: want{"openfief/sample", digestA, "openfief/sample@" + digestA},
		},
		{
			name: "separators in path", in: "registry.example.com/open_fief/sample__x.y-z@" + digestA,
			want: want{"registry.example.com/open_fief/sample__x.y-z", digestA, "registry.example.com/open_fief/sample__x.y-z@" + digestA},
		},
		{name: "tag only", in: "registry.example.com/sample:1.0.0", wantErr: plugins.ErrImageNotPinned},
		{name: "no tag no digest", in: "registry.example.com/sample", wantErr: plugins.ErrImageNotPinned},
		{name: "tag and digest", in: "registry.example.com/sample:1.0.0@" + digestA, wantErr: plugins.ErrInvalidImageRef},
		{name: "digest too short", in: "registry.example.com/sample@sha256:abc", wantErr: plugins.ErrInvalidImageRef},
		{name: "upper-case digest", in: "registry.example.com/sample@" + strings.ToUpper(digestA), wantErr: plugins.ErrInvalidImageRef},
		{name: "non-hex digest", in: "registry.example.com/sample@sha256:" + strings.Repeat("g", 64), wantErr: plugins.ErrInvalidImageRef},
		{name: "other algorithm", in: "registry.example.com/sample@sha512:" + strings.Repeat("a", 128), wantErr: plugins.ErrInvalidImageRef},
		{name: "upper-case path", in: "registry.example.com/Sample@" + digestA, wantErr: plugins.ErrInvalidImageRef},
		{name: "empty repository", in: "@" + digestA, wantErr: plugins.ErrInvalidImageRef},
		{name: "empty path component", in: "registry.example.com//sample@" + digestA, wantErr: plugins.ErrInvalidImageRef},
		{name: "bad separator", in: "registry.example.com/sample.-x@" + digestA, wantErr: plugins.ErrInvalidImageRef},
		{name: "bad registry port", in: "registry.example.com:http/sample@" + digestA, wantErr: plugins.ErrInvalidImageRef},
		{name: "two digests", in: imageA + "@" + digestB, wantErr: plugins.ErrInvalidImageRef},
		{name: "empty", in: "", wantErr: plugins.ErrImageNotPinned},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := plugins.ParseImageRef(tt.in)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("ParseImageRef(%q) error (-want +got):\n%s", tt.in, diff)
			}
			if diff := cmp.Diff(tt.want, want{got.Repository(), got.Digest(), got.String()}); diff != "" {
				t.Errorf("ParseImageRef(%q) (-want +got):\n%s", tt.in, diff)
			}
			if got.IsZero() != (tt.wantErr != nil) {
				t.Errorf("ParseImageRef(%q).IsZero() = %v", tt.in, got.IsZero())
			}
		})
	}
}

func TestImageRefEqual(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		a, b string
		want bool
	}{
		{name: "same", a: imageA, b: imageA, want: true},
		{name: "different digest", a: imageA, b: imageB, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			a, errA := plugins.ParseImageRef(tt.a)
			b, errB := plugins.ParseImageRef(tt.b)
			if errA != nil || errB != nil {
				t.Fatalf("ParseImageRef: %v, %v", errA, errB)
			}
			if got := a.Equal(b); got != tt.want {
				t.Errorf("Equal() = %v, want %v", got, tt.want)
			}
		})
	}
}
