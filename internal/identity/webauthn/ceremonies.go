// Package webauthn runs the WebAuthn side of passkey ceremonies for the
// identity context (docs/DESIGN.md D-26) on github.com/go-webauthn/webauthn,
// so no WebAuthn verification is written here (D-29). Each call is for one
// origin: the relying party ID is its host and the only accepted origin is
// itself, so a passkey works only on the host it was created on (D-60).
//
// Passkeys are discoverable credentials created with user verification
// required and no attestation; logins name no credentials and require user
// verification too.
package webauthn

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	"github.com/go-webauthn/webauthn/protocol"
	gowebauthn "github.com/go-webauthn/webauthn/webauthn"

	"github.com/arkadiuss/openfief/internal/identity"
	"github.com/arkadiuss/openfief/internal/identity/app"
)

// errNotOwner reports an assertion whose user handle is not the owner's.
var errNotOwner = errors.New("user handle is not the owner's")

const (
	// ownerName is the WebAuthn user name every passkey carries: there is
	// one user, the owner (GR-2).
	ownerName = "owner"
	// ownerDisplayName is what the browser shows for the user.
	ownerDisplayName = "OpenFief owner"
)

// Ceremonies is app.Ceremonies over go-webauthn.
type Ceremonies struct {
	rpName string
}

var _ app.Ceremonies = (*Ceremonies)(nil)

// New returns the ceremonies, naming the relying party rpName in the
// browser's prompts, such as "OpenFief".
func New(rpName string) *Ceremonies { return &Ceremonies{rpName: rpName} }

// user is the owner as go-webauthn sees them, with their passkeys.
type user struct {
	owner *identity.Owner
	creds []gowebauthn.Credential
}

// WebAuthnID returns the owner's random user handle.
func (u *user) WebAuthnID() []byte { return u.owner.Handle }

// WebAuthnName returns "owner".
func (*user) WebAuthnName() string { return ownerName }

// WebAuthnDisplayName returns "OpenFief owner".
func (*user) WebAuthnDisplayName() string { return ownerDisplayName }

// WebAuthnCredentials returns the owner's passkeys on this host.
func (u *user) WebAuthnCredentials() []gowebauthn.Credential { return u.creds }

// newUser returns owner with passkeys as go-webauthn credentials.
func newUser(owner *identity.Owner, passkeys []identity.Passkey) *user {
	u := &user{owner: owner, creds: make([]gowebauthn.Credential, 0, len(passkeys))}
	for i := range passkeys {
		u.creds = append(u.creds, credential(&passkeys[i]))
	}
	return u
}

// credential returns p as a go-webauthn credential record.
func credential(p *identity.Passkey) gowebauthn.Credential {
	transports := make([]protocol.AuthenticatorTransport, 0, len(p.Transports))
	for _, t := range p.Transports {
		transports = append(transports, protocol.AuthenticatorTransport(t))
	}
	return gowebauthn.Credential{
		ID:                p.CredentialID,
		PublicKey:         p.PublicKey,
		AttestationType:   p.AttestationType,
		AttestationFormat: p.AttestationFormat,
		Transport:         transports,
		Flags: gowebauthn.CredentialFlags{
			UserPresent: true, UserVerified: p.UserVerified,
			BackupEligible: p.BackupEligible, BackupState: p.BackupState,
		},
		Authenticator: gowebauthn.Authenticator{AAGUID: p.AAGUID, SignCount: p.SignCount},
	}
}

// BeginRegistration implements app.Ceremonies: the creation options as
// PublicKeyCredentialCreationOptionsJSON, excluding existing, and the
// ceremony state as JSON.
func (c *Ceremonies) BeginRegistration(origin identity.Origin, owner *identity.Owner, existing []identity.Passkey) (app.Begun, error) {
	wa, err := c.relyingParty(origin)
	if err != nil {
		return app.Begun{}, err
	}
	u := newUser(owner, existing)
	creation, session, err := wa.BeginRegistration(u, gowebauthn.WithExclusions(gowebauthn.Credentials(u.creds).CredentialDescriptors()))
	if err != nil {
		return app.Begun{}, fmt.Errorf("beginning a registration: %w", err)
	}
	return marshalPair(creation.Response, session)
}

