---
title: Recording a price that is not positive or names no instrument, an FX rate that is not positive or is for PLN, an index reading of an unknown kind or month, bond terms without an ISIN or series, or an unknown tax wrapper must fail and store nothing.
status: open
text_ref: docs/DESIGN.md#D-36
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-36
  - lane:strict
semantic_text: Recording a price that is not positive or names no instrument, an FX rate that is not positive or is for PLN, an index reading of an unknown kind or month, bond terms without an ISIN or series, or an unknown tax wrapper must fail and store nothing.
logic_claims:
  - CLAIM-7CD78B248D445C75
semantic_clauses:
  - Recording a price that is not positive or names no instrument, an FX rate that is not positive or is for PLN, an index reading of an unknown kind or month, bond terms without an ISIN or series, or an unknown tax wrapper must fail and store nothing.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: bfb2ef3a07425d92d937e9b5379e35a1c7482aa19a99c106377edfa6d0c8631f
semantic_inventory:
  - claim_key: CLAIM-7CD78B248D445C75
    claim_text: Recording a price that is not positive or names no instrument, an FX rate that is not positive or is for PLN, an index reading of an unknown kind or month, bond terms without an ISIN or series, or an unknown tax wrapper must fail and store nothing
    role: normative
    status: modeled
    span:
      start: 0
      end: 247
origin:
  kind: agent
  recorded_at: '2026-10-28T07:31:09.976Z'
id: REQ-valuation-input-validation
type: req
---
Recording a price that is not positive or names no instrument, an FX rate that is not positive or is for PLN, an index reading of an unknown kind or month, bond terms without an ISIN or series, or an unknown tax wrapper must fail and store nothing.

## Context

Stored inputs are what every valuation is rebuilt from (D-36), so a bad one would poison every later revaluation; rejecting it before the transaction keeps the store and the queue untouched. Reason not stated by the author. Modeled by Claude Code for the M0 revaluation task (D-36); the author has not reviewed it yet.

## Source

> Valuations are stored rows derived from the transaction log, prices, FX rates and index readings, and can be rebuilt from them at any time.

docs/DESIGN.md D-36
