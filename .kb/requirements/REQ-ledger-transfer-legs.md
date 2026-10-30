---
title: A transfer decision must name a stored withdrawal and a stored deposit whose current versions are in different accounts, of exactly the same amount and currency, at most five calendar days apart, and neither opening cash.
status: open
text_ref: docs/DESIGN.md#D-105
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-105
  - lane:strict
semantic_text: A transfer decision must name a stored withdrawal and a stored deposit whose current versions are in different accounts, of exactly the same amount and currency, at most five calendar days apart, and neither opening cash.
logic_claims:
  - CLAIM-048BED1F7B216729
semantic_clauses:
  - A transfer decision must name a stored withdrawal and a stored deposit whose current versions are in different accounts, of exactly the same amount and currency, at most five calendar days apart, and neither opening cash.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: cd3f7ca0c38ed633bc0599257ce083153377e23ccb1f62c58fe281f9f17da8e0
semantic_inventory:
  - claim_key: CLAIM-048BED1F7B216729
    claim_text: A transfer decision must name a stored withdrawal and a stored deposit whose current versions are in different accounts, of exactly the same amount and currency, at most five calendar days apart, and neither opening cash
    role: normative
    status: modeled
    span:
      start: 0
      end: 220
origin:
  kind: agent
  recorded_at: '2026-10-30T06:06:03.920Z'
id: REQ-ledger-transfer-legs
type: req
---
A transfer decision must name a stored withdrawal and a stored deposit whose current versions are in different accounts, of exactly the same amount and currency, at most five calendar days apart, and neither opening cash.

## Context

D-105 links a withdrawal and a deposit that the read API suggested by this rule. The ledger checks the rule again when the owner decides, against the legs' current versions, so a stale or forged pair is never stored. Refining REQ-ledger-transfer-suggestion for the stored side; the author asked for D-105 after the design session.

## Source

> Transfers between the owner's own accounts are identified by suggestion and confirmation. A withdrawal is suggested as a transfer to a deposit into another of the owner's accounts of the same amount and currency at most five calendar days apart, the nearest unmatched deposit first. The owner confirms or rejects each suggestion in the UI, and a confirmed link is stored in the ledger, append-only. Only confirmed transfers are left out of the time-weighted return's cash flows.

docs/DESIGN.md#D-105
