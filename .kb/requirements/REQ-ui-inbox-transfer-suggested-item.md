---
title: The Inbox must show a suggested transfer's attention item among the decisions with its legs, settled in place by the owner's Confirm or Not a transfer, and count it in the navigation's open count.
status: open
text_ref: docs/DESIGN.md#D-196
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-196
  - lane:strict
semantic_text: The Inbox must show a suggested transfer's attention item among the decisions with its legs, settled in place by the owner's Confirm or Not a transfer, and count it in the navigation's open count.
logic_claims:
  - CLAIM-85E393853C473191
semantic_clauses:
  - The Inbox must show a suggested transfer's attention item among the decisions with its legs, settled in place by the owner's Confirm or Not a transfer, and count it in the navigation's open count.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 17b81466576f4449d6fb4bb4bd633a881dd553a2cd567d845b50f038aa62a445
semantic_inventory:
  - claim_key: CLAIM-85E393853C473191
    claim_text: The Inbox must show a suggested transfer's attention item among the decisions with its legs, settled in place by the owner's Confirm or Not a transfer, and count it in the navigation's open count
    role: normative
    status: modeled
    span:
      start: 0
      end: 195
origin:
  kind: agent
  recorded_at: '2026-11-06T06:15:11.464Z'
id: REQ-ui-inbox-transfer-suggested-item
type: req
---
The Inbox must show a suggested transfer's attention item among the decisions with its legs, settled in place by the owner's Confirm or Not a transfer, and count it in the navigation's open count.

## Context

D-196 settles a transfer to confirm in place; with suggested transfers as attention items (D-130) the item itself carries the decision forms, and no separate list of suggestions remains. Modeled by Claude Code for the UI v2 overview slice; the author has not reviewed the wording.

## Source

> **[D-105]** Transfers between the owner's own accounts are identified by suggestion and confirmation. ... The owner confirms or rejects each suggestion in the UI, and a confirmed link is stored in the ledger, append-only.

docs/DESIGN.md#D-105

> **[D-130]** Every check that needs the owner raises an attention item in one append-only store ... and the UI lists them in one place.

docs/DESIGN.md#D-130
