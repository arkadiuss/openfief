---
title: 'The Ghostfolio view must make each treasury bond a MANUAL asset whose symbol is a UUID derived from the view instance''s name, the Ghostfolio user''s ID and the series code and whose name is the code, must key each currency''s fee asset the same way from the currency, and must send each bond''s full price history on every sync: each stored valuation divided by the quantity held at the end of that day.'
status: open
text_ref: docs/DESIGN.md#D-64
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-64
  - lane:strict
semantic_text: 'The Ghostfolio view must make each treasury bond a MANUAL asset whose symbol is a UUID derived from the view instance''s name, the Ghostfolio user''s ID and the series code and whose name is the code, must key each currency''s fee asset the same way from the currency, and must send each bond''s full price history on every sync: each stored valuation divided by the quantity held at the end of that day.'
logic_claims:
  - CLAIM-D44C838B03B0EBDD
semantic_clauses:
  - 'The Ghostfolio view must make each treasury bond a MANUAL asset whose symbol is a UUID derived from the view instance''s name, the Ghostfolio user''s ID and the series code and whose name is the code, must key each currency''s fee asset the same way from the currency, and must send each bond''s full price history on every sync: each stored valuation divided by the quantity held at the end of that day.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 72d95f2bf4c35c7c0527aca6f9296e8f224a7a8146c0837e319e0dd57d86bb50
semantic_inventory:
  - claim_key: CLAIM-D44C838B03B0EBDD
    claim_text: 'The Ghostfolio view must make each treasury bond a MANUAL asset whose symbol is a UUID derived from the view instance''s name, the Ghostfolio user''s ID and the series code and whose name is the code, must key each currency''s fee asset the same way from the currency, and must send each bond''s full price history on every sync: each stored valuation divided by the quantity held at the end of that day'
    role: normative
    status: modeled
    span:
      start: 0
      end: 399
origin:
  kind: agent
  recorded_at: '2026-10-30T01:21:14.957Z'
id: REQ-views-ghostfolio-asset-keys
type: req
---
The Ghostfolio view must make each treasury bond a MANUAL asset whose symbol is a UUID derived from the view instance's name, the Ghostfolio user's ID and the series code and whose name is the code, must key each currency's fee asset the same way from the currency, and must send each bond's full price history on every sync: each stored valuation divided by the quantity held at the end of that day.

## Context

Refines the firm REQ-views-ghostfolio-bond-asset-symbol (D-64 amended) and REQ-views-ghostfolio-standalone-fee (D-108). Ghostfolio's asset symbols are unique across its users and two installs may name their instances alike, so the user's ID joins the derivation; that is this package's reading (docs/implementation-questions.md, Ghostfolio view). Modeled by Claude Code for the Ghostfolio work package (impl/ghostfolio-2); the author has not reviewed it yet.

## Source

> - A bond is a `MANUAL` asset whose symbol is a UUID derived from the plugin instance and its series code, so two installs that share one Ghostfolio never collide; the code is its name. Every sync sends its full price history.
> - Each sync imports missing activities first, then deletes stale ones, because Ghostfolio deletes an asset and its price history along with its last activity. The one exception is a stale activity that Ghostfolio's duplicate check would match against an import: it is deleted first, or the sync would never settle.

docs/DESIGN.md#D-64

> The Ghostfolio plugin books a dividend as quantity 1 at the gross amount, with the withheld tax as its fee, and a standalone fee against a manual "OpenFief fees" asset per currency. Each of its Ghostfolio accounts' cash balance follows OpenFief's cash, so Ghostfolio's totals include it. Listed instruments need the owner's mapping from ISIN to Ghostfolio's data source and symbol.

docs/DESIGN.md#D-108

## Why this changed

D-64 was amended: a bond's MANUAL asset symbol now derives from the plugin instance as well as the series code, so two installs sharing one Ghostfolio never collide. This package also keys the fee assets of D-108 the same way and adds the Ghostfolio user's ID, since symbols are unique across all of Ghostfolio's users.
