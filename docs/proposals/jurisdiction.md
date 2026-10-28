# Proposals: jurisdiction

Not decided. When the author decides an item, it moves into [DESIGN.md](../DESIGN.md) as a decision and leaves this file.

This area covers GR-13: regulatory and tax logic as jurisdiction plugins, Poland first. Bond valuation and NBP rates are already decided for M0 (D-9, D-20, D-25).

- **[P-39] Polish pack, first version.** PIT-38 and PIT/ZG data, IKE and IKZE rules, alongside the decided bond valuation. It brings the full history backfill that D-23 defers, since past tax years need complete lot history (see R-9 for the tax rules).
- **[P-40] After-tax liquidation value.** D-25 counts bonds in totals net of tax while stocks are gross, so the total mixes two bases. With the tax pack, "net" becomes a portfolio-wide after-tax liquidation value: stocks also lose Belka tax on their unrealized PLN gain, per lot at NBP rates. Until then, the UI marks the total as mixed.
- **[P-41] Jurisdiction pack contract.** Interface outline, not code.

  | Aspect | Outline |
  |---|---|
  | Purpose | All country-specific regulatory and tax logic (GR-13). |
  | Manifest | Country, supported tax years, version, rule sources. |
  | Operations | Classify accounts and tax wrappers (IKE, IKZE). Value instruments with statutory valuation (bond series terms, CPI, NBP reference rate). Compute realized gains per tax year (FIFO, NBP table A rate from the day before). Produce tax-form data (PIT-38, PIT/ZG per country). Supply local limits to the policy engine as policy data. |
  | Inputs | Normalized transactions and holdings; reference data (CPI, exchange rates). |
  | Outputs | Deterministic, reproducible results carrying the rule version. |
  | Models | None. Deterministic code. |