// FinishRegistration implements app.Ceremonies.
func (c *Ceremonies) FinishRegistration(origin identity.Origin, owner *identity.Owner, state, credentialJSON []byte) (identity.Passkey, error) {
	wa, err := c.relyingParty(origin)
	if err != nil {
		return identity.Passkey{}, err
	}
	var session gowebauthn.SessionData
	if jsonErr := json.Unmarshal(state, &session); jsonErr != nil {
		return identity.Passkey{}, fmt.Errorf("reading the ceremony state: %w", jsonErr)
	}
	parsed, err := protocol.ParseCredentialCreationResponseBytes(credentialJSON)
	if err != nil {
		return identity.Passkey{}, refused(err)
	}
	cred, err := wa.CreateCredential(newUser(owner, nil), session, parsed)
	if err != nil {
		return identity.Passkey{}, refused(err)
	}
	transports := make([]string, 0, len(cred.Transport))
	for _, t := range cred.Transport {
		transports = append(transports, string(t))
	}
	return identity.Passkey{
		CredentialID:      slices.Clone(cred.ID),
		RPID:              origin.RPID(),
		PublicKey:         slices.Clone(cred.PublicKey),
		AttestationType:   cred.AttestationType,
		AttestationFormat: cred.AttestationFormat,
		AAGUID:            slices.Clone(cred.Authenticator.AAGUID),
		SignCount:         cred.Authenticator.SignCount,
		Transports:        transports,
		UserVerified:      cred.Flags.UserVerified,
		BackupEligible:    cred.Flags.BackupEligible,
		BackupState:       cred.Flags.BackupState,
	}, nil
}

// BeginLogin implements app.Ceremonies: request options for a
// discoverable-credential login as PublicKeyCredentialRequestOptionsJSON,
// and the ceremony state as JSON.
func (c *Ceremonies) BeginLogin(origin identity.Origin) (app.Begun, error) {
	wa, err := c.relyingParty(origin)
	if err != nil {
		return app.Begun{}, err
	}
	assertion, session, err := wa.BeginDiscoverableLogin(gowebauthn.WithUserVerification(protocol.VerificationRequired))
	if err != nil {
		return app.Begun{}, fmt.Errorf("beginning a login: %w", err)
	}
	return marshalPair(assertion.Response, session)
}

// FinishLogin implements app.Ceremonies. It returns the assertion's own
// signature counter and backup state, for the caller to check the counter
// against the stored one.
func (c *Ceremonies) FinishLogin(origin identity.Origin, owner *identity.Owner, passkeys []identity.Passkey, state, credentialJSON []byte) (app.Assertion, error) {
	wa, err := c.relyingParty(origin)
	if err != nil {
		return app.Assertion{}, err
	}
	var session gowebauthn.SessionData
	if jsonErr := json.Unmarshal(state, &session); jsonErr != nil {
		return app.Assertion{}, fmt.Errorf("reading the ceremony state: %w", jsonErr)
	}
	parsed, err := protocol.ParseCredentialRequestResponseBytes(credentialJSON)
	if err != nil {
		return app.Assertion{}, refused(err)
	}
	u := newUser(owner, passkeys)
	handler := func(_, userHandle []byte) (gowebauthn.User, error) {
		if !bytes.Equal(userHandle, owner.Handle) {
			return nil, errNotOwner
		}
		return u, nil
	}
	_, cred, err := wa.ValidatePasskeyLogin(handler, session, parsed)
	if err != nil {
		return app.Assertion{}, refused(err)
	}
	data := parsed.Response.AuthenticatorData
	return app.Assertion{
		CredentialID: slices.Clone(cred.ID),
		SignCount:    data.Counter,
		BackupState:  data.Flags.HasBackupState(),
	}, nil
}

// marshalPair returns options and state as JSON.
func marshalPair(options, state any) (app.Begun, error) {
	o, err := json.Marshal(options)
	if err != nil {
		return app.Begun{}, fmt.Errorf("encoding the options: %w", err)
	}
	s, err := json.Marshal(state)
	if err != nil {
		return app.Begun{}, fmt.Errorf("encoding the ceremony state: %w", err)
	}
	return app.Begun{Options: o, State: s}, nil
}

// refused wraps a verification failure in app.ErrCredentialRefused, with
// go-webauthn's details for the log; they never contain a secret.
func refused(err error) error {
	var perr *protocol.Error
	if errors.As(err, &perr) && perr.DevInfo != "" {
		return fmt.Errorf("%w: %w: %s", app.ErrCredentialRefused, err, perr.DevInfo)
	}
	return fmt.Errorf("%w: %w", app.ErrCredentialRefused, err)
}

// relyingParty returns the go-webauthn relying party for origin alone.
func (c *Ceremonies) relyingParty(origin identity.Origin) (*gowebauthn.WebAuthn, error) {
	wa, err := gowebauthn.New(&gowebauthn.Config{
		RPID:                  origin.RPID(),
		RPDisplayName:         c.rpName,
		RPOrigins:             []string{origin.String()},
		AttestationPreference: protocol.PreferNoAttestation,
		AuthenticatorSelection: protocol.AuthenticatorSelection{
			ResidentKey:        protocol.ResidentKeyRequirementRequired,
			RequireResidentKey: protocol.ResidentKeyRequired(),
			UserVerification:   protocol.VerificationRequired,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("relying party for %s: %w", origin, err)
	}
	return wa, nil
}
