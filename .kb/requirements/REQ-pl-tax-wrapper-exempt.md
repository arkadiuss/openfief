---
title: A bond lot held in an IKE or IKZE tax wrapper must have no Belka tax deducted from its net-if-redeemed value.
status: open
text_ref: docs/DESIGN.md#D-25
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-25
  - lane:strict
semantic_text: A bond lot held in an IKE or IKZE tax wrapper must have no Belka tax deducted from its net-if-redeemed value.
logic_claims:
  - CLAIM-5BEF5564038773C1
semantic_clauses:
  - A bond lot held in an IKE or IKZE tax wrapper must have no Belka tax deducted from its net-if-redeemed value.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 7ca99b7af51bcb8ea59e002ff488e41871c9ff0914af43d286879a7176e57a15
semantic_inventory:
  - claim_key: CLAIM-5BEF5564038773C1
    claim_text: A bond lot held in an IKE or IKZE tax wrapper must have no Belka tax deducted from its net-if-redeemed value
    role: normative
    status: modeled
    span:
      start: 0
      end: 108
origin:
  kind: agent
  recorded_at: '2026-10-28T05:22:15.429Z'
id: REQ-pl-tax-wrapper-exempt
type: req
---
A bond lot held in an IKE or IKZE tax wrapper must have no Belka tax deducted from its net-if-redeemed value.

## Context

D-25 names IKE explicitly. IKZE was added by the orchestrating agent's task: no Belka tax is withheld inside IKZE either, and the flat IKZE withdrawal tax applies to the whole account, so it is not part of one bond's value. Refines REQ-AUTO-E10BE8CE13E67657. Modeled by Claude Code for the M0 valuation task; the author has not reviewed it yet. obligacjeskarbowe.pl and gov.pl were not reachable from the agent's sandbox, so the rule was checked only against search-result excerpts of issue letters and secondary calculators.

## Source

> no tax inside IKE

docs/DESIGN.md D-25
