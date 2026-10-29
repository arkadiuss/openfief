---
title: A sleeve whose role differs from its sleeve kind's default role, defensive in an offensive kind or offensive in a defensive kind, must carry the owner's explicit role override, a sleeve whose role is its kind's must not carry it, and the database must refuse either.
status: open
text_ref: docs/DESIGN.md#D-71
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-71
  - lane:strict
semantic_text: A sleeve whose role differs from its sleeve kind's default role, defensive in an offensive kind or offensive in a defensive kind, must carry the owner's explicit role override, a sleeve whose role is its kind's must not carry it, and the database must refuse either.
logic_claims:
  - CLAIM-D42F9B192168432C
semantic_clauses:
  - A sleeve whose role differs from its sleeve kind's default role, defensive in an offensive kind or offensive in a defensive kind, must carry the owner's explicit role override, a sleeve whose role is its kind's must not carry it, and the database must refuse either.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: f17b77e110b2593414dfcc1f4a50a24bf0c00528dfeb5567ca51129d87541a3e
semantic_inventory:
  - claim_key: CLAIM-D42F9B192168432C
    claim_text: A sleeve whose role differs from its sleeve kind's default role, defensive in an offensive kind or offensive in a defensive kind, must carry the owner's explicit role override, a sleeve whose role is its kind's must not carry it, and the database must refuse either
    role: normative
    status: modeled
    span:
      start: 0
      end: 265
origin:
  kind: agent
  recorded_at: '2026-10-29T11:36:16.778Z'
id: REQ-portfolio-role-override-explicit
type: req
---
A sleeve whose role differs from its sleeve kind's default role, defensive in an offensive kind or offensive in a defensive kind, must carry the owner's explicit role override, a sleeve whose role is its kind's must not carry it, and the database must refuse either.

## Context

D-71 makes any role that differs from the kind's default the owner's explicit choice, in either direction, so capital moves into or out of agents' reach (GR-6) only when the owner asks for exactly that. The earlier rule let a sleeve be defensive in an offensive kind without it; migration 0003 gives such stored sleeves the override. Refines REQ-portfolio-role-differs-explicit. Modeled by Claude Code for the onboarding work package (D-71, D-72, D-73, impl/onboarding); the author has not reviewed it yet.

## Source

> **[D-71]** Every sleeve has a sleeve kind, which gives it a default role: defensive (never available to agents) or offensive (fiefs are leased from it, GR-6). The built-in kinds are agent pool and satellite, which are offensive, and bonds, core, emergency fund and retirement, which are defensive. The owner can add kinds and rename any kind, but a kind's default role never changes, and a sleeve whose role differs from its kind's needs the owner's explicit choice. An install starts with three sleeves: bonds, core and satellite.

docs/DESIGN.md#D-71
