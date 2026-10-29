// Package virtual is a software WebAuthn authenticator for tests: it holds
// one ES256 passkey and answers creation and request options the way a
// browser and a platform authenticator would, with attestation "none", user
// presence and user verification. It signs with crypto/ecdsa and encodes
// with go-webauthn's CBOR settings; it is never used outside tests.
package virtual

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/go-webauthn/webauthn/protocol/webauthncbor"
)

// ErrBadOptions reports options that are not WebAuthn options JSON.
var ErrBadOptions = errors.New("bad WebAuthn options")

// Authenticator flags and COSE constants (WebAuthn §6.1, RFC 9053).
const (
	flagUP      = 0x01
	flagUV      = 0x04
	flagBE      = 0x08
	flagBS      = 0x10
	flagAT      = 0x40
	coseKty     = 1
	coseAlg     = 3
	coseCrv     = -1
	coseX       = -2
	coseY       = -3
	coseEC2     = 2
	coseES256   = -7
	coseP256    = 1
	credIDBytes = 32
	aaguidBytes = 16
	coordBytes  = 32
)

// Authenticator is one passkey on a software authenticator.
type Authenticator struct {
	key        *ecdsa.PrivateKey
	credID     []byte
	userHandle []byte
	// SignCount is the counter the next assertion carries, after it is
	// incremented. Set it to make the counter go backwards.
	SignCount uint32
	// Synced marks the passkey backup eligible and backed up, as synced
	// passkeys are; synced passkeys keep their counter at zero.
	Synced bool
	// RPID, when set, replaces the relying party ID the options name, to
	// test a passkey answering for another host.
	RPID string
}

// New returns an authenticator with a new ES256 key and credential ID from
// random (crypto/rand.Reader when nil).
func New(random io.Reader) (*Authenticator, error) {
	if random == nil {
		random = rand.Reader
	}
	key, err := ecdsa.GenerateKey(elliptic.P256(), random)
	if err != nil {
		return nil, fmt.Errorf("generating a key: %w", err)
	}
	id := make([]byte, credIDBytes)
	if _, err := io.ReadFull(random, id); err != nil {
		return nil, fmt.Errorf("credential ID: %w", err)
	}
	return &Authenticator{key: key, credID: id}, nil
}

// CredentialID returns the passkey's credential ID.
func (a *Authenticator) CredentialID() []byte { return a.credID }

// creationOptions is the part of PublicKeyCredentialCreationOptionsJSON the
// authenticator reads.
type creationOptions struct {
	Challenge string `json:"challenge"`
	RP        struct {
		ID string `json:"id"`
	} `json:"rp"`
	User struct {
		ID string `json:"id"`
	} `json:"user"`
}

// Create answers creation options as a browser would for origin, and
// returns RegistrationResponseJSON.
func (a *Authenticator) Create(optionsJSON []byte, origin string) ([]byte, error) {
	var opts creationOptions
	if err := json.Unmarshal(optionsJSON, &opts); err != nil || opts.Challenge == "" {
		return nil, fmt.Errorf("%w: creation options", ErrBadOptions)
	}
	handle, err := base64.RawURLEncoding.DecodeString(opts.User.ID)
	if err != nil {
		return nil, fmt.Errorf("%w: user ID", ErrBadOptions)
	}
	a.userHandle = handle
	clientData, err := clientDataJSON("webauthn.create", opts.Challenge, origin)
	if err != nil {
		return nil, err
	}
	cose, err := a.coseKey()
	if err != nil {
		return nil, err
	}
	auth := a.authData(opts.RP.ID, flagAT)
	auth = append(auth, make([]byte, aaguidBytes)...)
	auth = binary.BigEndian.AppendUint16(auth, credIDBytes)
	auth = append(auth, a.credID...)
	auth = append(auth, cose...)
	att, err := webauthncbor.Marshal(map[string]any{"fmt": "none", "attStmt": map[string]any{}, "authData": auth})
	if err != nil {
		return nil, fmt.Errorf("encoding the attestation: %w", err)
	}
	return a.response(map[string]any{
		"clientDataJSON":    b64(clientData),
		"attestationObject": b64(att),
		"transports":        []string{"internal"},
	})
}

