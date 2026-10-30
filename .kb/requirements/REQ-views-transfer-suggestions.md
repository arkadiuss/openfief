---
title: The read API must list as suggested transfers each withdrawal that is not in a confirmed transfer, paired with the nearest deposit in date into another account of exactly the same amount and currency at most five calendar days apart that is not in a confirmed transfer, not already suggested and not rejected with it, together with the confirmed and rejected pairs.
status: open
text_ref: docs/DESIGN.md#D-105
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-105
  - lane:strict
semantic_text: The read API must list as suggested transfers each withdrawal that is not in a confirmed transfer, paired with the nearest deposit in date into another account of exactly the same amount and currency at most five calendar days apart that is not in a confirmed transfer, not already suggested and not rejected with it, together with the confirmed and rejected pairs.
logic_claims:
  - CLAIM-7070D9D4423D1BE3
semantic_clauses:
  - The read API must list as suggested transfers each withdrawal that is not in a confirmed transfer, paired with the nearest deposit in date into another account of exactly the same amount and currency at most five calendar days apart that is not in a confirmed transfer, not already suggested and not rejected with it, together with the confirmed and rejected pairs.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: a948ecb1da5065ad42a23d28ab7b9a85c0452d54c3650da5303644078e40e70d
semantic_inventory:
  - claim_key: CLAIM-7070D9D4423D1BE3
    claim_text: The read API must list as suggested transfers each withdrawal that is not in a confirmed transfer, paired with the nearest deposit in date into another account of exactly the same amount and currency at most five calendar days apart that is not in a confirmed transfer, not already suggested and not rejected with it, together with the confirmed and rejected pairs
    role: normative
    status: modeled
    span:
      start: 0
      end: 364
origin:
  kind: agent
  recorded_at: '2026-10-30T06:20:30.354Z'
id: REQ-views-transfer-suggestions
type: req
---
The read API must list as suggested transfers each withdrawal that is not in a confirmed transfer, paired with the nearest deposit in date into another account of exactly the same amount and currency at most five calendar days apart that is not in a confirmed transfer, not already suggested and not rejected with it, together with the confirmed and rejected pairs.

## Context

D-105 has OpenFief suggest each transfer and the owner decide in the UI; the UI reads the suggestions from the read API, which stays read-only (D-6). Refines the firm REQ-ledger-transfer-suggestion.

## Source

> Transfers between the owner's own accounts are identified by suggestion and confirmation. A withdrawal is suggested as a transfer to a deposit into another of the owner's accounts of the same amount and currency at most five calendar days apart, the nearest unmatched deposit first. The owner confirms or rejects each suggestion in the UI, and a confirmed link is stored in the ledger, append-only. Only confirmed transfers are left out of the time-weighted return's cash flows.

docs/DESIGN.md#D-105
