---
title: The operator must mark each PluginGrant as Valid or not with the reason, and must treat an invalid PluginGrant as approving nothing.
status: open
text_ref: docs/DESIGN.md#D-16
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-16
  - lane:strict
semantic_text: The operator must mark each PluginGrant as Valid or not with the reason, and must treat an invalid PluginGrant as approving nothing.
logic_claims:
  - CLAIM-17DB729F3E39FE97
semantic_clauses:
  - The operator must mark each PluginGrant as Valid or not with the reason, and must treat an invalid PluginGrant as approving nothing.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: ce9784f070d53783d635f05b1321b508e6ac37f34f4612ad73e261548f875d07
semantic_inventory:
  - claim_key: CLAIM-17DB729F3E39FE97
    claim_text: The operator must mark each PluginGrant as Valid or not with the reason, and must treat an invalid PluginGrant as approving nothing
    role: normative
    status: modeled
    span:
      start: 0
      end: 131
origin:
  kind: agent
  recorded_at: '2026-10-28T07:00:39.477Z'
id: REQ-operator-grant-validity
type: req
---
The operator must mark each PluginGrant as Valid or not with the reason, and must treat an invalid PluginGrant as approving nothing.

## Context

D-16 makes the PluginGrant the owner's approval. A grant the domain refuses, such as one with a wildcard host, cannot be partly enforced safely, so the plugin runs nothing and both resources say why. Refines REQ-AUTO-4E891679754B3360. Modeled by Claude Code for the M0 operator task; the author has not reviewed it yet.

## Source

> a `PluginGrant` (what the user approved)

docs/DESIGN.md D-16
