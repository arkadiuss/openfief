---
title: An ISIN listed on several exchanges the EODHD connector supports must stay unpriced and be reported as ambiguous until the owner picks the exchange.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-96-3
  - lane:strict
text_ref: design-decisions:D-96#3
origin:
  kind: import
  ref: design-decisions:D-96#3
  recorded_at: '2026-10-29T09:43:39.797Z'
semantic_text: An ISIN listed on several exchanges the EODHD connector supports must stay unpriced and be reported as ambiguous until the owner picks the exchange.
logic_claims:
  - CLAIM-19E12CFCD60A76F2
semantic_clauses:
  - An ISIN listed on several exchanges the EODHD connector supports must stay unpriced and be reported as ambiguous until the owner picks the exchange.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: e518fc77aa77c4cf5a50278605b08e893cf0f902811590ec07f7cd78e1df774e
semantic_inventory:
  - claim_key: CLAIM-19E12CFCD60A76F2
    claim_text: An ISIN listed on several exchanges the EODHD connector supports must stay unpriced and be reported as ambiguous until the owner picks the exchange
    role: normative
    status: modeled
    span:
      start: 0
      end: 147
id: REQ-plugins-eodhd-ambiguous-isin
type: req
---
An ISIN listed on several exchanges the EODHD connector supports must stay unpriced and be reported as ambiguous until the owner picks the exchange.

## Context

D-51 has EODHD find instruments by ISIN, and one ISIN can be listed on several of the exchanges the connector supports, each with its own price. D-96 has the connector leave such an instrument unpriced and report it as ambiguous rather than guess, until the owner picks the exchange. REQ-valuation-eodhd-listing and REQ-plugins-eodhd-sync, which the author has not reviewed, already report several listings with no exchange named as an error or ingest problem. The author recorded D-96 to D-99 after the design session that answered the implementation's questions.

## Source

> The EODHD connector identifies a market index by the name the owner gives it, such as WIG20. Its Euronext covers EODHD's Paris, Amsterdam, Brussels, Lisbon, Dublin, Milan and Oslo markets. An ISIN listed on several supported exchanges stays unpriced and is reported as ambiguous until the owner picks the exchange, and an instrument it can't price is reported again on every sync.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-96#3
