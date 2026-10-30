---
title: Each Ghostfolio sync must keep an activity of an owned account only when a desired activity agrees with it in every field, comment included, import every desired activity nothing agrees with, delete every other activity of the owned accounts, deleting before the imports only those an import would skip as duplicates and, when it imports onto a MANUAL asset, those on MANUAL assets it does not desire; must empty with one account-filtered bulk delete an owned account that should hold nothing; must set every desired cash balance an owned account lacks or holds with another value and delete its balances on other dates; and must leave nothing to change for a second sync.
status: open
text_ref: docs/DESIGN.md#D-64
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-64
  - lane:strict
semantic_text: Each Ghostfolio sync must keep an activity of an owned account only when a desired activity agrees with it in every field, comment included, import every desired activity nothing agrees with, delete every other activity of the owned accounts, deleting before the imports only those an import would skip as duplicates and, when it imports onto a MANUAL asset, those on MANUAL assets it does not desire; must empty with one account-filtered bulk delete an owned account that should hold nothing; must set every desired cash balance an owned account lacks or holds with another value and delete its balances on other dates; and must leave nothing to change for a second sync.
logic_claims:
  - CLAIM-C3D542B224EBA9F9
semantic_clauses:
  - Each Ghostfolio sync must keep an activity of an owned account only when a desired activity agrees with it in every field, comment included, import every desired activity nothing agrees with, delete every other activity of the owned accounts, deleting before the imports only those an import would skip as duplicates and, when it imports onto a MANUAL asset, those on MANUAL assets it does not desire; must empty with one account-filtered bulk delete an owned account that should hold nothing; must set every desired cash balance an owned account lacks or holds with another value and delete its balances on other dates; and must leave nothing to change for a second sync.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 16653df8525d2c27582af1084f575204b5df0a55f84f65445c368e7df0349bd7
semantic_inventory:
  - claim_key: CLAIM-C3D542B224EBA9F9
    claim_text: Each Ghostfolio sync must keep an activity of an owned account only when a desired activity agrees with it in every field, comment included, import every desired activity nothing agrees with, delete every other activity of the owned accounts, deleting before the imports only those an import would skip as duplicates and, when it imports onto a MANUAL asset, those on MANUAL assets it does not desire; must empty with one account-filtered bulk delete an owned account that should hold nothing; must set every desired cash balance an owned account lacks or holds with another value and delete its balances on other dates; and must leave nothing to change for a second sync
    role: normative
    status: modeled
    span:
      start: 0
      end: 671
origin:
  kind: agent
  recorded_at: '2026-10-30T01:21:20.977Z'
id: REQ-views-ghostfolio-reconcile-cash
type: req
---
Each Ghostfolio sync must keep an activity of an owned account only when a desired activity agrees with it in every field, comment included, import every desired activity nothing agrees with, delete every other activity of the owned accounts, deleting before the imports only those an import would skip as duplicates and, when it imports onto a MANUAL asset, those on MANUAL assets it does not desire; must empty with one account-filtered bulk delete an owned account that should hold nothing; must set every desired cash balance an owned account lacks or holds with another value and delete its balances on other dates; and must leave nothing to change for a second sync.

## Context

Refines the firm D-4 and D-64 requirements and REQ-views-ghostfolio-cash-balance (D-108). Ghostfolio's import reuses the user's MANUAL asset with the name of a new one, so the bond assets an earlier version keyed by the series code alone would capture every import of their re-keyed successors and the sync would never settle; deleting them first is a second exception to the firm REQ-views-ghostfolio-import-first-blocking-delete, recorded as an open question for the author (docs/implementation-questions.md, Ghostfolio view). Modeled by Claude Code for the Ghostfolio work package (impl/ghostfolio-2); the author has not reviewed it yet.

## Source

> - A bond is a `MANUAL` asset whose symbol is a UUID derived from the plugin instance and its series code, so two installs that share one Ghostfolio never collide; the code is its name. Every sync sends its full price history.
> - Each sync imports missing activities first, then deletes stale ones, because Ghostfolio deletes an asset and its price history along with its last activity. The one exception is a stale activity that Ghostfolio's duplicate check would match against an import: it is deleted first, or the sync would never settle.

docs/DESIGN.md#D-64

> The Ghostfolio plugin books a dividend as quantity 1 at the gross amount, with the withheld tax as its fee, and a standalone fee against a manual "OpenFief fees" asset per currency. Each of its Ghostfolio accounts' cash balance follows OpenFief's cash, so Ghostfolio's totals include it. Listed instruments need the owner's mapping from ISIN to Ghostfolio's data source and symbol.

docs/DESIGN.md#D-108
