-- Three values per holding (docs/DESIGN.md D-77), tax wrappers from the
-- jurisdiction pack (D-76), fee rules and the owner's settings (D-78,
-- D-113). D-77 replaced the bond basis setting (D-25): every stored
-- valuation now holds its gross value (value and pln, on the market or
-- gross_accrued basis) with its net and liquidation values in PLN.

-- Rows stored before hold one value on a basis and can't be read as three,
-- so they go, with the valued-through dates they set. The core revalues
-- every account when it records the rules version at start-up
-- (app.InputRecorder.ApplyRules), since none is recorded yet.
DELETE FROM valuation.holding_valuations;
DELETE FROM valuation.account_status;
DELETE FROM valuation.settings WHERE key = 'bond_basis';

ALTER TABLE valuation.holding_valuations
    DROP CONSTRAINT holding_valuations_basis_check,
    ADD CONSTRAINT holding_valuations_basis_check CHECK (basis IN ('market', 'gross_accrued')),
    ADD COLUMN net numeric,
    ADD COLUMN liquidation numeric,
    ADD COLUMN net_missing text CHECK (net_missing IN ('fx_rate', 'lots')),
    ADD COLUMN liquidation_missing text CHECK (liquidation_missing IN ('fx_rate', 'marginal_rate', 'lots')),
    ADD CONSTRAINT holding_valuations_net_check
        CHECK ((net IS NOT NULL) = (missing IS NULL AND net_missing IS NULL)),
    ADD CONSTRAINT holding_valuations_liquidation_check
        CHECK ((liquidation IS NOT NULL) = (missing IS NULL AND liquidation_missing IS NULL));

-- The jurisdiction pack names the wrappers (OKI joins taxable, IKE and
-- IKZE); the store keeps any key the pack accepted.
ALTER TABLE valuation.tax_wrappers DROP CONSTRAINT tax_wrappers_wrapper_check;
ALTER TABLE valuation.tax_wrappers ADD CONSTRAINT tax_wrappers_wrapper_check CHECK (wrapper <> '');

-- Each account's fee rule (D-78); an account without a row charges
-- nothing. Rates are fractions.
CREATE TABLE valuation.fee_rules (
    account_id       text PRIMARY KEY CHECK (account_id <> ''),
    rate             numeric NOT NULL CHECK (rate >= 0 AND rate < 1),
    minimum          numeric NOT NULL CHECK (minimum >= 0),
    minimum_currency char(3),
    conversion_rate  numeric NOT NULL CHECK (conversion_rate >= 0 AND conversion_rate < 1),
    CHECK (minimum = 0 OR minimum_currency IS NOT NULL)
);

-- valuation.settings now also holds the rules version (rules_version), the
-- owner's marginal income tax rate (marginal_rate) and number format
-- (number_format, which revalues nothing).
COMMENT ON TABLE valuation.settings IS
    'The rules version, the owner''s marginal income tax rate and number format';
