---
title: On both kind and k3d, the core's plugin port must refuse a plugin's call without a valid view plugin token.
status: open
text_ref: docs/DESIGN.md#D-200
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-200
  - lane:strict
semantic_text: On both kind and k3d, the core's plugin port must refuse a plugin's call without a valid view plugin token.
logic_claims:
  - CLAIM-1545E0DF73A90666
semantic_clauses:
  - On both kind and k3d, the core's plugin port must refuse a plugin's call without a valid view plugin token.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: bc9dd10152fc03579645b90c6e341727c16fdfeed4f87eda2ab9e66d2f03f68f
semantic_inventory:
  - claim_key: CLAIM-1545E0DF73A90666
    claim_text: On both kind and k3d, the core's plugin port must refuse a plugin's call without a valid view plugin token
    role: normative
    status: modeled
    span:
      start: 0
      end: 106
origin:
  kind: agent
  recorded_at: '2026-11-07T03:31:20.394Z'
id: REQ-e2e-core-plugin-port-refuses-bad-tokens
type: req
---
On both kind and k3d, the core's plugin port must refuse a plugin's call without a valid view plugin token.

## Context

A refinement next to REQ-m0-exit-17-core-to-plugin, which covers calls to connectors; D-115 keeps the plugin port for plugins with valid tokens. It carries this clause of REQ-e2e-core-plugin-tokens, which the firm requirement replaced. Modeled by Claude Code for the D-200 follow-ups; the author has not reviewed the wording.

## Source

> **Core-to-plugin.** A call to a connector from any pod other than the core fails, and a call from the core without a valid token for that instance is refused.

design-decisions:D-200#17 (docs/DESIGN.md, D-200, criterion 17); D-115
