---
title: 'The Ghostfolio view must make each treasury bond a MANUAL asset whose symbol is a UUID derived from its series code and whose name is the code, and must send its full price history on every sync: each stored valuation divided by the quantity held at the end of that day.'
status: open
text_ref: docs/DESIGN.md#D-64
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-64
  - lane:strict
semantic_text: 'The Ghostfolio view must make each treasury bond a MANUAL asset whose symbol is a UUID derived from its series code and whose name is the code, and must send its full price history on every sync: each stored valuation divided by the quantity held at the end of that day.'
logic_claims:
  - CLAIM-DCB52589DDA330C5
semantic_clauses:
  - 'The Ghostfolio view must make each treasury bond a MANUAL asset whose symbol is a UUID derived from its series code and whose name is the code, and must send its full price history on every sync: each stored valuation divided by the quantity held at the end of that day.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 7618ee536928256e38155a370cc315664ae0cd0ac4b643ff06456b6c0046ce62
semantic_inventory:
  - claim_key: CLAIM-DCB52589DDA330C5
    claim_text: 'The Ghostfolio view must make each treasury bond a MANUAL asset whose symbol is a UUID derived from its series code and whose name is the code, and must send its full price history on every sync: each stored valuation divided by the quantity held at the end of that day'
    role: normative
    status: modeled
    span:
      start: 0
      end: 269
origin:
  kind: agent
  recorded_at: '2026-10-29T00:54:53.513Z'
id: REQ-views-ghostfolio-bond-assets
type: req
---
The Ghostfolio view must make each treasury bond a MANUAL asset whose symbol is a UUID derived from its series code and whose name is the code, and must send its full price history on every sync: each stored valuation divided by the quantity held at the end of that day.

## Context

D-64 decides the asset's symbol and name and that every sync sends the full price history, so prices lost with a deleted asset come back. The read API's stored valuations (D-36) are per holding, so the price per unit is the value divided by the quantity the transactions give for that day. Modeled by Claude Code for the Ghostfolio view task; the author has not reviewed it yet.

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
