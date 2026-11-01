---
title: 'The XTB connector must seed an account from a current xStation export in opening mode: one opening lot per open stock or ETF position, on the report''s date, with its volume, open date and open price in its trading currency, by ticker until the owner confirms its ISIN.'
status: open
text_ref: docs/DESIGN.md#D-159
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-159
  - lane:strict
semantic_text: 'The XTB connector must seed an account from a current xStation export in opening mode: one opening lot per open stock or ETF position, on the report''s date, with its volume, open date and open price in its trading currency, by ticker until the owner confirms its ISIN.'
logic_claims:
  - CLAIM-1182E83067890575
semantic_clauses:
  - 'The XTB connector must seed an account from a current xStation export in opening mode: one opening lot per open stock or ETF position, on the report''s date, with its volume, open date and open price in its trading currency, by ticker until the owner confirms its ISIN.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 51c0a680bb2f6cdb2bff67a91f95443da4738bb58b2184483fab6dc038cbcd85
semantic_inventory:
  - claim_key: CLAIM-1182E83067890575
    claim_text: 'The XTB connector must seed an account from a current xStation export in opening mode: one opening lot per open stock or ETF position, on the report''s date, with its volume, open date and open price in its trading currency, by ticker until the owner confirms its ISIN'
    role: normative
    status: modeled
    span:
      start: 0
      end: 267
origin:
  kind: agent
  recorded_at: '2026-11-01T16:50:44.129Z'
id: REQ-ingest-xtb-opening-lot-trading-currency
type: req
---
The XTB connector must seed an account from a current xStation export in opening mode: one opening lot per open stock or ETF position, on the report's date, with its volume, open date and open price in its trading currency, by ticker until the owner confirms its ISIN.

## Context

D-159 replaces the narrow reading that seeded only positions quoted in the account's currency. Recorded by the implementation agent of D-159 to D-162 (impl/d159-162) with default firmness: no firm requirement for these decisions existed in kibi when they were built, so this traces the code to the decision directly until the design session imports the firm ones.

## Why this changed

The author decided this in D-159, which replaces the narrower reading REQ-ingest-xtb-start-open-positions recorded before the decision existed, so that requirement is closed.

## Source

> An XTB position's open price is in its instrument's trading currency, which the export doesn't name. It is the currency of the confirmed ticker's listing (D-116), or, when the listing gives none, the currency of the exchange its market suffix names, and the opening lot's cost is booked in it. A price quoted in a minor unit, such as London's pence, is converted to the major unit.

docs/DESIGN.md#D-159
