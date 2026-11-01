---
title: The read API must list every sleeve kind with its default role, every connector binding and each sleeve's role override, as additions only.
status: open
text_ref: docs/DESIGN.md#D-121
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-121
  - lane:strict
semantic_text: The read API must list every sleeve kind with its default role, every connector binding and each sleeve's role override, as additions only.
logic_claims:
  - CLAIM-C302C07BB47402D6
semantic_clauses:
  - The read API must list every sleeve kind with its default role, every connector binding and each sleeve's role override, as additions only.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 5039663f1c7f5961690a75cc113edf37fb077117846fec54f06ab98ec18bd311
semantic_inventory:
  - claim_key: CLAIM-C302C07BB47402D6
    claim_text: The read API must list every sleeve kind with its default role, every connector binding and each sleeve's role override, as additions only
    role: normative
    status: modeled
    span:
      start: 0
      end: 138
origin:
  kind: agent
  recorded_at: '2026-10-29T12:46:11.268Z'
id: REQ-views-onboarding-lists
type: req
---
The read API must list every sleeve kind with its default role, every connector binding and each sleeve's role override, as additions only.

## Context

Refined for D-121: broker account numbers are listed only through the owner's account service, so the read API no longer lists broker accounts. Recorded by the implementation agent of impl/decisions-117; default firmness, not reviewed by the author yet.

## Source

> Broker account numbers are listed only through the owner's account service, for setting up accounts. The read API never serves them, so read sessions, script tokens and plugins never see one.

docs/DESIGN.md#D-121
