---
title: The Ghostfolio client must log in by exchanging the security token, read from its Secret on every login, for a bearer token; must refuse to list or bulk-delete activities without an account filter, sending nothing; must count deleting an activity Ghostfolio no longer has as done; must send numbers as JSON numbers with their exact digits and dates as midnight UTC; and must report an error status by call and status only.
status: open
text_ref: docs/DESIGN.md#D-64
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-64
  - lane:strict
semantic_text: The Ghostfolio client must log in by exchanging the security token, read from its Secret on every login, for a bearer token; must refuse to list or bulk-delete activities without an account filter, sending nothing; must count deleting an activity Ghostfolio no longer has as done; must send numbers as JSON numbers with their exact digits and dates as midnight UTC; and must report an error status by call and status only.
logic_claims:
  - CLAIM-698921AEF5C0BEB1
semantic_clauses:
  - The Ghostfolio client must log in by exchanging the security token, read from its Secret on every login, for a bearer token; must refuse to list or bulk-delete activities without an account filter, sending nothing; must count deleting an activity Ghostfolio no longer has as done; must send numbers as JSON numbers with their exact digits and dates as midnight UTC; and must report an error status by call and status only.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: d0cc0a3c95f7d1a824141cb8563142fe32ed15d78c1f5dd9daa164081acb03cc
semantic_inventory:
  - claim_key: CLAIM-698921AEF5C0BEB1
    claim_text: The Ghostfolio client must log in by exchanging the security token, read from its Secret on every login, for a bearer token; must refuse to list or bulk-delete activities without an account filter, sending nothing; must count deleting an activity Ghostfolio no longer has as done; must send numbers as JSON numbers with their exact digits and dates as midnight UTC; and must report an error status by call and status only
    role: normative
    status: modeled
    span:
      start: 0
      end: 421
origin:
  kind: agent
  recorded_at: '2026-10-29T00:55:14.090Z'
id: REQ-views-ghostfolio-client
type: req
---
The Ghostfolio client must log in by exchanging the security token, read from its Secret on every login, for a bearer token; must refuse to list or bulk-delete activities without an account filter, sending nothing; must count deleting an activity Ghostfolio no longer has as done; must send numbers as JSON numbers with their exact digits and dates as midnight UTC; and must report an error status by call and status only.

## Context

D-64: an unfiltered bulk delete deletes every activity the user has, so the client makes one impossible. A 404 on a repeated delete keeps steps safe to repeat. Ghostfolio's anonymous login exchanges the security token for a bearer token (R-15). Error bodies can quote the data sent, which must not reach logs (GR-12). Refines REQ-views-ghostfolio-filtered-delete. Modeled by Claude Code for the Ghostfolio view task; the author has not reviewed it yet.

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
