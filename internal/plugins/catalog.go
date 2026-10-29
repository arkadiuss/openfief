package plugins

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
)

// ErrInvalidCatalog reports a catalog or owner-signer document that cannot
// be used. Nothing is trusted from a document that fails.
var ErrInvalidCatalog = errors.New("invalid catalog")

// CatalogEntry pins one reviewed plugin version (D-44): the image digest the
// author reviewed and the identity of the publisher who signed it. Each new
// version needs a new entry.
type CatalogEntry struct {
	// Name is the plugin's catalog name, for messages and the CLI.
	Name string
	// Version is the product release that added the entry, such as
	// 2026.11.0, by which `openfief plugin upgrade` names it (D-88, D-90);
	// empty for an entry added by hand.
	Version string
	// Image is the reviewed image; only its digest is pinned, so a mirror of
	// the same image is verified too.
	Image ImageRef
	// Publisher signed the image. Its identity is exact: a keyless subject,
	// never a pattern, or a public key.
	Publisher Signer
}

// Catalog lists the first-party and verified plugins (D-43). It is a file in
// the OpenFief repository, released with the product.
type Catalog struct {
	Entries []CatalogEntry
}

// Validate reports an entry without a name or image, or whose publisher is
// not a valid SignerPublisher with an exact identity.
func (c *Catalog) Validate() error {
	var errs []error
	for i := range c.Entries {
		e := &c.Entries[i]
		field := indexed("plugins", i)
		switch {
		case e.Name == "":
			errs = append(errs, &FieldError{Field: field + ".name", Err: ErrInvalidCatalog})
		case e.Image.IsZero():
			errs = append(errs, &FieldError{Field: field + ".image", Err: ErrInvalidCatalog})
		case e.Publisher.Kind != SignerPublisher:
			errs = append(errs, fieldErr(field+".publisher", ErrInvalidCatalog, string(e.Publisher.Kind)))
		case e.Publisher.Identity.SubjectPattern != "":
			errs = append(errs, fieldErr(field+".publisher.subjectPattern", ErrInvalidCatalog, e.Publisher.Identity.SubjectPattern))
		default:
			if err := e.Publisher.Validate(); err != nil {
				errs = append(errs, &FieldError{Field: field + ".publisher", Err: err})
			}
		}
	}
	return errors.Join(errs...)
}

// signerDoc is a signer as the catalog and the owner's signer list write it.
type signerDoc struct {
	Name           string `json:"name"`
	Issuer         string `json:"issuer,omitempty"`
	Subject        string `json:"subject,omitempty"`
	SubjectPattern string `json:"subjectPattern,omitempty"`
	PublicKey      string `json:"publicKey,omitempty"`
}

func (d *signerDoc) signer(kind SignerKind) Signer {
	return Signer{Name: d.Name, Kind: kind, Identity: SignerIdentity{
		Issuer: d.Issuer, Subject: d.Subject, SubjectPattern: d.SubjectPattern, PublicKey: d.PublicKey,
	}}
}

// ParseCatalog reads the catalog's JSON form:
//
//	{"plugins": [{"name": "...", "version": "2026.11.0", "image": "repo@sha256:...",
//	  "publisher": {"name": "...", "issuer": "...", "subject": "..."}}]}
//
// where a publisher gives either an issuer and exact subject or a PEM
// "publicKey". Unknown fields are refused, so a newer catalog format is never
// half-read.
func ParseCatalog(data []byte) (Catalog, error) {
	var doc struct {
		Plugins []struct {
			Name      string    `json:"name"`
			Version   string    `json:"version,omitempty"`
			Image     string    `json:"image"`
			Publisher signerDoc `json:"publisher"`
		} `json:"plugins"`
	}
	if err := decodeStrict(data, &doc); err != nil {
		return Catalog{}, fmt.Errorf("%w: %w", ErrInvalidCatalog, err)
	}
	var c Catalog
	var errs []error
	for i := range doc.Plugins {
		p := &doc.Plugins[i]
		img, err := ParseImageRef(p.Image)
		if err != nil {
			errs = append(errs, &FieldError{Field: indexed("plugins", i) + ".image", Err: err})
		}
		c.Entries = append(c.Entries, CatalogEntry{Name: p.Name, Version: p.Version, Image: img, Publisher: p.Publisher.signer(SignerPublisher)})
	}
	errs = append(errs, c.Validate())
	if err := errors.Join(errs...); err != nil {
		return Catalog{}, fmt.Errorf("%w: %w", ErrInvalidCatalog, err)
	}
	return c, nil
}

// ParseOwnerSigners reads the owner's trusted signers (D-45) in JSON form:
//
//	{"signers": [{"name": "...", "issuer": "...", "subject": "..."},
//	             {"name": "...", "issuer": "...", "subjectPattern": "..."},
//	             {"name": "...", "publicKey": "-----BEGIN PUBLIC KEY-----..."}]}
//
// Unknown fields are refused.
func ParseOwnerSigners(data []byte) ([]Signer, error) {
	var doc struct {
		Signers []signerDoc `json:"signers"`
	}
	if err := decodeStrict(data, &doc); err != nil {
		return nil, fmt.Errorf("%w: owner signers: %w", ErrInvalidCatalog, err)
	}
	out := make([]Signer, 0, len(doc.Signers))
	var errs []error
	for i := range doc.Signers {
		s := doc.Signers[i].signer(SignerOwner)
		if err := s.Validate(); err != nil {
			errs = append(errs, &FieldError{Field: indexed("signers", i), Err: err})
		}
		out = append(out, s)
	}
	if err := errors.Join(errs...); err != nil {
		return nil, fmt.Errorf("%w: owner signers: %w", ErrInvalidCatalog, err)
	}
	return out, nil
}

// decodeStrict decodes one JSON document into v, refusing unknown fields and
// trailing data.
func decodeStrict(data []byte, v any) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidJSON, err)
	}
	if dec.More() {
		return fmt.Errorf("%w: trailing data after the JSON document", ErrInvalidJSON)
	}
	return nil
}
