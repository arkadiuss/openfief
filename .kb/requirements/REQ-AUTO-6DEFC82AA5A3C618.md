---
title: The OpenFief core must be the system of record for transactions, tax lots and valuations.
status: open
text_ref: design-decisions:D-2#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-2-1
  - lane:strict
semantic_text: The OpenFief core must be the system of record for transactions, tax lots and valuations.
logic_claims:
  - CLAIM-AD2931B4FD856153
semantic_clauses:
  - The OpenFief core must be the system of record for transactions, tax lots and valuations.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 76356698d764ae836b9edebd9ae51183c47f7c4848f7f748a499c069f540e97e
semantic_inventory:
  - claim_key: CLAIM-AD2931B4FD856153
    claim_text: The OpenFief core must be the system of record for transactions, tax lots and valuations
    role: normative
    status: modeled
    span:
      start: 0
      end: 88
origin:
  kind: import
  ref: design-decisions:D-2#1
  recorded_at: '2026-10-27T20:12:34.117Z'
firmness: firm
required_evidence: asserted
id: REQ-AUTO-6DEFC82AA5A3C618
type: req
---
The OpenFief core must be the system of record for transactions, tax lots and valuations.

## Source

> OpenFief's own core is the system of record for transactions, tax lots and valuations. Trackers are plugins that show a one-way *view* of it, a new plugin type next to P-2's four.

Source: OpenFief author decisions (DESIGN.md section 10) - D-2#1
