---
title: A deletion must be a new version of its transaction's key with no postings that supersedes the current version, keeps every earlier version, leaves the transaction out of current transactions and everything derived from them, and must be refused for a key that has no version.
status: open
text_ref: docs/DESIGN.md#D-100
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-100
  - lane:strict
semantic_text: A deletion must be a new version of its transaction's key with no postings that supersedes the current version, keeps every earlier version, leaves the transaction out of current transactions and everything derived from them, and must be refused for a key that has no version.
logic_claims:
  - CLAIM-C465E8F1865B2A99
semantic_clauses:
  - A deletion must be a new version of its transaction's key with no postings that supersedes the current version, keeps every earlier version, leaves the transaction out of current transactions and everything derived from them, and must be refused for a key that has no version.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: b1e4e32dd92cd15c6e1303275ce3caa8e2c3fe6b8c8f73adbbce5b7cc8109e29
semantic_inventory:
  - claim_key: CLAIM-C465E8F1865B2A99
    claim_text: A deletion must be a new version of its transaction's key with no postings that supersedes the current version, keeps every earlier version, leaves the transaction out of current transactions and everything derived from them, and must be refused for a key that has no version
    role: normative
    status: modeled
    span:
      start: 0
      end: 275
origin:
  kind: agent
  recorded_at: '2026-10-30T05:01:53.980Z'
id: REQ-ledger-deletion-supersedes
type: req
---
A deletion must be a new version of its transaction's key with no postings that supersedes the current version, keeps every earlier version, leaves the transaction out of current transactions and everything derived from them, and must be refused for a key that has no version.

## Context

D-100 lets the owner delete a bond lot as a correction that supersedes it and erases nothing; D-18 keeps superseded versions. A deletion as a ledger kind with no postings, and refusing to delete what was never recorded, are this package's narrow reading. Recorded by the bonds implementation agent (impl/bonds-2).

## Source

> The owner can delete a lot as a correction, which supersedes it (D-18) and erases nothing.

docs/DESIGN.md#D-100
