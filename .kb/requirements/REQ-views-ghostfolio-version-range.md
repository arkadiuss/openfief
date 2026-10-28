---
title: The Ghostfolio plugin must support a pinned range of Ghostfolio versions, checked by contract tests on recorded fixtures.
status: open
text_ref: design-decisions:D-64#4
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-64-4
  - lane:strict
semantic_text: The Ghostfolio plugin must support a pinned range of Ghostfolio versions, checked by contract tests on recorded fixtures.
logic_claims:
  - CLAIM-37C406257962CA25
semantic_clauses:
  - The Ghostfolio plugin must support a pinned range of Ghostfolio versions, checked by contract tests on recorded fixtures.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: aa820bf4205c6886d1a6b9c2a41e3ab17258a14246d0f9c6ff4b0a69dbf91011
semantic_inventory:
  - claim_key: CLAIM-37C406257962CA25
    claim_text: The Ghostfolio plugin must support a pinned range of Ghostfolio versions, checked by contract tests on recorded fixtures
    role: normative
    status: modeled
    span:
      start: 0
      end: 120
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-64#4
  recorded_at: '2026-10-28T06:19:01.462Z'
id: REQ-views-ghostfolio-version-range
type: req
---
The Ghostfolio plugin must support a pinned range of Ghostfolio versions, checked by contract tests on recorded fixtures.

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

OpenFief author decisions (docs/DESIGN.md, section 3.1), design-decisions:D-64#4
