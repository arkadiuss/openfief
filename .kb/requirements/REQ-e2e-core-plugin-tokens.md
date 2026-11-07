---
title: On both kind and k3d, a call to a connector from any pod other than the core must fail, a call from the core without the core's valid token for that instance must be refused, and the core's plugin port must refuse a plugin's call without a valid view plugin token.
status: closed
text_ref: docs/DESIGN.md#D-200
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-200
  - lane:strict
semantic_text: On both kind and k3d, a call to a connector from any pod other than the core must fail, a call from the core without the core's valid token for that instance must be refused, and the core's plugin port must refuse a plugin's call without a valid view plugin token.
logic_claims:
  - CLAIM-B7B0E0D9879D0282
semantic_clauses:
  - On both kind and k3d, a call to a connector from any pod other than the core must fail, a call from the core without the core's valid token for that instance must be refused, and the core's plugin port must refuse a plugin's call without a valid view plugin token.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 612fc56698c142ffd024576607d38a2dff668d667510fd18b6e37f19adcb8d88
semantic_inventory:
  - claim_key: CLAIM-B7B0E0D9879D0282
    claim_text: On both kind and k3d, a call to a connector from any pod other than the core must fail, a call from the core without the core's valid token for that instance must be refused, and the core's plugin port must refuse a plugin's call without a valid view plugin token
    role: normative
    status: modeled
    span:
      start: 0
      end: 263
origin:
  kind: agent
  recorded_at: '2026-11-06T20:50:06.136Z'
id: REQ-e2e-core-plugin-tokens
type: req
---
On both kind and k3d, a call to a connector from any pod other than the core must fail, a call from the core without the core's valid token for that instance must be refused, and the core's plugin port must refuse a plugin's call without a valid view plugin token.

## Context

D-200 makes criterion 17 of M0's exit an end-to-end scenario on the firm requirement it checks, D-59's core-only caller with a token per instance, on both kind and k3d. The orchestrator asked to cover the other direction too: each plugin-facing core API refuses a call without a valid token (D-69, D-83, D-115).

## Source

> **Core-to-plugin.** A call to a connector from any pod other than the core fails, and a call from the core without a valid token for that instance is refused.

design-decisions:D-200#17 (docs/DESIGN.md, D-200, criterion 17)
