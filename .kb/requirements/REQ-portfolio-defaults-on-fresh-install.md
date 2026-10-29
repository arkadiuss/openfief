---
title: Ensuring the defaults must store every missing built-in sleeve kind and, only while no sleeve is stored at all, the default sleeves Bonds, Core and Satellite, in one database transaction, and must never change a stored kind or sleeve.
status: open
text_ref: docs/DESIGN.md#D-71
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-71
  - lane:strict
semantic_text: Ensuring the defaults must store every missing built-in sleeve kind and, only while no sleeve is stored at all, the default sleeves Bonds, Core and Satellite, in one database transaction, and must never change a stored kind or sleeve.
logic_claims:
  - CLAIM-572B3C5D2E082FC7
semantic_clauses:
  - Ensuring the defaults must store every missing built-in sleeve kind and, only while no sleeve is stored at all, the default sleeves Bonds, Core and Satellite, in one database transaction, and must never change a stored kind or sleeve.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: a6670af4419d28fc88326df2f8f806ad1bf16b72095e3e89b701910f48f6d703
semantic_inventory:
  - claim_key: CLAIM-572B3C5D2E082FC7
    claim_text: Ensuring the defaults must store every missing built-in sleeve kind and, only while no sleeve is stored at all, the default sleeves Bonds, Core and Satellite, in one database transaction, and must never change a stored kind or sleeve
    role: normative
    status: modeled
    span:
      start: 0
      end: 233
origin:
  kind: agent
  recorded_at: '2026-10-29T11:36:27.454Z'
id: REQ-portfolio-defaults-on-fresh-install
type: req
---
Ensuring the defaults must store every missing built-in sleeve kind and, only while no sleeve is stored at all, the default sleeves Bonds, Core and Satellite, in one database transaction, and must never change a stored kind or sleeve.

## Context

D-71: an install starts with three sleeves. The core ensures the defaults on every start, so once any sleeve is stored, the owner's or the demo dataset's, the install has started and no default sleeve is added again, which keeps the demo dataset's own two sleeves alone. Refines REQ-portfolio-install-sleeves and REQ-portfolio-builtin-kinds. Modeled by Claude Code for the onboarding work package (D-71, D-72, D-73, impl/onboarding); the author has not reviewed it yet.

## Source

> **[D-71]** Every sleeve has a sleeve kind, which gives it a default role: defensive (never available to agents) or offensive (fiefs are leased from it, GR-6). The built-in kinds are agent pool and satellite, which are offensive, and bonds, core, emergency fund and retirement, which are defensive. The owner can add kinds and rename any kind, but a kind's default role never changes, and a sleeve whose role differs from its kind's needs the owner's explicit choice. An install starts with three sleeves: bonds, core and satellite.

docs/DESIGN.md#D-71
