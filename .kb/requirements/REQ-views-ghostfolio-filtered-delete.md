---
title: The Ghostfolio plugin must never send a bulk delete without an account filter.
status: open
text_ref: design-decisions:D-64#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-64-2
  - lane:strict
semantic_text: The Ghostfolio plugin must never send a bulk delete without an account filter.
logic_claims:
  - CLAIM-6883370B156E4528
semantic_clauses:
  - The Ghostfolio plugin must never send a bulk delete without an account filter.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 9c7e6b0127ef837fa93e58da840ba5016b1aba824fea475b693d1329bc6f0d34
semantic_inventory:
  - claim_key: CLAIM-6883370B156E4528
    claim_text: The Ghostfolio plugin must never send a bulk delete without an account filter
    role: normative
    status: modeled
    span:
      start: 0
      end: 77
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-64#2
  recorded_at: '2026-10-28T06:18:55.495Z'
id: REQ-views-ghostfolio-filtered-delete
type: req
---
The Ghostfolio plugin must never send a bulk delete without an account filter.

## Context

Existing trackers can't hold what OpenFief has to get exactly right: FIFO tax lots priced in PLN, statutory bond values, and sleeve attribution that the later lease ledger reconciles against. So OpenFief owns the record, and trackers only display it.

## Source

> How the Ghostfolio plugin syncs, given how Ghostfolio behaves:
>   - It creates its Ghostfolio accounts with UUIDs it generates and stores; that's how it knows which accounts it owns.
>   - Each sync imports missing activities first, then deletes stale ones, because Ghostfolio deletes an asset and its price history along with its last activity. A bulk delete always carries an account filter; an unfiltered one deletes every activity the user has.
>   - A bond is a `MANUAL` asset whose symbol is a UUID derived from its series code, with the code as its name. Every sync sends its full price history.
>   - Edits are found by comparing fields, since a manually edited activity keeps its comment.
>   - Imports aren't atomic, so every step is safe to repeat, and each import is first checked with `dryRun=true`.
>   - It refuses to sync while Ghostfolio's "restricted view" setting is on: that setting blanks comments, so every activity would look stale.
>   - Ghostfolio renames these internal endpoints within `/v1`, so the plugin supports a pinned range of Ghostfolio versions, checked by contract tests on recorded fixtures.

OpenFief author decisions (docs/DESIGN.md, section 3.1), design-decisions:D-64#2
