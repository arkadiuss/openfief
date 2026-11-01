---
title: Balance assertions must not count as transactions in the check that allows a broker account's remap.
status: open
text_ref: design-decisions:D-156#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-156-2
  - lane:strict
semantic_text: Balance assertions must not count as transactions in the check that allows a broker account's remap.
logic_claims:
  - CLAIM-FB44FC63E183D8AE
semantic_clauses:
  - Balance assertions must not count as transactions in the check that allows a broker account's remap.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: fc09068302ea878369d2b75be3b6b3166476fa05f1d9519933b8a3e08d6af894
semantic_inventory:
  - claim_key: CLAIM-FB44FC63E183D8AE
    claim_text: Balance assertions must not count as transactions in the check that allows a broker account's remap
    role: normative
    status: modeled
    span:
      start: 0
      end: 99
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-156#2
  recorded_at: '2026-11-01T12:30:33.618Z'
id: REQ-ingest-remap-check-ignores-assertions
type: req
---
Balance assertions must not count as transactions in the check that allows a broker account's remap.

## Context

The author recorded it on 2026-11-01, settling the last implementation questions (#33, #38, #39). Balance assertions are D-18's broker position snapshots; they are checked against the ledger, not booked into it. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> A broker account counts as having no transactions in its OpenFief account (D-120) while that account holds no version of a transaction from the connector instance that reports it; balance assertions don't count. The ledger doesn't record which broker account a transaction came from, so when one instance feeds several broker accounts into the same account the check refuses some remaps it could allow, and never allows an unsafe one.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-156#2
