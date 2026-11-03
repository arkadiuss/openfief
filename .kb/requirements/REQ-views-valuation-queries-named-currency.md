---
title: The read API must return an account's stored valuations dated in an inclusive date range, each with its basis, source, original-currency value and its value in the reporting currency it names, with totals per date, and the whole portfolio's totals per date summed over every account.
status: open
text_ref: docs/DESIGN.md#D-180
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-180
  - lane:strict
semantic_text: The read API must return an account's stored valuations dated in an inclusive date range, each with its basis, source, original-currency value and its value in the reporting currency it names, with totals per date, and the whole portfolio's totals per date summed over every account.
logic_claims:
  - CLAIM-D458D59774CA93E4
semantic_clauses:
  - The read API must return an account's stored valuations dated in an inclusive date range, each with its basis, source, original-currency value and its value in the reporting currency it names, with totals per date, and the whole portfolio's totals per date summed over every account.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: daa46eb3c32553b1fd2c5d02d03379db656f9e1a47db54375888e6c14dad60f8
semantic_inventory:
  - claim_key: CLAIM-D458D59774CA93E4
    claim_text: The read API must return an account's stored valuations dated in an inclusive date range, each with its basis, source, original-currency value and its value in the reporting currency it names, with totals per date, and the whole portfolio's totals per date summed over every account
    role: normative
    status: modeled
    span:
      start: 0
      end: 282
origin:
  kind: agent
  recorded_at: '2026-11-02T23:55:37.012Z'
id: REQ-views-valuation-queries-named-currency
type: req
---
The read API must return an account's stored valuations dated in an inclusive date range, each with its basis, source, original-currency value and its value in the reporting currency it names, with totals per date, and the whole portfolio's totals per date summed over every account.

## Why this changed

On 2026-11-02 the author recorded D-180: stored valuations name their currency and stored exchange rates their base currency, both taken from the jurisdiction pack, and neither is assumed to be PLN. REQ-views-valuation-queries stated it in PLN ("The read API must return an account's stored valuations dated in an inclusive date range, each with its basis, source, original-currency value and PLN value, with totals per date, and the whole portfolio's totals per date summed over every account."); this requirement replaces it, naming the currency instead of assuming PLN.

## Context

The Polish pack's reporting currency is PLN, so its behavior is unchanged; another pack's currency flows the same way, and a reporting currency the pack can't convert to stays refused (T-12).

## Source

> Stored valuations name their currency and stored exchange rates their base currency, both taken from the jurisdiction pack; neither is assumed to be PLN. A reporting currency the pack can't convert to stays refused until T-12 is resolved.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-180#1

