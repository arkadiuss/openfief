---
title: When one connector instance feeds several broker accounts into the same OpenFief account, the remap check must refuse a remap it can't prove safe rather than allow an unsafe one.
status: open
text_ref: design-decisions:D-156#3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-156-3
  - lane:strict
semantic_text: When one connector instance feeds several broker accounts into the same OpenFief account, the remap check must refuse a remap it can't prove safe rather than allow an unsafe one.
logic_claims:
  - CLAIM-369F6097891E1F53
semantic_clauses:
  - When one connector instance feeds several broker accounts into the same OpenFief account, the remap check must refuse a remap it can't prove safe rather than allow an unsafe one.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: a693feb0e39f74adb077669c92609f209a76637bf87c7bbcec82a6a9ab4c42ee
semantic_inventory:
  - claim_key: CLAIM-369F6097891E1F53
    claim_text: When one connector instance feeds several broker accounts into the same OpenFief account, the remap check must refuse a remap it can't prove safe rather than allow an unsafe one
    role: condition
    status: modeled
    span:
      start: 0
      end: 177
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-156#3
  recorded_at: '2026-11-01T12:30:41.665Z'
id: REQ-ingest-remap-check-never-unsafe
type: req
---
When one connector instance feeds several broker accounts into the same OpenFief account, the remap check must refuse a remap it can't prove safe rather than allow an unsafe one.

## Context

The author recorded it on 2026-11-01, settling the last implementation questions (#33, #38, #39). The decision gives the reason: the ledger doesn't record which broker account a transaction came from, so the check refuses some remaps it could allow. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> A broker account counts as having no transactions in its OpenFief account (D-120) while that account holds no version of a transaction from the connector instance that reports it; balance assertions don't count. The ledger doesn't record which broker account a transaction came from, so when one instance feeds several broker accounts into the same account the check refuses some remaps it could allow, and never allows an unsafe one.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-156#3
