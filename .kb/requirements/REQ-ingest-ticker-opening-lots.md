---
title: A ticker entry may be an opening lot carrying the date its units were first acquired; once its ticker's ISIN is confirmed it must be added to its account's seeding on the go-live date, and an opening batch of a broker account already seeded must leave it out.
status: open
text_ref: docs/DESIGN.md#D-136
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-136
  - lane:strict
semantic_text: A ticker entry may be an opening lot carrying the date its units were first acquired; once its ticker's ISIN is confirmed it must be added to its account's seeding on the go-live date, and an opening batch of a broker account already seeded must leave it out.
logic_claims:
  - CLAIM-5DC10116F6894036
semantic_clauses:
  - A ticker entry may be an opening lot carrying the date its units were first acquired; once its ticker's ISIN is confirmed it must be added to its account's seeding on the go-live date, and an opening batch of a broker account already seeded must leave it out.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 0fb1994ef85369eed421e1779d39112331dc4701ad9efc4f8bbe968e1c3faa70
semantic_inventory:
  - claim_key: CLAIM-5DC10116F6894036
    claim_text: A ticker entry may be an opening lot carrying the date its units were first acquired; once its ticker's ISIN is confirmed it must be added to its account's seeding on the go-live date, and an opening batch of a broker account already seeded must leave it out
    role: normative
    status: modeled
    span:
      start: 0
      end: 258
origin:
  kind: agent
  recorded_at: '2026-11-01T07:36:08.764Z'
id: REQ-ingest-ticker-opening-lots
type: req
---
A ticker entry may be an opening lot carrying the date its units were first acquired; once its ticker's ISIN is confirmed it must be added to its account's seeding on the go-live date, and an opening batch of a broker account already seeded must leave it out.

## Context

D-136 seeds opening lots from XTB positions, which name tickers only and wait for the owner's ISIN (D-116); confirmed ones join the seeding through the ledger's one-lot-at-a-time path that hand-entered bond lots take today. Recorded by the XTB start implementation agent (impl/xtb-start); default firmness, not reviewed by the author yet; it refines REQ-ingest-xtb-start-from-open-positions.

## Source

> An XTB account starts from an export covering its whole history, or from the open positions of a current export, which seed one opening lot per position at its open time and price, with the cash balance the owner enters once.

docs/DESIGN.md#D-136
