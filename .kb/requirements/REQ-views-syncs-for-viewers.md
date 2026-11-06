---
title: The read API must list a connector instance's sync status, attempts, failed attempts and imports that booked into no account only to a caller who may view an account the instance is bound to, or, for an instance bound to none, to the owner and today's other holders of read access, never a viewer.
status: open
text_ref: docs/DESIGN.md#D-199
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-199
  - lane:strict
semantic_text: The read API must list a connector instance's sync status, attempts, failed attempts and imports that booked into no account only to a caller who may view an account the instance is bound to, or, for an instance bound to none, to the owner and today's other holders of read access, never a viewer.
logic_claims:
  - CLAIM-E75FEAB02F01B017
semantic_clauses:
  - The read API must list a connector instance's sync status, attempts, failed attempts and imports that booked into no account only to a caller who may view an account the instance is bound to, or, for an instance bound to none, to the owner and today's other holders of read access, never a viewer.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 286d9a6995fac5291e5c229397d8bacc89cba1654ee9291857c04468fbe7a871
semantic_inventory:
  - claim_key: CLAIM-E75FEAB02F01B017
    claim_text: The read API must list a connector instance's sync status, attempts, failed attempts and imports that booked into no account only to a caller who may view an account the instance is bound to, or, for an instance bound to none, to the owner and today's other holders of read access, never a viewer
    role: normative
    status: modeled
    span:
      start: 0
      end: 296
origin:
  kind: agent
  recorded_at: '2026-11-06T06:15:15.714Z'
id: REQ-views-syncs-for-viewers
type: req
---
The read API must list a connector instance's sync status, attempts, failed attempts and imports that booked into no account only to a caller who may view an account the instance is bound to, or, for an instance bound to none, to the owner and today's other holders of read access, never a viewer.

## Context

Refines D-199: an instance's syncs belong to the accounts its connector bindings name, and any one of them suffices; script tokens, view plugins and demo sessions count with the owner for instance-only records until viewers exist (D-65, D-165). Modeled by Claude Code for the UI v2 overview slice; the author has not reviewed the wording.

## Source

> **[D-199]** Once viewers exist (D-65), a viewer sees the syncs, sync attempts and import activity of the accounts they may view; anything tied to a connector instance and no account stays the owner's, as D-165 has for attention items.

docs/DESIGN.md#D-199

> **[D-165]** An attention item about something tied to no account, such as an unmapped broker account, names only its connector instance on the read API. Once viewers exist (D-65), only the owner sees such items.

docs/DESIGN.md#D-165
