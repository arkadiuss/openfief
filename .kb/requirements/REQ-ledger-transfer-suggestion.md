---
title: A withdrawal must be suggested as a transfer to a deposit into another of the owner's accounts of the same amount and currency at most five calendar days apart, the nearest unmatched deposit first.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-105-1
  - lane:strict
text_ref: design-decisions:D-105#1
origin:
  kind: import
  ref: design-decisions:D-105#1
  recorded_at: '2026-10-29T10:09:26.975Z'
semantic_text: A withdrawal must be suggested as a transfer to a deposit into another of the owner's accounts of the same amount and currency at most five calendar days apart, the nearest unmatched deposit first.
logic_claims:
  - CLAIM-B3882A0F98494B51
semantic_clauses:
  - A withdrawal must be suggested as a transfer to a deposit into another of the owner's accounts of the same amount and currency at most five calendar days apart, the nearest unmatched deposit first.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: ff495d0520d97321152181b552650b21a9d957f618ec875e26a2073739a75b66
semantic_inventory:
  - claim_key: CLAIM-B3882A0F98494B51
    claim_text: A withdrawal must be suggested as a transfer to a deposit into another of the owner's accounts of the same amount and currency at most five calendar days apart, the nearest unmatched deposit first
    role: normative
    status: modeled
    span:
      start: 0
      end: 196
id: REQ-ledger-transfer-suggestion
type: req
---
A withdrawal must be suggested as a transfer to a deposit into another of the owner's accounts of the same amount and currency at most five calendar days apart, the nearest unmatched deposit first.

## Context

D-61 says transfers between the owner's own accounts are not cash flows of the time-weighted return. D-105 decides how a transfer is identified: OpenFief suggests a match by this rule, and the owner confirms or rejects it (REQ-ui-transfer-confirmation). The author recorded D-105 to D-107 after the design session that answered the implementation's questions.

## Source

> Transfers between the owner's own accounts are identified by suggestion and confirmation. A withdrawal is suggested as a transfer to a deposit into another of the owner's accounts of the same amount and currency at most five calendar days apart, the nearest unmatched deposit first. The owner confirms or rejects each suggestion in the UI, and a confirmed link is stored in the ledger, append-only. Only confirmed transfers are left out of the time-weighted return's cash flows.

OpenFief author decisions (docs/DESIGN.md, section 3.4), design-decisions:D-105#1
