package app

import (
	"context"
	"fmt"
	"strings"
)

// ChangeAuditor keeps the owner's changes together with their entries in the audit
// log (D-211). cmd/core implements it over the identity context's log, for
// the owner's session that asked.
type ChangeAuditor interface {
	// Audited has change made as the owner's change of kind action, naming
	// refs ("<kind>:<ID>"), entered in the audit log: the entry is kept with
	// the change or neither is, and change is not called when the log can't
	// take the entry. An error of change is returned as it is, wrapped.
	Audited(ctx context.Context, origin, sessionToken, action string, refs []string, change func(context.Context) error) error
}

// The audit action of the owner's settings, as the identity context spells
// it, and the settings it names (identity.Settings).
const (
	auditSettingsChanged       = "settings_changed"
	auditPriceCheckAcknowledge = "price_check_acknowledged"

	settingNumberFormat     = "display:number_format"
	settingMarginalRate     = "tax:marginal_rate"
	settingExemptionClass   = "tax:exemption_class"
	settingEarlyRedemption  = "tax:early_redemption_fee"
	settingChartStart       = "chart:start"
	settingPortfolioStart   = "chart:portfolio_start"
	settingPriceCheck       = "pricecheck:threshold"
	settingSeriesSource     = "marketdata:series_source"
	settingRankingPrefix    = "marketdata:ranking:"
	settingGapPrefix        = "marketdata:gap:"
	auditSettingRef         = "setting:"
	auditInstrumentRef      = "instrument:"
	auditAccountRef         = "account:"
	instrumentSymbolDivider = "/"
)

// audited runs change as the owner's change of kind action, entered in the
// audit log when a is set. Without a ChangeAuditor, as in the tools that build the
// use cases without a core, nothing is entered.
func audited(ctx context.Context, a ChangeAuditor, origin, sessionToken, action string, refs []string, change func(context.Context) error) error {
	if a == nil {
		return change(ctx)
	}
	if err := a.Audited(ctx, origin, sessionToken, action, refs, change); err != nil {
		return fmt.Errorf("entering %s in the audit log: %w", action, err)
	}
	return nil
}

// auditedSetting is audited for a change of the setting, naming also the
// records of extra, each "<kind>:<ID>".
func auditedSetting(ctx context.Context, a ChangeAuditor, origin, sessionToken, setting string, change func(context.Context) error, extra ...string) error {
	return audited(ctx, a, origin, sessionToken, auditSettingsChanged, append([]string{auditSettingRef + setting}, extra...), change)
}

// isinOfInstrument returns the ISIN of an instrument written as
// ledger.Instrument.String() does, "ISIN/symbol" or the ISIN alone.
func isinOfInstrument(instrument string) string {
	isin, _, _ := strings.Cut(instrument, instrumentSymbolDivider)
	return isin
}
