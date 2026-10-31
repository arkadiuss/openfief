---
title: The core must hand each bond lot to the treasury-bonds connector with the account the owner entered it on, the connector must book it there, and the plugin's config must name no account.
status: open
text_ref: docs/DESIGN.md#D-118
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-118
  - lane:strict
semantic_text: The core must hand each bond lot to the treasury-bonds connector with the account the owner entered it on, the connector must book it there, and the plugin's config must name no account.
logic_claims:
  - CLAIM-0C67A9ED96FA810F
semantic_clauses:
  - The core must hand each bond lot to the treasury-bonds connector with the account the owner entered it on, the connector must book it there, and the plugin's config must name no account.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 5840487d85710ab4eceb67f3f509dff0b800bd5c483fb55bed4d8884236422f3
semantic_inventory:
  - claim_key: CLAIM-0C67A9ED96FA810F
    claim_text: The core must hand each bond lot to the treasury-bonds connector with the account the owner entered it on, the connector must book it there, and the plugin's config must name no account
    role: normative
    status: modeled
    span:
      start: 0
      end: 185
origin:
  kind: agent
  recorded_at: '2026-10-31T13:25:50.243Z'
id: REQ-plugins-treasury-bonds-serves-every-account
type: req
---
The core must hand each bond lot to the treasury-bonds connector with the account the owner entered it on, the connector must book it there, and the plugin's config must name no account.

## Context

D-118: one treasury-bonds instance serves every bond account bound to it. Before it, the plugin's config named the one account its lots went into. Recorded by the implementation agent of impl/decisions-117; default firmness, not reviewed by the author yet.

## Source

> The core hands each bond lot to the connector together with the account the owner entered it on, so one treasury-bonds instance serves every bond account and its config names none.

docs/DESIGN.md#D-118
