---
title: The core must sync every running connector instance when its sync policy says it is due, at once when it never synced and then every interval plus a random jitter after a success, counting an account seeded by an opening statement upload as seeded, and must not sync an instance that no longer runs while keeping its state.
status: open
text_ref: docs/DESIGN.md#D-109
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-109
  - lane:strict
semantic_text: The core must sync every running connector instance when its sync policy says it is due, at once when it never synced and then every interval plus a random jitter after a success, counting an account seeded by an opening statement upload as seeded, and must not sync an instance that no longer runs while keeping its state.
logic_claims:
  - CLAIM-B4EB1779F051106D
semantic_clauses:
  - The core must sync every running connector instance when its sync policy says it is due, at once when it never synced and then every interval plus a random jitter after a success, counting an account seeded by an opening statement upload as seeded, and must not sync an instance that no longer runs while keeping its state.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 1a53d3f6de6a5b45e6c1a60dad3f3fb40cc636ab8e9da276d7387a40ddda8b9e
semantic_inventory:
  - claim_key: CLAIM-B4EB1779F051106D
    claim_text: The core must sync every running connector instance when its sync policy says it is due, at once when it never synced and then every interval plus a random jitter after a success, counting an account seeded by an opening statement upload as seeded, and must not sync an instance that no longer runs while keeping its state
    role: normative
    status: modeled
    span:
      start: 0
      end: 322
origin:
  kind: agent
  recorded_at: '2026-10-29T17:20:16.524Z'
id: REQ-ingest-scheduler-when-due
type: req
---
The core must sync every running connector instance when its sync policy says it is due, at once when it never synced and then every interval plus a random jitter after a success, counting an account seeded by an opening statement upload as seeded, and must not sync an instance that no longer runs while keeping its state.

## Context

D-49 has the core own the schedules; D-95 picks the mode from the ledger's seeding, and D-109 makes an opening upload seed the account for scheduled syncs too.

## Source

> A statement upload goes to the connector instance bound to the account. A statement that covers several accounts, such as an IBKR Flex statement of linked accounts, books every mapped account bound to that instance and reports the broker accounts that aren't mapped (D-72). The core chooses the mode, opening for an account not yet seeded and transactions after that (D-95), and the owner may override it; an opening upload seeds the account for scheduled syncs too. The owner's UI learns through its owner service (D-79) which accounts have a connector that takes files, and shows the upload form only there. The core accepts files up to 64 MiB.

docs/DESIGN.md#D-109
