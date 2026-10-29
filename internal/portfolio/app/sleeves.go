package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/arkadiuss/openfief/internal/portfolio"
)

// Errors of the owner's sleeve entries.
var (
	// ErrExists reports a kind or sleeve whose key or ID, made from the
	// name, is taken.
	ErrExists = errors.New("already exists")
	// ErrNotFound reports a kind or sleeve to change that is not stored.
	ErrNotFound = errors.New("not found")
)

// SleeveStore stores sleeve kinds and sleeves, checking each against the
// portfolio rules (portfolio.CheckKind, portfolio.CheckSleeve) before it
// stores it. *ledger/app.Ledger implements it: the ledger keeps sleeves next
// to the accounts that refer to them.
type SleeveStore interface {
	// Kinds returns every sleeve kind, in key order.
	Kinds(ctx context.Context) ([]portfolio.Kind, error)
	// SaveKind adds k or renames the stored kind with its key.
	SaveKind(ctx context.Context, k portfolio.Kind) error
	// Sleeves returns every sleeve, in ID order.
	Sleeves(ctx context.Context) ([]portfolio.Sleeve, error)
	// SaveSleeve adds s or replaces the stored sleeve with its ID.
	SaveSleeve(ctx context.Context, s portfolio.Sleeve) error
}

// Sleeves are the owner's sleeve entries (D-71, D-73): adding and renaming
// sleeve kinds and sleeves, each for an owner-scoped session only (D-79).
// The owner names them; their keys and IDs are made from the names
// (portfolio.KeyFromName) and never change.
type Sleeves struct {
	Owners Owners
	Store  SleeveStore
}

// AddKind adds a sleeve kind named name with default role role, which
// never changes. A name whose key is taken fails with ErrExists.
func (s *Sleeves) AddKind(ctx context.Context, origin, sessionToken, name string, role portfolio.Role) (portfolio.Kind, error) {
	if err := s.Owners.CheckOwner(ctx, origin, sessionToken); err != nil {
		return portfolio.Kind{}, fmt.Errorf("adding a sleeve kind: %w", err)
	}
	key, err := portfolio.KeyFromName(name)
	if err != nil {
		return portfolio.Kind{}, fmt.Errorf("adding a sleeve kind: %w: %w", portfolio.ErrInvalidKind, err)
	}
	k := portfolio.Kind{Key: portfolio.KindKey(key), Name: name, DefaultRole: role}
	kinds, err := s.Store.Kinds(ctx)
	if err != nil {
		return portfolio.Kind{}, fmt.Errorf("adding a sleeve kind: %w", err)
	}
	if _, ok := kindOf(kinds, k.Key); ok {
		return portfolio.Kind{}, fmt.Errorf("adding a sleeve kind: kind %s %w", k.Key, ErrExists)
	}
	if err := s.Store.SaveKind(ctx, k); err != nil {
		return portfolio.Kind{}, fmt.Errorf("adding a sleeve kind: %w", err)
	}
	return k, nil
}

// RenameKind renames the stored kind key to name; its key and default role
// stay. An unknown kind fails with ErrNotFound.
func (s *Sleeves) RenameKind(ctx context.Context, origin, sessionToken string, key portfolio.KindKey, name string) (portfolio.Kind, error) {
	if err := s.Owners.CheckOwner(ctx, origin, sessionToken); err != nil {
		return portfolio.Kind{}, fmt.Errorf("renaming a sleeve kind: %w", err)
	}
	kinds, err := s.Store.Kinds(ctx)
	if err != nil {
		return portfolio.Kind{}, fmt.Errorf("renaming a sleeve kind: %w", err)
	}
	k, ok := kindOf(kinds, key)
	if !ok {
		return portfolio.Kind{}, fmt.Errorf("renaming sleeve kind %q: %w", key, ErrNotFound)
	}
	k.Name = name
	if err := s.Store.SaveKind(ctx, k); err != nil {
		return portfolio.Kind{}, fmt.Errorf("renaming sleeve kind %s: %w", key, err)
	}
	return k, nil
}

