---
title: A plugin without a grant, or whose grant approves none of the capabilities of the version that would run, must not run.
status: open
text_ref: docs/DESIGN.md#D-16
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-16
  - lane:strict
semantic_text: A plugin without a grant, or whose grant approves none of the capabilities of the version that would run, must not run.
logic_claims:
  - CLAIM-8D88EB49D3524F8C
semantic_clauses:
  - A plugin without a grant, or whose grant approves none of the capabilities of the version that would run, must not run.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: c76d119ac0f567ff7e672a1d87d8daa55bd02daaf9ff66e75782a207a2e2f0ce
semantic_inventory:
  - claim_key: CLAIM-8D88EB49D3524F8C
    claim_text: A plugin without a grant, or whose grant approves none of the capabilities of the version that would run, must not run
    role: normative
    status: modeled
    span:
      start: 0
      end: 118
origin:
  kind: agent
  recorded_at: '2026-10-28T06:36:35.026Z'
id: REQ-plugins-rollout-no-grant
type: req
---
A plugin without a grant, or whose grant approves none of the capabilities of the version that would run, must not run.

## Context

D-16 has the operator enforce only what is both requested and approved. With no grant nothing is approved, and a plugin with no granted capability has nothing it may do (ErrNoCapabilityGranted), so neither runs; everything the current version requests is reported as pending, so the owner sees what to approve. Refines REQ-AUTO-12177128A788CEDC. Modeled by Claude Code for the M0 operator task; the author has not reviewed it yet.

## Source

> The operator enforces only what is in both.

docs/DESIGN.md D-16
