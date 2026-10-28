---
title: A plugin's identity must be the ServiceAccount plugin-<name> in the openfief namespace, and only such a ServiceAccount username may map back to a plugin.
status: open
text_ref: docs/DESIGN.md#D-32
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-32
  - lane:strict
semantic_text: A plugin's identity must be the ServiceAccount plugin-<name> in the openfief namespace, and only such a ServiceAccount username may map back to a plugin.
logic_claims:
  - CLAIM-AA0B541901981031
semantic_clauses:
  - A plugin's identity must be the ServiceAccount plugin-<name> in the openfief namespace, and only such a ServiceAccount username may map back to a plugin.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: b767d4f294493e78b73843cf53e5d78d0d3caa464370c53cf53d7be81b6836a3
semantic_inventory:
  - claim_key: CLAIM-AA0B541901981031
    claim_text: A plugin's identity must be the ServiceAccount plugin-<name> in the openfief namespace, and only such a ServiceAccount username may map back to a plugin
    role: normative
    status: modeled
    span:
      start: 0
      end: 152
origin:
  kind: agent
  recorded_at: '2026-10-28T05:28:55.597Z'
id: REQ-plugins-identity-service-account
type: req
---
A plugin's identity must be the ServiceAccount plugin-<name> in the openfief namespace, and only such a ServiceAccount username may map back to a plugin.

## Context

D-32 has the egress proxy identify plugins by their projected ServiceAccount token through a custom role function, and D-27 has the core authenticate plugins the same way. Both need one deterministic mapping between a plugin and its ServiceAccount; restricting it to the openfief namespace (D-15) and the plugin- prefix keeps the operator's own and other namespaces' accounts from ever counting as a plugin. Refines REQ-AUTO-15828B9727AFA488. Modeled by Claude Code for the M0 plugins task (pure domain package internal/plugins); the author has not reviewed it yet.

## Source

> Plugins identify themselves with their projected ServiceAccount token, through a custom role function.

docs/DESIGN.md D-32