// AddSleeve adds sl, whose ID is made from its name; any ID it carries is
// ignored. A zero role takes the kind's default role. A name whose ID is
// taken fails with ErrExists. It returns the sleeve as stored.
func (s *Sleeves) AddSleeve(ctx context.Context, origin, sessionToken string, sl portfolio.Sleeve) (portfolio.Sleeve, error) {
	if err := s.Owners.CheckOwner(ctx, origin, sessionToken); err != nil {
		return portfolio.Sleeve{}, fmt.Errorf("adding a sleeve: %w", err)
	}
	id, err := portfolio.KeyFromName(sl.Name)
	if err != nil {
		return portfolio.Sleeve{}, fmt.Errorf("adding a sleeve: %w: %w", portfolio.ErrInvalidSleeve, err)
	}
	sl.ID = portfolio.SleeveID(id)
	sleeves, err := s.Store.Sleeves(ctx)
	if err != nil {
		return portfolio.Sleeve{}, fmt.Errorf("adding a sleeve: %w", err)
	}
	if _, ok := sleeveOf(sleeves, sl.ID); ok {
		return portfolio.Sleeve{}, fmt.Errorf("adding a sleeve: sleeve %s %w", sl.ID, ErrExists)
	}
	return s.save(ctx, sl)
}

// UpdateSleeve replaces the stored sleeve with sl's ID by sl: a new name,
// kind, role or role override. A zero role takes the kind's default role.
// An unknown sleeve fails with ErrNotFound. It returns the sleeve as
// stored.
func (s *Sleeves) UpdateSleeve(ctx context.Context, origin, sessionToken string, sl portfolio.Sleeve) (portfolio.Sleeve, error) {
	if err := s.Owners.CheckOwner(ctx, origin, sessionToken); err != nil {
		return portfolio.Sleeve{}, fmt.Errorf("changing a sleeve: %w", err)
	}
	sleeves, err := s.Store.Sleeves(ctx)
	if err != nil {
		return portfolio.Sleeve{}, fmt.Errorf("changing a sleeve: %w", err)
	}
	if _, ok := sleeveOf(sleeves, sl.ID); !ok {
		return portfolio.Sleeve{}, fmt.Errorf("changing sleeve %q: %w", sl.ID, ErrNotFound)
	}
	return s.save(ctx, sl)
}

// save stores sl and returns it as stored, its role filled in.
func (s *Sleeves) save(ctx context.Context, sl portfolio.Sleeve) (portfolio.Sleeve, error) {
	if err := s.Store.SaveSleeve(ctx, sl); err != nil {
		return portfolio.Sleeve{}, fmt.Errorf("saving sleeve %s: %w", sl.ID, err)
	}
	sleeves, err := s.Store.Sleeves(ctx)
	if err != nil {
		return portfolio.Sleeve{}, fmt.Errorf("saving sleeve %s: %w", sl.ID, err)
	}
	stored, ok := sleeveOf(sleeves, sl.ID)
	if !ok {
		return portfolio.Sleeve{}, fmt.Errorf("saving sleeve %s: %w", sl.ID, ErrNotFound)
	}
	return stored, nil
}

// kindOf returns the kind with key among kinds.
func kindOf(kinds []portfolio.Kind, key portfolio.KindKey) (portfolio.Kind, bool) {
	for _, k := range kinds {
		if k.Key == key {
			return k, true
		}
	}
	return portfolio.Kind{}, false
}

// sleeveOf returns the sleeve with id among sleeves.
func sleeveOf(sleeves []portfolio.Sleeve, id portfolio.SleeveID) (portfolio.Sleeve, bool) {
	for _, sl := range sleeves {
		if sl.ID == id {
			return sl, true
		}
	}
	return portfolio.Sleeve{}, false
}
