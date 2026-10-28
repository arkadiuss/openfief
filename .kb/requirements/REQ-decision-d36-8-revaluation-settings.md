---
title: Worker concurrency, the end-of-day schedule and the bond valuation basis must be configurable.
status: open
text_ref: design-decisions:D-36#8
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-36-8
  - lane:strict
semantic_text: Worker concurrency, the end-of-day schedule and the bond valuation basis must be configurable.
logic_claims:
  - CLAIM-745815536C0ED4B6
semantic_clauses:
  - Worker concurrency, the end-of-day schedule and the bond valuation basis must be configurable.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: eb530f6ff9db9a51e35027a6113b2aa87380c6b9b385a5e90baf6264894083d4
semantic_inventory:
  - claim_key: CLAIM-745815536C0ED4B6
    claim_text: Worker concurrency, the end-of-day schedule and the bond valuation basis must be configurable
    role: normative
    status: modeled
    span:
      start: 0
      end: 93
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-36#8
  recorded_at: '2026-10-28T06:30:10.038Z'
id: REQ-decision-d36-8-revaluation-settings
type: req
---
Worker concurrency, the end-of-day schedule and the bond valuation basis must be configurable.

## Context

The author decided D-36 in docs/DESIGN.md section 3.3; this requirement imports its claim D-36#8 as a firm rule, so changing it needs the author. Reason not stated in the decision. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Configurable: worker concurrency, the end-of-day schedule and the bond basis (D-25); changing the basis revalues everything.

Source: OpenFief author decisions (docs/DESIGN.md section 3.3) - D-36#8
