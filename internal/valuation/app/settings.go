package app

import (
	"context"
	"fmt"

	"github.com/arkadiuss/openfief/internal/jurisdiction"
	"github.com/arkadiuss/openfief/internal/kernel"
)

// DisplaySettings are the settings any reader of the record sees with it:
// the reporting currency (D-62) and the number format amounts are written
// in (D-113).
type DisplaySettings struct {
	ReportingCurrency kernel.Currency
	NumberFormat      jurisdiction.NumberFormat
	// NumberFormatDefault reports that the owner has not chosen a number
	// format: NumberFormat is the jurisdiction pack's default.
	NumberFormatDefault bool
}

// OwnerSettings are the owner's settings: what everyone sees, the tax
// wrappers an account can have, and the settings only the owner sees,
// such as the marginal income tax rate (D-78).
type OwnerSettings struct {
	DisplaySettings

	Wrappers []jurisdiction.Wrapper
	Values   jurisdiction.Settings
}

// Settings are the owner's settings use cases (D-79): reading the display
// settings for anyone, and reading and changing every setting for the
// owner, each change checked by the InputRecorder and the owner's session
// by owners.
type Settings struct {
	recorder  *InputRecorder
	store     SettingsReader
	owners    Owners
	pack      jurisdiction.Pack
	reporting kernel.Currency
}

// NewSettings returns the settings use cases over store and recorder, with
// pack's defaults and wrappers, the configured reporting currency, and the
// owner's sessions checked by owners.
func NewSettings(recorder *InputRecorder, store SettingsReader, owners Owners, pack jurisdiction.Pack, reporting kernel.Currency) *Settings {
	return &Settings{recorder: recorder, store: store, owners: owners, pack: pack, reporting: reporting}
}

// Display returns the reporting currency and the number format: the
// owner's choice, or the pack's default until the owner makes one.
func (s *Settings) Display(ctx context.Context) (DisplaySettings, error) {
	f, chosen, err := s.store.NumberFormat(ctx)
	if err != nil {
		return DisplaySettings{}, fmt.Errorf("loading the number format: %w", err)
	}
	if !chosen {
		f = s.pack.NumberFormat()
	}
	return DisplaySettings{ReportingCurrency: s.reporting, NumberFormat: f, NumberFormatDefault: !chosen}, nil
}

// Owner returns every setting to the owner: a session of owner scope on
// origin.
func (s *Settings) Owner(ctx context.Context, origin, sessionToken string) (OwnerSettings, error) {
	if err := s.checkOwner(ctx, origin, sessionToken); err != nil {
		return OwnerSettings{}, err
	}
	display, err := s.Display(ctx)
	if err != nil {
		return OwnerSettings{}, err
	}
	values, err := s.store.ValueSettings(ctx)
	if err != nil {
		return OwnerSettings{}, fmt.Errorf("loading the owner's settings: %w", err)
	}
	return OwnerSettings{DisplaySettings: display, Wrappers: s.pack.Wrappers(), Values: values}, nil
}

// SetNumberFormat makes f the owner's number format, for the owner. A
// format breaking NumberFormat.Validate is an error wrapping
// ErrInvalidInput.
func (s *Settings) SetNumberFormat(ctx context.Context, origin, sessionToken string, f jurisdiction.NumberFormat) error {
	if err := s.checkOwner(ctx, origin, sessionToken); err != nil {
		return err
	}
	if err := s.recorder.RecordNumberFormat(ctx, f); err != nil {
		return fmt.Errorf("setting the number format: %w", err)
	}
	return nil
}

// SetMarginalRate makes rate the owner's marginal income tax rate, for the
// owner, and revalues every account (RecordMarginalRate). A rate not above
// 0 and below 1 is an error wrapping ErrInvalidInput.
func (s *Settings) SetMarginalRate(ctx context.Context, origin, sessionToken string, rate kernel.Decimal) error {
	if err := s.checkOwner(ctx, origin, sessionToken); err != nil {
		return err
	}
	if err := s.recorder.RecordMarginalRate(ctx, rate); err != nil {
		return fmt.Errorf("setting the marginal rate: %w", err)
	}
	return nil
}

// checkOwner checks that the session on origin is the owner's.
func (s *Settings) checkOwner(ctx context.Context, origin, sessionToken string) error {
	if err := s.owners.CheckOwner(ctx, origin, sessionToken); err != nil {
		return fmt.Errorf("the owner's settings: %w", err)
	}
	return nil
}
