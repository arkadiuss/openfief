---
title: 'Confirming a ticker whose listing names no trading currency, on an exchange that quotes in more than one, must require the owner''s pick of one of that exchange''s currencies: a confirmation without one, or with a pick where there is nothing to pick, must be refused with INVALID_ARGUMENT naming the choices and record nothing, so the ticker''s entries keep waiting.'
status: closed
text_ref: docs/DESIGN.md#D-170
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-170
  - lane:strict
semantic_text: 'Confirming a ticker whose listing names no trading currency, on an exchange that quotes in more than one, must require the owner''s pick of one of that exchange''s currencies: a confirmation without one, or with a pick where there is nothing to pick, must be refused with INVALID_ARGUMENT naming the choices and record nothing, so the ticker''s entries keep waiting.'
logic_claims:
  - CLAIM-5491D9D37B051211
semantic_clauses:
  - 'Confirming a ticker whose listing names no trading currency, on an exchange that quotes in more than one, must require the owner''s pick of one of that exchange''s currencies: a confirmation without one, or with a pick where there is nothing to pick, must be refused with INVALID_ARGUMENT naming the choices and record nothing, so the ticker''s entries keep waiting.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 7271be3eea914e37236112d9b0bc9a2af1b9261a3dd866e144beab60aeecd814
semantic_inventory:
  - claim_key: CLAIM-5491D9D37B051211
    claim_text: 'Confirming a ticker whose listing names no trading currency, on an exchange that quotes in more than one, must require the owner''s pick of one of that exchange''s currencies: a confirmation without one, or with a pick where there is nothing to pick, must be refused with INVALID_ARGUMENT naming the choices and record nothing, so the ticker''s entries keep waiting'
    role: normative
    status: modeled
    span:
      start: 0
      end: 362
origin:
  kind: agent
  recorded_at: '2026-11-01T23:38:16.579Z'
id: REQ-ingest-ticker-currency-pick-required
type: req
---
Confirming a ticker whose listing names no trading currency, on an exchange that quotes in more than one, must require the owner's pick of one of that exchange's currencies: a confirmation without one, or with a pick where there is nothing to pick, must be refused with INVALID_ARGUMENT naming the choices and record nothing, so the ticker's entries keep waiting.

## Context

D-170 has the owner pick the trading currency when confirming the ticker and its positions wait until then; the work package made the request's currency optional and required only when ambiguous. Recorded by the implementation agent of D-170 (impl/d170) with default firmness: no firm requirement for D-170 existed in kibi when it was built, so this traces the code to the decision directly until the design session imports the firm one.

## Source

> When the listing gives no currency and the exchange its market suffix names quotes in more than one, as London does in pence, pounds and dollars, the owner picks the trading currency when confirming the ticker's ISIN (D-116), and its positions wait until then; no currency is assumed.

docs/DESIGN.md#D-170
