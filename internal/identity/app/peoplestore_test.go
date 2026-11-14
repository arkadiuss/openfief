package app_test

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/arkadiuss/openfief/internal/identity"
	"github.com/arkadiuss/openfief/internal/identity/app"
)

var _ app.PersonStore = (*memAuthStore)(nil)

func (m *memAuthStore) Person(_ context.Context, id string) (identity.Person, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if id == identity.OwnerPersonID && m.owner != nil {
		return identity.OwnerPerson(m.owner), nil
	}
	for i := range m.viewers {
		if m.viewers[i].ID == id {
			return m.viewers[i], nil
		}
	}
	return identity.Person{}, fmt.Errorf("%w: %s", app.ErrUnknownPerson, id)
}

func (m *memAuthStore) People(context.Context) ([]identity.Person, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []identity.Person
	if m.owner != nil {
		out = append(out, identity.OwnerPerson(m.owner))
	}
	return append(out, m.viewers...), nil
}

func (m *memAuthStore) SavePerson(_ context.Context, p *identity.Person) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.viewers = append(m.viewers, *p)
	return nil
}

func (m *memAuthStore) RevokePerson(_ context.Context, id string, at time.Time) ([]identity.Passkey, int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	i := slices.IndexFunc(m.viewers, func(p identity.Person) bool { return p.ID == id })
	if i < 0 || !m.viewers[i].Active() {
		return nil, 0, fmt.Errorf("%w: %s", app.ErrUnknownPerson, id)
	}
	m.viewers[i] = m.viewers[i].Revoked(at)
	var keys []identity.Passkey
	for j := range m.passkeys {
		if m.passkeys[j].PersonID == id && m.passkeys[j].DisabledAt.IsZero() {
			m.passkeys[j].DisabledAt = at
			keys = append(keys, m.passkeys[j])
		}
	}
	sessions := 0
	for sid := range m.sessions {
		s := m.sessions[sid]
		if s.session.PersonID == id && s.session.RevokedAt.IsZero() {
			s.session.RevokedAt = at
			m.sessions[sid] = s
			sessions++
		}
	}
	return keys, sessions, nil
}

func (m *memAuthStore) PasskeyCounts(context.Context) (map[string]int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := map[string]int{}
	for i := range m.passkeys {
		if m.passkeys[i].DisabledAt.IsZero() {
			out[m.passkeys[i].Person()]++
		}
	}
	return out, nil
}

func (m *memAuthStore) ActiveSessionCounts(_ context.Context, at time.Time) (map[string]int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := map[string]int{}
	for id := range m.sessions {
		s := m.sessions[id].session
		if s.Check(at, "", identity.SessionScopeRead) == nil {
			out[s.Person()]++
		}
	}
	return out, nil
}
