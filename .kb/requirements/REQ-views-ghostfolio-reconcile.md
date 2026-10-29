---
title: Each Ghostfolio sync must keep an activity of an owned account only when a desired activity agrees with it in every field, comment included, import every desired activity nothing agrees with, delete every other activity of the owned accounts, deleting before the imports only those an import would skip as duplicates, empty with one account-filtered bulk delete an owned account that should hold nothing, and leave nothing to change for a second sync.
status: open
text_ref: docs/DESIGN.md#D-64
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-64
  - lane:strict
semantic_text: Each Ghostfolio sync must keep an activity of an owned account only when a desired activity agrees with it in every field, comment included, import every desired activity nothing agrees with, delete every other activity of the owned accounts, deleting before the imports only those an import would skip as duplicates, empty with one account-filtered bulk delete an owned account that should hold nothing, and leave nothing to change for a second sync.
logic_claims:
  - CLAIM-B85BFCF2D59CA400
semantic_clauses:
  - Each Ghostfolio sync must keep an activity of an owned account only when a desired activity agrees with it in every field, comment included, import every desired activity nothing agrees with, delete every other activity of the owned accounts, deleting before the imports only those an import would skip as duplicates, empty with one account-filtered bulk delete an owned account that should hold nothing, and leave nothing to change for a second sync.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: fe3570d0d2f203debd59e653143932367fb9143698d0c1a840ad4f85c510bea3
semantic_inventory:
  - claim_key: CLAIM-B85BFCF2D59CA400
    claim_text: Each Ghostfolio sync must keep an activity of an owned account only when a desired activity agrees with it in every field, comment included, import every desired activity nothing agrees with, delete every other activity of the owned accounts, deleting before the imports only those an import would skip as duplicates, empty with one account-filtered bulk delete an owned account that should hold nothing, and leave nothing to change for a second sync
    role: normative
    status: modeled
    span:
      start: 0
      end: 450
origin:
  kind: agent
  recorded_at: '2026-10-29T00:54:58.897Z'
id: REQ-views-ghostfolio-reconcile
type: req
---
Each Ghostfolio sync must keep an activity of an owned account only when a desired activity agrees with it in every field, comment included, import every desired activity nothing agrees with, delete every other activity of the owned accounts, deleting before the imports only those an import would skip as duplicates, empty with one account-filtered bulk delete an owned account that should hold nothing, and leave nothing to change for a second sync.

## Context

D-4 has the view delete stale activities, import missing ones and revert manual edits; D-64 finds edits by comparing fields and makes every step safe to repeat. Ghostfolio's import skips a duplicate regardless of its account, so an activity moved between owned accounts must go first or it would never be imported; that exception to import-first is this view's reading of D-64. Refines REQ-AUTO-54773934AD125DF8, REQ-AUTO-4F8D1E2BADB7AED0 and REQ-views-ghostfolio-import-before-delete. Modeled by Claude Code for the Ghostfolio view task; the author has not reviewed it yet.

## Source

> The Ghostfolio plugin runs a reconcile loop. It owns only the Ghostfolio accounts it created, puts OpenFief's transaction ID in each activity's comment, deletes stale activities and imports missing ones. Manual edits inside its accounts are reverted on the next sync.

docs/DESIGN.md D-4

> How the Ghostfolio plugin syncs, given how Ghostfolio behaves:
> - It creates its Ghostfolio accounts with UUIDs it generates and stores; that's how it knows which accounts it owns.
> - Each sync imports missing activities first, then deletes stale ones, because Ghostfolio deletes an asset and its price history along with its last activity. A bulk delete always carries an account filter; an unfiltered one deletes every activity the user has.
> - A bond is a `MANUAL` asset whose symbol is a UUID derived from its series code, with the code as its name. Every sync sends its full price history.
> - Edits are found by comparing fields, since a manually edited activity keeps its comment.
> - Imports aren't atomic, so every step is safe to repeat, and each import is first checked with `dryRun=true`.
> - It refuses to sync while Ghostfolio's "restricted view" setting is on: that setting blanks comments, so every activity would look stale.
> - Ghostfolio renames these internal endpoints within `/v1`, so the plugin supports a pinned range of Ghostfolio versions, checked by contract tests on recorded fixtures.

docs/DESIGN.md D-64
