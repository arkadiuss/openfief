---
title: A decision on a transfer pair that differs from its current one must be appended as the pair's next version, superseding the earlier ones, and repeating the current decision must append nothing.
status: open
text_ref: docs/DESIGN.md#D-105
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-105
  - lane:strict
semantic_text: A decision on a transfer pair that differs from its current one must be appended as the pair's next version, superseding the earlier ones, and repeating the current decision must append nothing.
logic_claims:
  - CLAIM-9C5D6D1E56CCAE07
semantic_clauses:
  - A decision on a transfer pair that differs from its current one must be appended as the pair's next version, superseding the earlier ones, and repeating the current decision must append nothing.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 7895b87d0841128b1bf5e2397a93ae76fff0fdd9635daa21b3b97121b384b3ec
semantic_inventory:
  - claim_key: CLAIM-9C5D6D1E56CCAE07
    claim_text: A decision on a transfer pair that differs from its current one must be appended as the pair's next version, superseding the earlier ones, and repeating the current decision must append nothing
    role: normative
    status: modeled
    span:
      start: 0
      end: 193
origin:
  kind: agent
  recorded_at: '2026-10-30T06:06:18.442Z'
id: REQ-ledger-transfer-correction
type: req
---
A decision on a transfer pair that differs from its current one must be appended as the pair's next version, superseding the earlier ones, and repeating the current decision must append nothing.

## Context

D-105 stores a confirmed link append-only, and D-18 corrects by superseding, never by changing what was stored. A rejection is stored the same way, so a rejected pair is not suggested again and the owner can still change their mind. Refines REQ-ledger-transfer-link-append-only.

## Source

> Transfers between the owner's own accounts are identified by suggestion and confirmation. A withdrawal is suggested as a transfer to a deposit into another of the owner's accounts of the same amount and currency at most five calendar days apart, the nearest unmatched deposit first. The owner confirms or rejects each suggestion in the UI, and a confirmed link is stored in the ledger, append-only. Only confirmed transfers are left out of the time-weighted return's cash flows.

docs/DESIGN.md#D-105
