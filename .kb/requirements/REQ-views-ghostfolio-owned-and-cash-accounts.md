---
title: The Ghostfolio view must own one Ghostfolio account per account of the record, identified by a UUID derived from the view instance's name and the account's ID, and one per other currency of that account's cash, identified by a UUID derived from the instance's name, the account's ID and the currency, create each with its UUID when it is missing, and never change or delete an account or activity outside its owned accounts.
status: open
text_ref: docs/DESIGN.md#D-133
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-133
  - lane:strict
semantic_text: The Ghostfolio view must own one Ghostfolio account per account of the record, identified by a UUID derived from the view instance's name and the account's ID, and one per other currency of that account's cash, identified by a UUID derived from the instance's name, the account's ID and the currency, create each with its UUID when it is missing, and never change or delete an account or activity outside its owned accounts.
logic_claims:
  - CLAIM-3B70B911B00513F2
semantic_clauses:
  - The Ghostfolio view must own one Ghostfolio account per account of the record, identified by a UUID derived from the view instance's name and the account's ID, and one per other currency of that account's cash, identified by a UUID derived from the instance's name, the account's ID and the currency, create each with its UUID when it is missing, and never change or delete an account or activity outside its owned accounts.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: b2dd091bc01c5430d9d6d6dc9cd688ffdc0dd5e586c92bc331b0946066384b8e
semantic_inventory:
  - claim_key: CLAIM-3B70B911B00513F2
    claim_text: The Ghostfolio view must own one Ghostfolio account per account of the record, identified by a UUID derived from the view instance's name and the account's ID, and one per other currency of that account's cash, identified by a UUID derived from the instance's name, the account's ID and the currency, create each with its UUID when it is missing, and never change or delete an account or activity outside its owned accounts
    role: normative
    status: modeled
    span:
      start: 0
      end: 423
origin:
  kind: agent
  recorded_at: '2026-10-31T19:00:23.863Z'
id: REQ-views-ghostfolio-owned-and-cash-accounts
type: req
---
The Ghostfolio view must own one Ghostfolio account per account of the record, identified by a UUID derived from the view instance's name and the account's ID, and one per other currency of that account's cash, identified by a UUID derived from the instance's name, the account's ID and the currency, create each with its UUID when it is missing, and never change or delete an account or activity outside its owned accounts.

## Context

D-133 gives each currency of an account's cash other than its Ghostfolio account's an owned account of its own, with a UUID from the instance, the account and the currency; D-64 derives the account's own UUID from the instance and the account. Refines REQ-views-ghostfolio-derived-account-uuids (firm, D-64). Modeled by Claude Code while building D-133; the author has not reviewed this wording.

## Why this changed

REQ-views-ghostfolio-owned-accounts owned exactly one Ghostfolio account per account of the record, so cash in any other currency had nowhere to go; D-133 decided that each such currency gets an owned account of its own.

## Source

> Each currency of an account's cash other than its Ghostfolio account's gets an owned Ghostfolio account of its own, holding only that balance, with a UUID derived from the instance, the account and the currency, so Ghostfolio's totals include all of the account's cash.

docs/DESIGN.md#D-133
