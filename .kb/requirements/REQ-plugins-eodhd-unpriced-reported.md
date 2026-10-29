---
title: An instrument the EODHD connector can't price must be reported again on every sync.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-96-4
  - lane:strict
text_ref: design-decisions:D-96#4
origin:
  kind: import
  ref: design-decisions:D-96#4
  recorded_at: '2026-10-29T09:44:23.969Z'
semantic_text: An instrument the EODHD connector can't price must be reported again on every sync.
logic_claims:
  - CLAIM-111657A3896CA63D
semantic_clauses:
  - An instrument the EODHD connector can't price must be reported again on every sync.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 6d8cef8869b5554dd8dc96d1bccbf335720fd23acdb5e2d43dca68ccb6f110c6
semantic_inventory:
  - claim_key: CLAIM-111657A3896CA63D
    claim_text: An instrument the EODHD connector can't price must be reported again on every sync
    role: normative
    status: modeled
    span:
      start: 0
      end: 82
id: REQ-plugins-eodhd-unpriced-reported
type: req
---
An instrument the EODHD connector can't price must be reported again on every sync.

## Context

D-93 has every market-data sync carry the instruments the owner holds (REQ-ingest-market-data-sync-instruments), so an instrument EODHD can't price, whether unlisted, unknown or ambiguous (REQ-plugins-eodhd-ambiguous-isin), comes back on each sync. D-96 has the connector report it each time rather than once, so the owner keeps seeing that it has no price. The author recorded D-96 to D-99 after the design session that answered the implementation's questions.

## Source

> The EODHD connector identifies a market index by the name the owner gives it, such as WIG20. Its Euronext covers EODHD's Paris, Amsterdam, Brussels, Lisbon, Dublin, Milan and Oslo markets. An ISIN listed on several supported exchanges stays unpriced and is reported as ambiguous until the owner picks the exchange, and an instrument it can't price is reported again on every sync.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-96#4
