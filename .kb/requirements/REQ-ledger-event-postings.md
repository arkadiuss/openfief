---
title: Buys, sells, dividends and interest with withholding tax, fees, FX conversions, deposits, withdrawals and opening lots must each be booked as balanced postings in their original currencies, with no conversion to PLN.
status: open
text_ref: docs/DESIGN.md#D-20
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-20
  - lane:strict
semantic_text: Buys, sells, dividends and interest with withholding tax, fees, FX conversions, deposits, withdrawals and opening lots must each be booked as balanced postings in their original currencies, with no conversion to PLN.
logic_claims:
  - CLAIM-B17A46D82DE7520B
semantic_clauses:
  - Buys, sells, dividends and interest with withholding tax, fees, FX conversions, deposits, withdrawals and opening lots must each be booked as balanced postings in their original currencies, with no conversion to PLN.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 764ae3ed240caa2ee72ebb791fa2b458bc4ecc3dbe33abbb911895a7a65ec202
semantic_inventory:
  - claim_key: CLAIM-B17A46D82DE7520B
    claim_text: Buys, sells, dividends and interest with withholding tax, fees, FX conversions, deposits, withdrawals and opening lots must each be booked as balanced postings in their original currencies, with no conversion to PLN
    role: normative
    status: modeled
    span:
      start: 0
      end: 215
origin:
  kind: agent
  recorded_at: '2026-10-28T05:24:51.843Z'
id: REQ-ledger-event-postings
type: req
---
Buys, sells, dividends and interest with withholding tax, fees, FX conversions, deposits, withdrawals and opening lots must each be booked as balanced postings in their original currencies, with no conversion to PLN.

## Context

D-19 and D-33 give the ledger accounts; D-101 adds an interest kind for bond coupons, refined here from the earlier list that had no interest. Recorded by the bonds implementation agent (impl/bonds-2).

## Source

> The ledger has an interest transaction kind. Bond coupons, such as COI's yearly interest, are booked as interest, not as dividends.

docs/DESIGN.md#D-101
