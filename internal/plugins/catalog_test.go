package plugins_test

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/arkadiuss/openfief/internal/plugins"
)

func TestParseCatalog(t *testing.T) {
	t.Parallel()
	img, err := plugins.ParseImageRef(imageA)
	if err != nil {
		t.Fatalf("image: %v", err)
	}
	keyPEM := newPublicKeyPEM(t)
	keyJSON, err := json.Marshal(keyPEM)
	if err != nil {
		t.Fatalf("marshal key: %v", err)
	}
	tests := []struct {
		name    string
		in      string
		want    plugins.Catalog
		wantErr error
	}{
		{name: "empty list", in: `{"plugins":[]}`},
		{name: "no plugins field", in: `{}`},
		{
			name: "keyless publisher",
			in:   `{"plugins":[{"name":"sample","image":"` + imageA + `","publisher":{"name":"example","issuer":"` + issuerExample + `","subject":"` + subjectCI + `"}}]}`,
			want: plugins.Catalog{Entries: []plugins.CatalogEntry{{
				Name: "sample", Image: img,
				Publisher: plugins.Signer{Name: "example", Kind: plugins.SignerPublisher, Identity: keyless(issuerExample, subjectCI)},
			}}},
		},
		{
			name: "key publisher",
			in:   `{"plugins":[{"name":"sample","image":"` + imageA + `","publisher":{"name":"example","publicKey":` + string(keyJSON) + `}}]}`,
			want: plugins.Catalog{Entries: []plugins.CatalogEntry{{
				Name: "sample", Image: img,
				Publisher: plugins.Signer{Name: "example", Kind: plugins.SignerPublisher, Identity: plugins.SignerIdentity{PublicKey: keyPEM}},
			}}},
		},
		{
			name:    "publisher pinned by a pattern",
			in:      `{"plugins":[{"name":"sample","image":"` + imageA + `","publisher":{"name":"example","issuer":"` + issuerExample + `","subjectPattern":".*"}}]}`,
			wantErr: plugins.ErrInvalidCatalog,
		},
		{
			name:    "image by tag",
			in:      `{"plugins":[{"name":"sample","image":"registry.example.com/sample:1.0","publisher":{"name":"example","issuer":"` + issuerExample + `","subject":"` + subjectCI + `"}}]}`,
			wantErr: plugins.ErrInvalidCatalog,
		},
		{
			name:    "no name",
			in:      `{"plugins":[{"image":"` + imageA + `","publisher":{"name":"example","issuer":"` + issuerExample + `","subject":"` + subjectCI + `"}}]}`,
			wantErr: plugins.ErrInvalidCatalog,
		},
		{
			name:    "publisher without identity",
			in:      `{"plugins":[{"name":"sample","image":"` + imageA + `","publisher":{"name":"example"}}]}`,
			wantErr: plugins.ErrInvalidCatalog,
		},
		{name: "unknown field", in: `{"plugins":[],"trustEverything":true}`, wantErr: plugins.ErrInvalidCatalog},
		{name: "trailing data", in: `{"plugins":[]} {}`, wantErr: plugins.ErrInvalidCatalog},
		{name: "not JSON", in: `plugins: []`, wantErr: plugins.ErrInvalidCatalog},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := plugins.ParseCatalog([]byte(tt.in))
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("ParseCatalog error (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ParseCatalog (-want +got):\n%s", diff)
			}
		})
	}
}

func TestCatalogValidate(t *testing.T) {
	t.Parallel()
	img, err := plugins.ParseImageRef(imageA)
	if err != nil {
		t.Fatalf("image: %v", err)
	}
	publisher := plugins.Signer{Name: "p", Kind: plugins.SignerPublisher, Identity: keyless(issuerExample, subjectCI)}
	tests := []struct {
		name  string
		entry plugins.CatalogEntry
		want  []fieldError
	}{
		{name: "valid", entry: plugins.CatalogEntry{Name: "sample", Image: img, Publisher: publisher}},
		{
			name: "no image", entry: plugins.CatalogEntry{Name: "sample", Publisher: publisher},
			want: []fieldError{{Field: "plugins[0].image", Err: plugins.ErrInvalidCatalog}},
		},
		{
			name: "invalid publisher identity",
			entry: plugins.CatalogEntry{Name: "sample", Image: img, Publisher: plugins.Signer{
				Name: "p", Kind: plugins.SignerPublisher, Identity: keyless("", subjectCI),
			}},
			want: []fieldError{{Field: "plugins[0].publisher", Err: plugins.ErrInvalidSigner}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := plugins.Catalog{Entries: []plugins.CatalogEntry{tt.entry}}
			if diff := cmp.Diff(tt.want, fieldErrors(c.Validate()), cmpopts.EquateErrors()); diff != "" {
				t.Errorf("Validate (-want +got):\n%s", diff)
			}
		})
	}
}

func TestParseOwnerSigners(t *testing.T) {
	t.Parallel()
	keyPEM := newPublicKeyPEM(t)
	keyJSON, err := json.Marshal(keyPEM)
	if err != nil {
		t.Fatalf("marshal key: %v", err)
	}
	tests := []struct {
		name    string
		in      string
		want    []plugins.Signer
		wantErr error
	}{
		{name: "none", in: `{"signers":[]}`, want: []plugins.Signer{}},
		{
			name: "keyless, pattern and key",
			in: `{"signers":[{"name":"my-ci","issuer":"` + issuerExample + `","subject":"` + subjectCI + `"},` +
				`{"name":"my-tags","issuer":"` + issuerExample + `","subjectPattern":"https://example\\.com/.*"},` +
				`{"name":"laptop","publicKey":` + string(keyJSON) + `}]}`,
			want: []plugins.Signer{
				{Name: "my-ci", Kind: plugins.SignerOwner, Identity: keyless(issuerExample, subjectCI)},
				{Name: "my-tags", Kind: plugins.SignerOwner, Identity: plugins.SignerIdentity{Issuer: issuerExample, SubjectPattern: `https://example\.com/.*`}},
				{Name: "laptop", Kind: plugins.SignerOwner, Identity: plugins.SignerIdentity{PublicKey: keyPEM}},
			},
		},
		{name: "no name", in: `{"signers":[{"issuer":"` + issuerExample + `","subject":"` + subjectCI + `"}]}`, wantErr: plugins.ErrInvalidSigner},
		{name: "no identity", in: `{"signers":[{"name":"x"}]}`, wantErr: plugins.ErrInvalidSigner},
		{name: "unknown field", in: `{"signers":[{"name":"x","allowUnsigned":true}]}`, wantErr: plugins.ErrInvalidCatalog},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := plugins.ParseOwnerSigners([]byte(tt.in))
			if diff := cmp.Diff(tt.wantErr, err, cmpopts.EquateErrors()); diff != "" {
				t.Fatalf("ParseOwnerSigners error (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("ParseOwnerSigners (-want +got):\n%s", diff)
			}
		})
	}
}
