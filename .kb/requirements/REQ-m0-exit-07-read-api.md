---
title: The read API's /v1 must serve accounts, transactions, tax lots, holdings and valuations as its OpenAPI spec describes, reject unauthenticated calls and have no write endpoint, and the Connect clients generated from the same definitions must return the same data.
status: open
text_ref: design-decisions:D-200#7
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-200-7
  - lane:strict
semantic_text: The read API's /v1 must serve accounts, transactions, tax lots, holdings and valuations as its OpenAPI spec describes, reject unauthenticated calls and have no write endpoint, and the Connect clients generated from the same definitions must return the same data.
logic_claims:
  - CLAIM-0219FB67F9C0322A
semantic_clauses:
  - The read API's /v1 must serve accounts, transactions, tax lots, holdings and valuations as its OpenAPI spec describes, reject unauthenticated calls and have no write endpoint, and the Connect clients generated from the same definitions must return the same data.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: a60517958fdb13af84e93945c168cc37854e9d967a5d3ea5192283eb3c764a7c
semantic_inventory:
  - claim_key: CLAIM-0219FB67F9C0322A
    claim_text: The read API's /v1 must serve accounts, transactions, tax lots, holdings and valuations as its OpenAPI spec describes, reject unauthenticated calls and have no write endpoint, and the Connect clients generated from the same definitions must return the same data
    role: normative
    status: modeled
    span:
      start: 0
      end: 261
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-200#7
  recorded_at: '2026-11-07T03:31:55.080Z'
id: REQ-m0-exit-07-read-api
type: req
---
The read API's /v1 must serve accounts, transactions, tax lots, holdings and valuations as its OpenAPI spec describes, reject unauthenticated calls and have no write endpoint, and the Connect clients generated from the same definitions must return the same data.

## Context

Author decision D-200: M0 is done when every one of its nineteen criteria holds. Each criterion is an end-to-end scenario on the firm requirement it checks, and the whole set passes on both kind and k3d (criterion 15); the scenarios that verify this criterion carry the tag acceptance:p10-7.

## Source

> **Read API.** `/v1` serves accounts, transactions, tax lots, holdings and valuations as its OpenAPI spec describes, rejects unauthenticated calls, and has no write endpoint. The Connect clients generated from the same definitions return the same data.

design-decisions:D-200#7 (docs/DESIGN.md, D-200, criterion 7)
