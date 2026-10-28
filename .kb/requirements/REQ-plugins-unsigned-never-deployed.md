---
title: An image without a valid signature from a trusted signer must never be deployed, with no override.
status: open
text_ref: design-decisions:D-57#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-57-1
  - lane:strict
semantic_text: An image without a valid signature from a trusted signer must never be deployed, with no override.
logic_claims:
  - CLAIM-7BAFDE8517FB6D01
semantic_clauses:
  - An image without a valid signature from a trusted signer must never be deployed, with no override.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 9207bd411f3c712a72beb88338fc7dfabaff6529b58d94590195178a6ff962f6
semantic_inventory:
  - claim_key: CLAIM-7BAFDE8517FB6D01
    claim_text: An image without a valid signature from a trusted signer must never be deployed, with no override
    role: normative
    status: modeled
    span:
      start: 0
      end: 97
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-57#1
  recorded_at: '2026-10-28T06:09:31.059Z'
id: REQ-plugins-unsigned-never-deployed
type: req
---
An image without a valid signature from a trusted signer must never be deployed, with no override.

## Context

Few plugins are expected. Anyone can write and run their own; the ones the author has reviewed are listed as verified. Every image is signed, so the only question is whose signatures count.

## Source

> An image without a valid signature from a trusted signer is never deployed, and there is no override. A developer testing a plugin locally signs it with their own key, added as a trusted signer (D-45).

OpenFief author decisions (docs/DESIGN.md, section 3.6), design-decisions:D-57#1
