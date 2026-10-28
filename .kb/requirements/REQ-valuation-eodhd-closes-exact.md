---
title: The EODHD adapter must map each day's close, not the adjusted close, to an exact decimal parsed from the JSON number's text and never through a float, leave out days without a positive close, quote it in the listing's currency with pence sterling converted to pounds exactly, and reject a response that is not EODHD's documented end-of-day array.
status: open
text_ref: docs/DESIGN.md#D-51
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-51
  - lane:strict
semantic_text: The EODHD adapter must map each day's close, not the adjusted close, to an exact decimal parsed from the JSON number's text and never through a float, leave out days without a positive close, quote it in the listing's currency with pence sterling converted to pounds exactly, and reject a response that is not EODHD's documented end-of-day array.
logic_claims:
  - CLAIM-DFB9F98A9489C96F
semantic_clauses:
  - The EODHD adapter must map each day's close, not the adjusted close, to an exact decimal parsed from the JSON number's text and never through a float, leave out days without a positive close, quote it in the listing's currency with pence sterling converted to pounds exactly, and reject a response that is not EODHD's documented end-of-day array.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: da5bd1f809e257fb026f6bb2280c1524600ca479b3463cc8e19692bc895ab851
semantic_inventory:
  - claim_key: CLAIM-DFB9F98A9489C96F
    claim_text: The EODHD adapter must map each day's close, not the adjusted close, to an exact decimal parsed from the JSON number's text and never through a float, leave out days without a positive close, quote it in the listing's currency with pence sterling converted to pounds exactly, and reject a response that is not EODHD's documented end-of-day array
    role: normative
    status: modeled
    span:
      start: 0
      end: 345
origin:
  kind: agent
  recorded_at: '2026-10-28T21:37:31.502Z'
id: REQ-valuation-eodhd-closes-exact
type: req
---
The EODHD adapter must map each day's close, not the adjusted close, to an exact decimal parsed from the JSON number's text and never through a float, leave out days without a positive close, quote it in the listing's currency with pence sterling converted to pounds exactly, and reject a response that is not EODHD's documented end-of-day array.

## Context

Prices are never float64 (AGENTS.md, docs/dev/code-organization.md). A stored price is what the instrument closed at that day (D-21), so adjusted_close, which EODHD back-adjusts for splits and dividends, is not used. GBX is not an ISO 4217 code, and the connector contract refuses a non-positive close. Refines REQ-valuation-eodhd-connector and REQ-AUTO-A129ED289B2239A8. Modeled by Claude Code for the EODHD market-data connector task; the author has not reviewed it yet.

## Source

> The first market-data connector uses EODHD. EODHD covers the Warsaw Stock Exchange, LSE, Xetra and Euronext, finds instruments by ISIN, includes end-of-day index history, and its terms allow storing the data for private use. Its closes are aggregated rather than each exchange's official close, which is one more reason for D-22's cross-check.

docs/DESIGN.md D-51
