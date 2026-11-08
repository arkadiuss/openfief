---
title: Every Instrument a read API answer names must carry isin_provisional, true when its ISIN is the provisional ISIN of the treasury bond series its symbol names, and an InstrumentDetail true when its ISIN is any series' provisional ISIN.
status: open
text_ref: docs/DESIGN.md#D-214
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-214
  - lane:strict
semantic_text: Every Instrument a read API answer names must carry isin_provisional, true when its ISIN is the provisional ISIN of the treasury bond series its symbol names, and an InstrumentDetail true when its ISIN is any series' provisional ISIN.
logic_claims:
  - CLAIM-2E4F3A184A6CADBF
semantic_clauses:
  - Every Instrument a read API answer names must carry isin_provisional, true when its ISIN is the provisional ISIN of the treasury bond series its symbol names, and an InstrumentDetail true when its ISIN is any series' provisional ISIN.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: d040120adfaf31e78dd3a60d9828e482cf61a620dfa58136e5d86a9adf4783b7
semantic_inventory:
  - claim_key: CLAIM-2E4F3A184A6CADBF
    claim_text: Every Instrument a read API answer names must carry isin_provisional, true when its ISIN is the provisional ISIN of the treasury bond series its symbol names, and an InstrumentDetail true when its ISIN is any series' provisional ISIN
    role: normative
    status: modeled
    span:
      start: 0
      end: 233
origin:
  kind: agent
  recorded_at: '2026-11-08T02:00:45.805Z'
id: REQ-views-instrument-isin-provisional
type: req
---
Every Instrument a read API answer names must carry isin_provisional, true when its ISIN is the provisional ISIN of the treasury bond series its symbol names, and an InstrumentDetail true when its ISIN is any series' provisional ISIN.

## Context

D-214 books a lot of a series the Ministry of Finance's data doesn't list yet under a provisional ISIN until the registry's arrives (REQ-treasurybonds-lot-provisional-isin-until-listed). The design session asked, in the D-214 follow-ups, that the UI show a Pending pill next to such a bond wherever its ISIN appears, with a tooltip saying it moves to the registry's ISIN when the ministry's data lists it, using the design system's existing pill. The UI needs the read API to say so, since it holds no jurisdiction rules. Modeled by Claude Code for the D-214 follow-ups; the author has not reviewed the wording.

## Source

> A treasury bond is identified by its real ISIN, the registry's, from the Ministry of Finance's data (D-97); the series code is its ticker. A lot of a series the data doesn't list yet is booked under a provisional ISIN, `PL`, the series code, `00` and a check digit, and moves to the registry's ISIN when the data lists it.

docs/DESIGN.md#D-214
