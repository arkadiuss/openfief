---
title: Every import the Ghostfolio view sends must first be sent with dryRun=true, in batches of at most 500 activities carrying the assets they use, and must not be sent when the preview shows an activity skipped as a duplicate or stored under another account or symbol than sent.
status: open
text_ref: docs/DESIGN.md#D-64
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-64
  - lane:strict
semantic_text: Every import the Ghostfolio view sends must first be sent with dryRun=true, in batches of at most 500 activities carrying the assets they use, and must not be sent when the preview shows an activity skipped as a duplicate or stored under another account or symbol than sent.
logic_claims:
  - CLAIM-6A16BB50F02CE3CA
semantic_clauses:
  - Every import the Ghostfolio view sends must first be sent with dryRun=true, in batches of at most 500 activities carrying the assets they use, and must not be sent when the preview shows an activity skipped as a duplicate or stored under another account or symbol than sent.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 5462569f0fece7e41f8b363730d0fcc463158e865181e213bc7a10b05e8ee828
semantic_inventory:
  - claim_key: CLAIM-6A16BB50F02CE3CA
    claim_text: Every import the Ghostfolio view sends must first be sent with dryRun=true, in batches of at most 500 activities carrying the assets they use, and must not be sent when the preview shows an activity skipped as a duplicate or stored under another account or symbol than sent
    role: normative
    status: modeled
    span:
      start: 0
      end: 273
origin:
  kind: agent
  recorded_at: '2026-10-29T00:55:03.802Z'
id: REQ-views-ghostfolio-dry-run
type: req
---
Every import the Ghostfolio view sends must first be sent with dryRun=true, in batches of at most 500 activities carrying the assets they use, and must not be sent when the preview shows an activity skipped as a duplicate or stored under another account or symbol than sent.

## Context

D-64: imports aren't atomic, so each is checked with dryRun=true first. Ghostfolio maps a MANUAL asset to the user's own asset of the same name and skips duplicates silently, and either would make the next sync see a difference forever, so the preview must show each activity stored exactly as sent. Batches keep each body well under Ghostfolio's 10 MB limit (R-15). Modeled by Claude Code for the Ghostfolio view task; the author has not reviewed it yet.

## Source

> How the Ghostfolio plugin syncs, given how Ghostfolio behaves:
> - It creates its Ghostfolio accounts with UUIDs it generates and stores; that's how it knows which accounts it owns.
> - Each sync imports missing activities first, then deletes stale ones, because Ghostfolio deletes an asset and its price history along with its last activity. A bulk delete always carries an account filter; an unfiltered one deletes every activity the user has.
> - A bond is a `MANUAL` asset whose symbol is a UUID derived from its series code, with the code as its name. Every sync sends its full price history.
> - Edits are found by comparing fields, since a manually edited activity keeps its comment.
> - Imports aren't atomic, so every step is safe to repeat, and each import is first checked with `dryRun=true`.
> - It refuses to sync while Ghostfolio's "restricted view" setting is on: that setting blanks comments, so every activity would look stale.
> - Ghostfolio renames these internal endpoints within `/v1`, so the plugin supports a pinned range of Ghostfolio versions, checked by contract tests on recorded fixtures.

docs/DESIGN.md D-64
