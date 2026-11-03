-- Stored valuations name their currency and stored exchange rates their base currency, both taken from the
-- jurisdiction pack; neither is assumed (docs/DESIGN.md D-180). A valuation's gross value in the reporting currency
-- moves from pln to reporting, with reporting_currency naming the currency of it and of the net, liquidation, tax cost
-- and sale cost values; each FX rate names its base, and its key includes it.
--
-- Nothing stored before the first release is converted (D-161): stored valuations and rates name no currency, so they
-- go, with the valued-through dates they set and the recorded rules version, so the core revalues every account at
-- start-up (app.InputRecorder.ApplyRules). An installation syncs its FX connector again from its configured start to
-- get its rates back; the demo reseeds itself.
DELETE FROM valuation.holding_valuations;
DELETE FROM valuation.account_status;
DELETE FROM valuation.settings WHERE key = 'rules_version';
DELETE FROM valuation.fx_rates;

ALTER TABLE valuation.holding_valuations DROP CONSTRAINT holding_valuations_check;
ALTER TABLE valuation.holding_valuations RENAME COLUMN pln TO reporting;
ALTER TABLE valuation.holding_valuations
    ADD COLUMN reporting_currency char(3),
    ADD CONSTRAINT holding_valuations_valued_check
        CHECK ((missing IS NULL) = (value IS NOT NULL AND value_currency IS NOT NULL AND reporting IS NOT NULL
                                    AND reporting_currency IS NOT NULL));

ALTER TABLE valuation.fx_rates DROP CONSTRAINT fx_rates_currency_check;
ALTER TABLE valuation.fx_rates DROP CONSTRAINT fx_rates_pkey;
ALTER TABLE valuation.fx_rates
    ADD COLUMN base char(3) NOT NULL,
    ADD CONSTRAINT fx_rates_base_check CHECK (currency <> base),
    ADD PRIMARY KEY (currency, base, date, source);
