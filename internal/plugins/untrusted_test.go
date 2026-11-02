package plugins_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/plugins"
)

// Synthetic fingerprint of a key nobody trusts.
const strangerKeySHA256 = "0a1b2c3d4e5f60718293a4b5c6d7e8f90a1b2c3d4e5f60718293a4b5c6d7e8f9"

// TestDecideNamesUntrustedSigners checks D-169: an image whose only valid
// signatures are by signers the policy doesn't hold is refused with an
// UntrustedSignerError that names each signer once, keyless by subject and
// issuer and a key by its fingerprint, apart from an image with no
// signature at all; and an untrusted signature never makes an image
// deployable or outweighs a trusted one.
func TestDecideNamesUntrustedSigners(t *testing.T) {
	t.Parallel()
	f := newPolicyFixture(t)
	strangerCI := plugins.UntrustedSigner{
		Issuer:  plugins.GitHubActionsIssuer,
		Subject: "https://github.com/stranger/plugin/.github/workflows/release.yml@refs/heads/main",
	}
	strangerKey := plugins.UntrustedSigner{KeySHA256: strangerKeySHA256}
	tests := []struct {
		name        string
		trusted     []plugins.SignerIdentity
		untrusted   []plugins.UntrustedSigner
		want        plugins.TrustDecision
		wantErr     error
		wantSigners []string
	}{
		{name: "no signature at all", wantErr: plugins.ErrUnsigned},
		{
			name: "keyless signer nobody trusts", untrusted: []plugins.UntrustedSigner{strangerCI},
			wantErr:     plugins.ErrUntrustedSigner,
			wantSigners: []string{strangerCI.Subject + " (issuer " + plugins.GitHubActionsIssuer + ")"},
		},
		{
			name: "key nobody trusts, signed twice", untrusted: []plugins.UntrustedSigner{strangerKey, strangerKey},
			wantErr:     plugins.ErrUntrustedSigner,
			wantSigners: []string{"public key sha256:" + strangerKeySHA256},
		},
		{
			name: "both", untrusted: []plugins.UntrustedSigner{strangerKey, strangerCI},
			wantErr: plugins.ErrUntrustedSigner,
			wantSigners: []string{
				"public key sha256:" + strangerKeySHA256,
				strangerCI.Subject + " (issuer " + plugins.GitHubActionsIssuer + ")",
			},
		},
		{
			name: "a trusted signature still decides", trusted: []plugins.SignerIdentity{f.ownerKey.Identity},
			untrusted: []plugins.UntrustedSigner{strangerCI},
			want:      plugins.TrustDecision{Trust: plugins.Unverified, Signer: f.ownerKey},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			img, err := plugins.ParseImageRef(imageB)
			if err != nil {
				t.Fatalf("image: %v", err)
			}
			signed := plugins.SignedImage{Image: img, Signatures: tt.trusted, Untrusted: tt.untrusted, Manifest: samplePlugin().Manifest}
			got, err := f.policy.Decide(&signed, nil)
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("Decide error (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("Decide (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.wantSigners, namedSigners(t, err, img)); diff != "" {
				t.Errorf("named signers (-want +got):\n%s", diff)
			}
		})
	}
}

// namedSigners returns the signers err names when it is an
// UntrustedSignerError, which must name img; nil otherwise.
func namedSigners(t *testing.T, err error, img plugins.ImageRef) []string {
	t.Helper()
	u, ok := errors.AsType[*plugins.UntrustedSignerError](err)
	if !ok {
		return nil
	}
	if u.Image != img {
		t.Errorf("error names image %s, want %s", u.Image, img)
	}
	return u.Signers
}

// TestUntrustedSignerErrorMessage checks that the refusal's message, which
// the operator writes into the Plugin's status, names the image and every
// signer (D-169).
func TestUntrustedSignerErrorMessage(t *testing.T) {
	t.Parallel()
	img, err := plugins.ParseImageRef(imageB)
	if err != nil {
		t.Fatalf("image: %v", err)
	}
	e := &plugins.UntrustedSignerError{Image: img, Signers: []string{"a (issuer https://issuer.example.com)", "public key sha256:" + strangerKeySHA256}}
	want := plugins.ErrUntrustedSigner.Error() + ": " + imageB + " is signed by a (issuer https://issuer.example.com), public key sha256:" + strangerKeySHA256
	if got := e.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}
