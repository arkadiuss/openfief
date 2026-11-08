---
title: Until the ministry's data lists a series' ISIN, or while it can't be read and no lot needs its rates, the treasury bond connector must book the series' lots under its provisional ISIN, PL, the series code, 00 and a check digit, keyed by series and purchase date as under the registry's ISIN.
status: open
text_ref: docs/DESIGN.md#D-214
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-214
  - lane:strict
semantic_text: Until the ministry's data lists a series' ISIN, or while it can't be read and no lot needs its rates, the treasury bond connector must book the series' lots under its provisional ISIN, PL, the series code, 00 and a check digit, keyed by series and purchase date as under the registry's ISIN.
logic_claims:
  - CLAIM-15B1643C9851CF3C
semantic_clauses:
  - Until the ministry's data lists a series' ISIN, or while it can't be read and no lot needs its rates, the treasury bond connector must book the series' lots under its provisional ISIN, PL, the series code, 00 and a check digit, keyed by series and purchase date as under the registry's ISIN.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 82f3e1bd6307c248084078d19bcddbf9e48556cc721228b82aa727aab9e5dc5b
semantic_inventory:
  - claim_key: CLAIM-15B1643C9851CF3C
    claim_text: Until the ministry's data lists a series' ISIN, or while it can't be read and no lot needs its rates, the treasury bond connector must book the series' lots under its provisional ISIN, PL, the series code, 00 and a check digit, keyed by series and purchase date as under the registry's ISIN
    role: normative
    status: modeled
    span:
      start: 0
      end: 290
origin:
  kind: agent
  recorded_at: '2026-11-07T22:45:08.655Z'
id: REQ-treasurybonds-provisional-isin
type: req
---
Until the ministry's data lists a series' ISIN, or while it can't be read and no lot needs its rates, the treasury bond connector must book the series' lots under its provisional ISIN, PL, the series code, 00 and a check digit, keyed by series and purchase date as under the registry's ISIN.

## Context

D-214 has a treasury bond identified by the registry's ISIN from the Ministry of Finance's data, and a lot of a series the data doesn't list yet wait for its ISIN as it waits for its terms; installs from before it hold every series under the ISIN D-99 derived from the series code, and D-161 asks that stored data be migrated. Modeled by Claude Code for D-214; the author has not reviewed the wording.

## Source

> A lot of a series the data doesn't list yet waits, like its terms, until its ISIN arrives, and a held series still without its ISIN or terms after seven days raises an attention item.

docs/DESIGN.md#D-214
