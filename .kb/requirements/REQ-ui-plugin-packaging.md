---
title: The built-in UI must ship as the plugin image ui whose manifest asks for the view capability only, with no egress host and no Secret, serving the owner's browser and a health endpoint on the plugin port.
status: open
text_ref: docs/DESIGN.md#D-3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-3
  - lane:strict
semantic_text: The built-in UI must ship as the plugin image ui whose manifest asks for the view capability only, with no egress host and no Secret, serving the owner's browser and a health endpoint on the plugin port.
logic_claims:
  - CLAIM-9F03990DC9DD7C03
semantic_clauses:
  - The built-in UI must ship as the plugin image ui whose manifest asks for the view capability only, with no egress host and no Secret, serving the owner's browser and a health endpoint on the plugin port.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 259f06fc85ed2e0af29f6c61e1e6c647b4f657be2a4b4637e4e3aef136a69d60
semantic_inventory:
  - claim_key: CLAIM-9F03990DC9DD7C03
    claim_text: The built-in UI must ship as the plugin image ui whose manifest asks for the view capability only, with no egress host and no Secret, serving the owner's browser and a health endpoint on the plugin port
    role: normative
    status: modeled
    span:
      start: 0
      end: 202
origin:
  kind: agent
  recorded_at: '2026-10-28T19:50:19.106Z'
id: REQ-ui-plugin-packaging
type: req
---
The built-in UI must ship as the plugin image ui whose manifest asks for the view capability only, with no egress host and no Secret, serving the owner's browser and a health endpoint on the plugin port.

## Context

D-3 (the built-in UI is the default view plugin), D-11 (manifests), D-42 (manifest label).

## Source

> M0 ships two view plugins: a very simple built-in UI as the default, and Ghostfolio auto-push.

docs/DESIGN.md#D-3
