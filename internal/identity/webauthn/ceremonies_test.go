package webauthn_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/arkadiuss/openfief/internal/identity"
	"github.com/arkadiuss/openfief/internal/identity/app"
	"github.com/arkadiuss/openfief/internal/identity/webauthn"
	"github.com/arkadiuss/openfief/internal/identity/webauthn/virtual"
)

// testOrigin parses a synthetic origin.
func testOrigin(t *testing.T, text string) identity.Origin {
	t.Helper()
	o, err := identity.ParseOrigin(text)
	if err != nil {
		t.Fatalf("ParseOrigin(%q): %v", text, err)
	}
	return o
}

// testOwner is an owner with a fixed handle.
func testOwner(t *testing.T) identity.Owner {
	t.Helper()
	o, err := identity.NewOwner(bytes.Repeat([]byte{7}, identity.OwnerHandleBytes), time.Date(2026, 10, 28, 9, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("NewOwner: %v", err)
	}
	return o
}

// register creates a passkey for owner on origin with a new virtual
// authenticator.
func register(t *testing.T, c *webauthn.Ceremonies, owner *identity.Owner, origin identity.Origin) (*virtual.Authenticator, identity.Passkey) {
	t.Helper()
	dev, err := virtual.New(nil)
	if err != nil {
		t.Fatalf("virtual.New: %v", err)
	}
	begun, err := c.BeginRegistration(origin, owner, nil)
	if err != nil {
		t.Fatalf("BeginRegistration: %v", err)
	}
	cred, err := dev.Create(begun.Options, origin.String())
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	p, err := c.FinishRegistration(origin, owner, begun.State, cred)
	if err != nil {
		t.Fatalf("FinishRegistration: %v", err)
	}
	return dev, p
}

// login begins a login on origin, lets dev answer it as a browser on
// browserOrigin would, and finishes it for owner with keys.
func login(t *testing.T, c *webauthn.Ceremonies, dev *virtual.Authenticator, origin identity.Origin, browserOrigin string, owner *identity.Owner, keys []identity.Passkey) (app.Assertion, error) {
	t.Helper()
	begun, err := c.BeginLogin(origin)
	if err != nil {
		t.Fatalf("BeginLogin: %v", err)
	}
	assertion, err := dev.Get(begun.Options, browserOrigin)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	got, err := c.FinishLogin(origin, owner, keys, begun.State, assertion)
	if err != nil {
		return app.Assertion{}, fmt.Errorf("FinishLogin: %w", err)
	}
	return got, nil
}

func TestRegistrationOptions(t *testing.T) {
	t.Parallel()
	c := webauthn.New("OpenFief")
	owner := testOwner(t)
	begun, err := c.BeginRegistration(testOrigin(t, "http://localhost:8080"), &owner, nil)
	if err != nil {
		t.Fatalf("BeginRegistration: %v", err)
	}
	var got struct {
		RP struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"rp"`
		Attestation            string `json:"attestation"`
		AuthenticatorSelection struct {
			ResidentKey      string `json:"residentKey"`
			UserVerification string `json:"userVerification"`
		} `json:"authenticatorSelection"`
	}
	if err := json.Unmarshal(begun.Options, &got); err != nil {
		t.Fatalf("options are not JSON: %v", err)
	}
	if got.RP.ID != "localhost" || got.RP.Name != "OpenFief" || got.Attestation != "none" ||
		got.AuthenticatorSelection.ResidentKey != "required" || got.AuthenticatorSelection.UserVerification != "required" {
		t.Errorf("options = %+v", got)
	}
}

func TestRegisterAndLogin(t *testing.T) {
	t.Parallel()
	c := webauthn.New("OpenFief")
	owner := testOwner(t)
	origin := testOrigin(t, "https://dev.example.com")
	dev, p := register(t, c, &owner, origin)
	if !bytes.Equal(p.CredentialID, dev.CredentialID()) || p.RPID != "dev.example.com" || !p.UserVerified || len(p.PublicKey) == 0 {
		t.Fatalf("passkey = %+v", p)
	}
	got, err := login(t, c, dev, origin, origin.String(), &owner, []identity.Passkey{p})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got.CredentialID, p.CredentialID) || got.SignCount != 1 {
		t.Errorf("FinishLogin = %+v, want the credential with sign count 1", got)
	}
}

func TestLoginRefused(t *testing.T) {
	t.Parallel()
	c := webauthn.New("OpenFief")
	owner := testOwner(t)
	stranger := identity.Owner{Handle: bytes.Repeat([]byte{9}, identity.OwnerHandleBytes)}
	origin := testOrigin(t, "https://dev.example.com")
	tests := []struct {
		name string
		// browserOrigin is where the browser says the assertion was made.
		browserOrigin string
		// rpID overrides the relying party ID the authenticator signs for.
		rpID   string
		owner  *identity.Owner
		noKeys bool
	}{
		{name: "another origin in the client data", browserOrigin: "https://other.example.com", owner: &owner},
		{name: "a passkey of another host", browserOrigin: origin.String(), rpID: "other.example.com", owner: &owner},
		{name: "not the owner's user handle", browserOrigin: origin.String(), owner: &stranger},
		{name: "unknown passkey", browserOrigin: origin.String(), owner: &owner, noKeys: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dev, p := register(t, c, &owner, origin)
			keys := []identity.Passkey{p}
			if tt.noKeys {
				keys = nil
			}
			dev.RPID = tt.rpID
			if _, err := login(t, c, dev, origin, tt.browserOrigin, tt.owner, keys); !errors.Is(err, app.ErrCredentialRefused) {
				t.Errorf("login = %v, want %v", err, app.ErrCredentialRefused)
			}
		})
	}
}

func TestRegistrationRefused(t *testing.T) {
	t.Parallel()
	c := webauthn.New("OpenFief")
	owner := testOwner(t)
	origin := testOrigin(t, "http://localhost:8080")
	dev, err := virtual.New(nil)
	if err != nil {
		t.Fatalf("virtual.New: %v", err)
	}
	begun, err := c.BeginRegistration(origin, &owner, nil)
	if err != nil {
		t.Fatalf("BeginRegistration: %v", err)
	}
	cred, err := dev.Create(begun.Options, "http://localhost:9999")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, err := c.FinishRegistration(origin, &owner, begun.State, cred); !errors.Is(err, app.ErrCredentialRefused) {
		t.Errorf("FinishRegistration from another origin = %v, want %v", err, app.ErrCredentialRefused)
	}
	if _, err := c.FinishRegistration(origin, &owner, begun.State, []byte(`{"id":"x"}`)); !errors.Is(err, app.ErrCredentialRefused) {
		t.Errorf("FinishRegistration of junk = %v, want %v", err, app.ErrCredentialRefused)
	}
}
