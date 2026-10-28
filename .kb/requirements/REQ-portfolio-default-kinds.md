---
title: OpenFief must provide the built-in sleeve kinds emergency-fund, bonds, retirement and core as defensive and satellite and agent-pool as offensive, and an install must start with the sleeves Bonds, Core and Satellite, creating each missing default once and never changing one that is stored.
status: open
text_ref: docs/DESIGN.md#GR-6
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-gr-6
  - lane:strict
semantic_text: OpenFief must provide the built-in sleeve kinds emergency-fund, bonds, retirement and core as defensive and satellite and agent-pool as offensive, and an install must start with the sleeves Bonds, Core and Satellite, creating each missing default once and never changing one that is stored.
logic_claims:
  - CLAIM-1DD11604954CF9D0
semantic_clauses:
  - OpenFief must provide the built-in sleeve kinds emergency-fund, bonds, retirement and core as defensive and satellite and agent-pool as offensive, and an install must start with the sleeves Bonds, Core and Satellite, creating each missing default once and never changing one that is stored.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: fc66f31a2e28a49bf5560069d734fd570b213d0248587fd2803e63b07a8ebad2
semantic_inventory:
  - claim_key: CLAIM-1DD11604954CF9D0
    claim_text: OpenFief must provide the built-in sleeve kinds emergency-fund, bonds, retirement and core as defensive and satellite and agent-pool as offensive, and an install must start with the sleeves Bonds, Core and Satellite, creating each missing default once and never changing one that is stored
    role: normative
    status: modeled
    span:
      start: 0
      end: 289
origin:
  kind: agent
  recorded_at: '2026-10-28T09:35:32.901Z'
id: REQ-portfolio-default-kinds
type: req
---
OpenFief must provide the built-in sleeve kinds emergency-fund, bonds, retirement and core as defensive and satellite and agent-pool as offensive, and an install must start with the sleeves Bonds, Core and Satellite, creating each missing default once and never changing one that is stored.

## Context

Author decision, sleeve kinds (given in chat to the orchestrator, not yet recorded in docs/DESIGN.md): role stays a fixed enum, defensive or offensive, because code enforces that agents may lease only from offensive sleeves (GR-6); a sleeve kind is a named category with a default role, from a built-in set the owner can extend; a sleeve has a kind, and its role defaults from the kind but is stored on the sleeve. The author listed the six built-in kinds and suggested Core, Bonds and Satellite as the default sleeves, created idempotently at install. Retirement covers the Polish IKE and IKZE tax wrappers; agent-pool is the pool fiefs are leased from in M1 and is only a kind in M0. Defaults the owner renamed or moved are left as stored. Modeled by Claude Code for the sleeve-kinds task; the author has not reviewed it yet.

## Source

> Role stays a fixed enum: defensive | offensive. Code enforces it: agents may only lease from offensive sleeves (GR-6), so roles are never owner-defined. Sleeve kind is new: a named category with a default role, from a built-in default set that the owner can extend. Defaults: emergency-fund (defensive), bonds (defensive), retirement (defensive; IKE/IKZE), core (defensive; long-term), satellite (offensive), agent-pool (offensive). A sleeve has a kind. Its role defaults from the kind but is stored on the sleeve. A default set of sleeves is created at install. The allocation chart (D-30) can group by kind.

author decision, sleeve kinds (chat); docs/DESIGN.md#GR-6, docs/DESIGN.md#D-30