// requestOptions is the part of PublicKeyCredentialRequestOptionsJSON the
// authenticator reads.
type requestOptions struct {
	Challenge string `json:"challenge"`
	RPID      string `json:"rpId"`
}

// Get answers request options as a browser would for origin, signing with
// the passkey after incrementing SignCount (unless Synced), and returns
// AuthenticationResponseJSON.
func (a *Authenticator) Get(optionsJSON []byte, origin string) ([]byte, error) {
	var opts requestOptions
	if err := json.Unmarshal(optionsJSON, &opts); err != nil || opts.Challenge == "" {
		return nil, fmt.Errorf("%w: request options", ErrBadOptions)
	}
	clientData, err := clientDataJSON("webauthn.get", opts.Challenge, origin)
	if err != nil {
		return nil, err
	}
	if !a.Synced {
		a.SignCount++
	}
	auth := a.authData(opts.RPID, 0)
	clientHash := sha256.Sum256(clientData)
	digest := sha256.Sum256(append(append([]byte{}, auth...), clientHash[:]...))
	sig, err := ecdsa.SignASN1(rand.Reader, a.key, digest[:])
	if err != nil {
		return nil, fmt.Errorf("signing: %w", err)
	}
	return a.response(map[string]any{
		"clientDataJSON":    b64(clientData),
		"authenticatorData": b64(auth),
		"signature":         b64(sig),
		"userHandle":        b64(a.userHandle),
	})
}

// authData returns the authenticator data's fixed part for rpID (or RPID
// when set) with the UP and UV flags, extra, and the backup flags.
func (a *Authenticator) authData(rpID string, extra byte) []byte {
	if a.RPID != "" {
		rpID = a.RPID
	}
	hash := sha256.Sum256([]byte(rpID))
	flags := byte(flagUP|flagUV) | extra
	if a.Synced {
		flags |= flagBE | flagBS
	}
	out := append(hash[:], flags)
	return binary.BigEndian.AppendUint32(out, a.SignCount)
}

// coseKey returns the passkey's public key as a COSE_Key.
func (a *Authenticator) coseKey() ([]byte, error) {
	pub, err := a.key.PublicKey.Bytes() // 0x04 || X || Y
	if err != nil {
		return nil, fmt.Errorf("public key: %w", err)
	}
	out, err := webauthncbor.Marshal(map[int]any{
		coseKty: coseEC2, coseAlg: coseES256, coseCrv: coseP256,
		coseX: pub[1 : 1+coordBytes], coseY: pub[1+coordBytes:],
	})
	if err != nil {
		return nil, fmt.Errorf("encoding the public key: %w", err)
	}
	return out, nil
}

// response wraps an authenticator response as a PublicKeyCredential's JSON.
func (a *Authenticator) response(resp map[string]any) ([]byte, error) {
	out, err := json.Marshal(map[string]any{
		"id": b64(a.credID), "rawId": b64(a.credID), "type": "public-key",
		"authenticatorAttachment": "platform", "response": resp, "clientExtensionResults": map[string]any{},
	})
	if err != nil {
		return nil, fmt.Errorf("encoding the response: %w", err)
	}
	return out, nil
}

// clientDataJSON returns the client data a browser collects.
func clientDataJSON(typ, challenge, origin string) ([]byte, error) {
	out, err := json.Marshal(map[string]any{"type": typ, "challenge": challenge, "origin": origin, "crossOrigin": false})
	if err != nil {
		return nil, fmt.Errorf("encoding client data: %w", err)
	}
	return out, nil
}

// b64 is unpadded base64url.
func b64(b []byte) string { return base64.RawURLEncoding.EncodeToString(b) }
