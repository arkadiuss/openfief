---
title: The EODHD connector's Euronext must cover EODHD's Paris, Amsterdam, Brussels, Lisbon, Dublin, Milan and Oslo markets.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-96-2
  - lane:strict
text_ref: design-decisions:D-96#2
origin:
  kind: import
  ref: design-decisions:D-96#2
  recorded_at: '2026-10-29T09:42:56.504Z'
semantic_text: The EODHD connector's Euronext must cover EODHD's Paris, Amsterdam, Brussels, Lisbon, Dublin, Milan and Oslo markets.
logic_claims:
  - CLAIM-07BF9CE680AFE995
semantic_clauses:
  - The EODHD connector's Euronext must cover EODHD's Paris, Amsterdam, Brussels, Lisbon, Dublin, Milan and Oslo markets.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 2db4bc1110f4b32ea068b88a4fb2c4e34d95b3b63c470037d666546fa4572b1c
semantic_inventory:
  - claim_key: CLAIM-07BF9CE680AFE995
    claim_text: The EODHD connector's Euronext must cover EODHD's Paris, Amsterdam, Brussels, Lisbon, Dublin, Milan and Oslo markets
    role: normative
    status: modeled
    span:
      start: 0
      end: 116
id: REQ-plugins-eodhd-euronext-markets
type: req
---
The EODHD connector's Euronext must cover EODHD's Paris, Amsterdam, Brussels, Lisbon, Dublin, Milan and Oslo markets.

## Context

D-51 has EODHD cover the Warsaw Stock Exchange, LSE, Xetra and Euronext. EODHD lists Euronext as separate markets, so D-96 names the ones the connector counts as Euronext: Paris, Amsterdam, Brussels, Lisbon, Dublin, Milan and Oslo. REQ-valuation-eodhd-listing, which the author has not reviewed, looks a listing up on a Euronext market without naming them. The author recorded D-96 to D-99 after the design session that answered the implementation's questions.

## Source

> The EODHD connector identifies a market index by the name the owner gives it, such as WIG20. Its Euronext covers EODHD's Paris, Amsterdam, Brussels, Lisbon, Dublin, Milan and Oslo markets. An ISIN listed on several supported exchanges stays unpriced and is reported as ambiguous until the owner picks the exchange, and an instrument it can't price is reported again on every sync.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-96#2
