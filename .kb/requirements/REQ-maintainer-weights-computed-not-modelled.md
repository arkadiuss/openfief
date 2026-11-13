---
title: OpenFief must compute each instrument target's current weight and its drift from the owner's valued holdings as exact decimals, and no model computes them.
status: open
firmness: default
origin:
  kind: agent
  ref: the M3 brief for the model-portfolio maintainer (D-204) on the model router, 2026-11-13
  recorded_at: '2026-11-13T21:07:39.603Z'
semantic_text: OpenFief must compute each instrument target's current weight and its drift from the owner's valued holdings as exact decimals, and no model computes them.
semantic_source_field: semantic_text
semantic_source_hash: 268f7f976b0d4aa086a70369ea80f727df0190e3693fe9da413bb0ac903adae1
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_inventory:
  - claim_key: CLAIM-B166D9E17A65D801
    claim_text: OpenFief must compute each instrument target's current weight and its drift from the owner's valued holdings as exact decimals, and no model computes them
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 154
logic_claims:
  - CLAIM-B166D9E17A65D801
semantic_clauses:
  - OpenFief must compute each instrument target's current weight and its drift from the owner's valued holdings as exact decimals, and no model computes them
required_evidence: unverified
id: REQ-maintainer-weights-computed-not-modelled
type: req
---
OpenFief must compute each instrument target's current weight and its drift from the owner's valued holdings as exact decimals, and no model computes them.

## Context

D-204 says OpenFief computes the weights and drift; the M3 brief adds that the figures are pure domain code over decimals, never a float, and that the model only reads them.

## Source

> OpenFief computes weights and drift (existing domain code - pure, no model).

Source: the M3 brief for the model-portfolio maintainer (D-204) on the model router, 2026-11-13