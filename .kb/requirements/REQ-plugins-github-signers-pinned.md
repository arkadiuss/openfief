---
title: Every keyless signer identity GitHub issues, a catalog publisher's or one the owner adds, must pin both its source repository and its source ref.
status: open
text_ref: design-decisions:D-187#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-187-1
  - lane:strict
semantic_text: Every keyless signer identity GitHub issues, a catalog publisher's or one the owner adds, must pin both its source repository and its source ref.
logic_claims:
  - CLAIM-A06252989A710E8D
semantic_clauses:
  - Every keyless signer identity GitHub issues, a catalog publisher's or one the owner adds, must pin both its source repository and its source ref.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: daae8d38169eab725d329f04cb8f70d6110e1cd23e717b9b466d8d1cfc1ce266
semantic_inventory:
  - claim_key: CLAIM-A06252989A710E8D
    claim_text: Every keyless signer identity GitHub issues, a catalog publisher's or one the owner adds, must pin both its source repository and its source ref
    role: normative
    status: modeled
    span:
      start: 0
      end: 144
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-187#1
  recorded_at: '2026-11-03T13:51:21.579Z'
id: REQ-plugins-github-signers-pinned
type: req
---
Every keyless signer identity GitHub issues, a catalog publisher's or one the owner adds, must pin both its source repository and its source ref.

## Context

The author decided it on 2026-11-03, in the review's second pass before the alpha tag. D-181 pinned OpenFief's CI to the repository and ref of the signing run, because a reusable workflow named in the certificate can be called from any repository; D-187 extends the pin to every GitHub-issued signer. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Every signer identity GitHub issues, a catalog publisher's or one the owner adds (D-45), is pinned to its source repository and ref as OpenFief's CI is (D-181).

OpenFief author decisions (docs/DESIGN.md, section 3.6), design-decisions:D-187#1
