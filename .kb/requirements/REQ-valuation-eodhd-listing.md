---
title: The EODHD adapter must find an instrument's listing by ISIN only on the Warsaw Stock Exchange, the LSE, Xetra or a Euronext market, on the exchange the owner names when there is one, and must report an ISIN with no such listing, or with several and no exchange named, as an error instead of guessing.
status: open
text_ref: docs/DESIGN.md#D-51
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-51
  - lane:strict
semantic_text: The EODHD adapter must find an instrument's listing by ISIN only on the Warsaw Stock Exchange, the LSE, Xetra or a Euronext market, on the exchange the owner names when there is one, and must report an ISIN with no such listing, or with several and no exchange named, as an error instead of guessing.
logic_claims:
  - CLAIM-D907D730BB8022F6
semantic_clauses:
  - The EODHD adapter must find an instrument's listing by ISIN only on the Warsaw Stock Exchange, the LSE, Xetra or a Euronext market, on the exchange the owner names when there is one, and must report an ISIN with no such listing, or with several and no exchange named, as an error instead of guessing.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 60c03190db83f1a76b744e397e1865a5990ce183b8ae1eb511627d7b8b9c2c46
semantic_inventory:
  - claim_key: CLAIM-D907D730BB8022F6
    claim_text: The EODHD adapter must find an instrument's listing by ISIN only on the Warsaw Stock Exchange, the LSE, Xetra or a Euronext market, on the exchange the owner names when there is one, and must report an ISIN with no such listing, or with several and no exchange named, as an error instead of guessing
    role: normative
    status: modeled
    span:
      start: 0
      end: 299
origin:
  kind: agent
  recorded_at: '2026-10-28T21:37:25.866Z'
id: REQ-valuation-eodhd-listing
type: req
---
The EODHD adapter must find an instrument's listing by ISIN only on the Warsaw Stock Exchange, the LSE, Xetra or a Euronext market, on the exchange the owner names when there is one, and must report an ISIN with no such listing, or with several and no exchange named, as an error instead of guessing.

## Context

D-51 names the exchanges EODHD is used for and that it finds instruments by ISIN. An ETF is often listed on several of them in different currencies, so picking one silently could price a holding in the wrong currency; the owner names the exchange instead. Exchange codes are EODHD's (WAR, LSE, XETRA, and PA, AS, BR, LS, IR, MI, OL for Euronext). Refines REQ-valuation-eodhd-connector. Modeled by Claude Code for the EODHD market-data connector task; the author has not reviewed it yet.

## Source

> The first market-data connector uses EODHD. EODHD covers the Warsaw Stock Exchange, LSE, Xetra and Euronext, finds instruments by ISIN, includes end-of-day index history, and its terms allow storing the data for private use. Its closes are aggregated rather than each exchange's official close, which is one more reason for D-22's cross-check.

docs/DESIGN.md D-51
