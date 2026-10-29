---
title: The EODHD connector must identify a market index by the name the owner gives it, such as WIG20.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-96-1
  - lane:strict
text_ref: design-decisions:D-96#1
origin:
  kind: import
  ref: design-decisions:D-96#1
  recorded_at: '2026-10-29T09:42:17.158Z'
semantic_text: The EODHD connector must identify a market index by the name the owner gives it, such as WIG20.
logic_claims:
  - CLAIM-6BF929BD377A6379
semantic_clauses:
  - The EODHD connector must identify a market index by the name the owner gives it, such as WIG20.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: e8d7a35c519b84222a795ed8f0e945539e01ea7f41950d00920f3ca8ee47e447
semantic_inventory:
  - claim_key: CLAIM-6BF929BD377A6379
    claim_text: The EODHD connector must identify a market index by the name the owner gives it, such as WIG20
    role: normative
    status: modeled
    span:
      start: 0
      end: 94
id: REQ-plugins-eodhd-index-by-name
type: req
---
The EODHD connector must identify a market index by the name the owner gives it, such as WIG20.

## Context

D-51 makes EODHD the first market-data connector, including end-of-day index history, and D-52 lets a model portfolio target be an index. D-96 settles how the connector names an index: by the name the owner gives it, such as WIG20, rather than by an identifier of EODHD's. REQ-plugins-eodhd-sync, which the author has not reviewed, delivers the levels of each configured market index. The author recorded D-96 to D-99 after the design session that answered the implementation's questions.

## Source

> The EODHD connector identifies a market index by the name the owner gives it, such as WIG20. Its Euronext covers EODHD's Paris, Amsterdam, Brussels, Lisbon, Dublin, Milan and Oslo markets. An ISIN listed on several supported exchanges stays unpriced and is reported as ambiguous until the owner picks the exchange, and an instrument it can't price is reported again on every sync.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-96#1
