---
title: The Models page must let the owner list, for each data tier, the models allowed to receive it, as a checkbox for each model and tier with nothing ticked until the owner ticks it, and must save only the tiers that changed.
status: open
firmness: default
required_evidence: unverified
origin:
  kind: agent
  ref: the MS1 brief, from D-206, 2026-11-13
  recorded_at: '2026-11-14T00:24:11.340Z'
semantic_text: The Models page must let the owner list, for each data tier, the models allowed to receive it, as a checkbox for each model and tier with nothing ticked until the owner ticks it, and must save only the tiers that changed.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 1943c3e09e325a351dcd3edeaf55df3ebecc5a21cb15b15477c015b5aafad880
semantic_inventory:
  - claim_key: CLAIM-EBBFB8F6996BFC04
    claim_text: The Models page must let the owner list, for each data tier, the models allowed to receive it, as a checkbox for each model
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 123
  - claim_key: CLAIM-99819D69DA29D7DF
    claim_text: tier with nothing ticked until the owner ticks it, and must save only the tiers that changed
    role: normative
    status: ontology_gap
    span:
      start: 128
      end: 220
logic_claims:
  - CLAIM-EBBFB8F6996BFC04
  - CLAIM-99819D69DA29D7DF
semantic_clauses:
  - The Models page must let the owner list, for each data tier, the models allowed to receive it, as a checkbox for each model
  - tier with nothing ticked until the owner ticks it, and must save only the tiers that changed
id: REQ-ui-models-tier-lists
type: req
---
The Models page must let the owner list, for each data tier, the models allowed to receive it, as a checkbox for each model and tier with nothing ticked until the owner ticks it, and must save only the tiers that changed.

## Context

D-206 makes the per-tier lists the owner's, with deny by default. The page shows the four tiers in plain words and a grid of models and tiers; a model a tier allows that no provider offers any more stays listed so it can be unticked.

## Source

> The owner lists, per tier, the models allowed to receive it, and the router sends a request only to a model allowed for every tier in it, denying by default.

Source: the MS1 brief, from D-206, 2026-11-13
