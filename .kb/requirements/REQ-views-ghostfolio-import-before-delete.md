---
title: Each Ghostfolio sync must import missing activities before deleting stale ones.
status: closed
text_ref: design-decisions:D-64#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-64-1
  - lane:strict
semantic_text: Each Ghostfolio sync must import missing activities before deleting stale ones.
logic_claims:
  - CLAIM-D75A4E2631099809
semantic_clauses:
  - Each Ghostfolio sync must import missing activities before deleting stale ones.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 52e463955fd48155a2a47c392ce486d73a9cb8d00a94d0e7e82b4efe5d801d54
semantic_inventory:
  - claim_key: CLAIM-D75A4E2631099809
    claim_text: Each Ghostfolio sync must import missing activities before deleting stale ones
    role: normative
    status: modeled
    span:
      start: 0
      end: 78
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-64#1
  recorded_at: '2026-10-28T06:18:52.487Z'
id: REQ-views-ghostfolio-import-before-delete
type: req
---
Each Ghostfolio sync must import missing activities before deleting stale ones.

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

OpenFief author decisions (docs/DESIGN.md, section 3.1), design-decisions:D-64#1
