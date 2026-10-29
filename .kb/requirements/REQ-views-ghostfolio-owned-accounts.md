---
title: The Ghostfolio view must own exactly one Ghostfolio account per account of the record, identified by a UUID derived from the view instance's name and the account's ID, create it with that UUID when it is missing, and never change or delete an account or activity outside its owned accounts.
status: open
text_ref: docs/DESIGN.md#D-64
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-64
  - lane:strict
semantic_text: The Ghostfolio view must own exactly one Ghostfolio account per account of the record, identified by a UUID derived from the view instance's name and the account's ID, create it with that UUID when it is missing, and never change or delete an account or activity outside its owned accounts.
logic_claims:
  - CLAIM-10C075DD3C338EF8
semantic_clauses:
  - The Ghostfolio view must own exactly one Ghostfolio account per account of the record, identified by a UUID derived from the view instance's name and the account's ID, create it with that UUID when it is missing, and never change or delete an account or activity outside its owned accounts.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: e13260f52bc6f2c7c5ad2037b9e62e8b8db31e2a376f68449e06f78fb83841fa
semantic_inventory:
  - claim_key: CLAIM-10C075DD3C338EF8
    claim_text: The Ghostfolio view must own exactly one Ghostfolio account per account of the record, identified by a UUID derived from the view instance's name and the account's ID, create it with that UUID when it is missing, and never change or delete an account or activity outside its owned accounts
    role: normative
    status: modeled
    span:
      start: 0
      end: 289
origin:
  kind: agent
  recorded_at: '2026-10-29T00:54:43.174Z'
id: REQ-views-ghostfolio-owned-accounts
type: req
---
The Ghostfolio view must own exactly one Ghostfolio account per account of the record, identified by a UUID derived from the view instance's name and the account's ID, create it with that UUID when it is missing, and never change or delete an account or activity outside its owned accounts.

## Context

D-4 limits the view to the accounts it created and D-64 has it know them by UUIDs it generates and stores. A plugin has a read-only file system, no volume and no Kubernetes API, so the UUIDs are derived (UUID version 5 in a fixed namespace) instead of stored; deriving also makes creating an account safe to repeat. Including the instance name keeps two installed instances (D-48) apart. Refines REQ-AUTO-54773934AD125DF8. Modeled by Claude Code for the Ghostfolio view task; the author has not reviewed it yet.

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
