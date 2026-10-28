package app_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/arkadiuss/openfief/internal/identity"
	"github.com/arkadiuss/openfief/internal/identity/app"
)

// memSession is one session in memAuthStore.
type memSession struct {
	session identity.Session
	hash    []byte
}

// memCode is one enrollment code in memAuthStore.
type memCode struct {
	code identity.EnrollmentCode
	hash []byte
}

// memAuthStore is an in-memory app.AuthStore.
type memAuthStore struct {
	mu         sync.Mutex
	owner      *identity.Owner
	passkeys   []identity.Passkey
	ceremonies map[string]identity.Ceremony
	codes      map[identity.EnrollmentCodeID]memCode
	sessions   map[identity.SessionID]memSession
}

func newMemAuthStore() *memAuthStore {
	return &memAuthStore{
		ceremonies: map[string]identity.Ceremony{},
		codes:      map[identity.EnrollmentCodeID]memCode{},
		sessions:   map[identity.SessionID]memSession{},
	}
}

func (m *memAuthStore) Owner(context.Context) (identity.Owner, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.owner == nil {
		return identity.Owner{}, app.ErrNoOwner
	}
	return *m.owner, nil
}

func (m *memAuthStore) EnsureOwner(_ context.Context, candidate identity.Owner) (identity.Owner, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.owner == nil {
		m.owner = &candidate
	}
	return *m.owner, nil
}

func (m *memAuthStore) SavePasskey(_ context.Context, p *identity.Passkey) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := range m.passkeys {
		if m.passkeys[i].SameCredential(p) {
			return fmt.Errorf("%w: %s", errDuplicate, p.ID())
		}
	}
	m.passkeys = append(m.passkeys, *p)
	return nil
}

func (m *memAuthStore) Passkeys(context.Context) ([]identity.Passkey, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return slices.Clone(m.passkeys), nil
}

func (m *memAuthStore) UpdatePasskeyUse(_ context.Context, p *identity.Passkey) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := range m.passkeys {
		if m.passkeys[i].SameCredential(p) {
			m.passkeys[i].SignCount, m.passkeys[i].BackupState, m.passkeys[i].LastUsedAt = p.SignCount, p.BackupState, p.LastUsedAt
			return nil
		}
	}
	return app.ErrUnknownPasskey
}

func (m *memAuthStore) DeletePasskey(_ context.Context, id []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	i := slices.IndexFunc(m.passkeys, func(p identity.Passkey) bool { return bytes.Equal(p.CredentialID, id) })
	switch {
	case i < 0:
		return app.ErrUnknownPasskey
	case len(m.passkeys) == 1:
		return app.ErrLastPasskey
	}
	m.passkeys = slices.Delete(m.passkeys, i, i+1)
	return nil
}

func (m *memAuthStore) SaveCeremony(_ context.Context, c *identity.Ceremony) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ceremonies[c.ID] = *c
	return nil
}

func (m *memAuthStore) TakeCeremony(_ context.Context, id string) (identity.Ceremony, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	c, ok := m.ceremonies[id]
	if !ok {
		return identity.Ceremony{}, app.ErrUnknownCeremony
	}
	delete(m.ceremonies, id)
	return c, nil
}

func (m *memAuthStore) SaveEnrollmentCode(_ context.Context, c *identity.EnrollmentCode, hash []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.codes[c.ID] = memCode{code: *c, hash: slices.Clone(hash)}
	return nil
}

func (m *memAuthStore) EnrollmentCode(_ context.Context, id identity.EnrollmentCodeID) (identity.EnrollmentCode, []byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	c, ok := m.codes[id]
	if !ok {
		return identity.EnrollmentCode{}, nil, app.ErrUnknownEnrollmentCode
	}
	return c.code, c.hash, nil
}

func (m *memAuthStore) UseEnrollmentCode(_ context.Context, id identity.EnrollmentCodeID, at time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	c, ok := m.codes[id]
	if !ok || !c.code.UsedAt.IsZero() {
		return identity.ErrEnrollmentCodeUsed
	}
	c.code.UsedAt = at
	m.codes[id] = c
	return nil
}

func (m *memAuthStore) SaveSession(_ context.Context, s *identity.Session, hash []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[s.ID] = memSession{session: *s, hash: slices.Clone(hash)}
	return nil
}

func (m *memAuthStore) Session(_ context.Context, id identity.SessionID) (identity.Session, []byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.sessions[id]
	if !ok {
		return identity.Session{}, nil, app.ErrUnknownSession
	}
	return s.session, s.hash, nil
}

func (m *memAuthStore) TouchSession(_ context.Context, id identity.SessionID, at time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	s := m.sessions[id]
	s.session.LastSeenAt = at
	m.sessions[id] = s
	return nil
}

func (m *memAuthStore) RevokeSession(_ context.Context, id identity.SessionID, at time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.sessions[id]
	if !ok {
		return app.ErrUnknownSession
	}
	if s.session.RevokedAt.IsZero() {
		s.session.RevokedAt = at
	}
	m.sessions[id] = s
	return nil
}

// errDuplicate reports a passkey stored twice.
var errDuplicate = errors.New("duplicate passkey")

// demoMarker is a fixed app.DemoMarker.
type demoMarker bool

func (d demoMarker) DemoMarked(context.Context) (bool, error) { return bool(d), nil }
