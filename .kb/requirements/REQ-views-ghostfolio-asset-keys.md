---
title: 'The Ghostfolio view must make each treasury bond a MANUAL asset whose symbol is a UUID derived from the view instance''s name, the Ghostfolio user''s ID and the series code and whose name is the code plus a suffix no earlier version used, must key each currency''s fee asset the same way from the currency, and must send each bond''s full price history on every sync: each stored valuation divided by the quantity held at the end of that day.'
status: open
text_ref: docs/DESIGN.md#D-64
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-64
  - lane:strict
semantic_text: 'The Ghostfolio view must make each treasury bond a MANUAL asset whose symbol is a UUID derived from the view instance''s name, the Ghostfolio user''s ID and the series code and whose name is the code plus a suffix no earlier version used, must key each currency''s fee asset the same way from the currency, and must send each bond''s full price history on every sync: each stored valuation divided by the quantity held at the end of that day.'
logic_claims:
  - CLAIM-ACA3C85A7BE8C73E
semantic_clauses:
  - 'The Ghostfolio view must make each treasury bond a MANUAL asset whose symbol is a UUID derived from the view instance''s name, the Ghostfolio user''s ID and the series code and whose name is the code plus a suffix no earlier version used, must key each currency''s fee asset the same way from the currency, and must send each bond''s full price history on every sync: each stored valuation divided by the quantity held at the end of that day.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: b06842bdde3b7597bb52d977ece3f1f2a79664e7f15694824d8caed09d674c0f
semantic_inventory:
  - claim_key: CLAIM-ACA3C85A7BE8C73E
    claim_text: 'The Ghostfolio view must make each treasury bond a MANUAL asset whose symbol is a UUID derived from the view instance''s name, the Ghostfolio user''s ID and the series code and whose name is the code plus a suffix no earlier version used, must key each currency''s fee asset the same way from the currency, and must send each bond''s full price history on every sync: each stored valuation divided by the quantity held at the end of that day'
    role: normative
    status: modeled
    span:
      start: 0
      end: 437
origin:
  kind: agent
  recorded_at: '2026-10-30T01:21:14.957Z'
id: REQ-views-ghostfolio-asset-keys
type: req
---
The Ghostfolio view must make each treasury bond a MANUAL asset whose symbol is a UUID derived from the view instance's name, the Ghostfolio user's ID and the series code and whose name is the code plus a suffix no earlier version used, must key each currency's fee asset the same way from the currency, and must send each bond's full price history on every sync: each stored valuation divided by the quantity held at the end of that day.

## Why this changed

D-64 was amended twice. First, a bond's MANUAL asset symbol came to derive from the plugin instance and the Ghostfolio user's ID as well as the series code, so two installs sharing one Ghostfolio never collide, and the fee assets of D-108 are keyed the same way. Then the bond asset's name gained a suffix no earlier version used, because Ghostfolio's import reuses an existing MANUAL asset of the same name: an import never lands on an asset left by an earlier version, so import-first keeps its single exception.

## Context

Refined for D-64 as amended: the bond asset's name is the series code plus ' (OpenFief bond)', which no earlier version used, so Ghostfolio's name-based reuse never lands an import on a leftover asset. Recorded by the implementation agent of impl/decisions-117; default firmness, not reviewed by the author yet.

## Source

> A bond is a `MANUAL` asset whose symbol is a UUID derived from the plugin instance and its series code, so two installs that share one Ghostfolio never collide. Its name is the code plus a suffix no earlier version used, because Ghostfolio's import reuses an existing `MANUAL` asset of the same name; so an import never lands on an asset left by an earlier version, and those are removed after the import like any stale one.

docs/DESIGN.md#D-64
