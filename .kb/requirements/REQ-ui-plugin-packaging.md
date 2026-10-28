---
title: The built-in UI must ship as the plugin image ui whose manifest asks for the view capability only, with no egress host and no Secret, serving the owner's browser and a health endpoint on the plugin port, and the demo must serve it at / through its Ingress while /v1 stays the core's.
status: open
text_ref: docs/DESIGN.md#D-3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-3
  - lane:strict
semantic_text: The built-in UI must ship as the plugin image ui whose manifest asks for the view capability only, with no egress host and no Secret, serving the owner's browser and a health endpoint on the plugin port, and the demo must serve it at / through its Ingress while /v1 stays the core's.
logic_claims:
  - CLAIM-02EC642940108645
semantic_clauses:
  - The built-in UI must ship as the plugin image ui whose manifest asks for the view capability only, with no egress host and no Secret, serving the owner's browser and a health endpoint on the plugin port, and the demo must serve it at / through its Ingress while /v1 stays the core's.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: e054fae52af8c128a362b9bfa0d44ce1ac944c1093ca40d4668d2ba05f9472fa
semantic_inventory:
  - claim_key: CLAIM-02EC642940108645
    claim_text: The built-in UI must ship as the plugin image ui whose manifest asks for the view capability only, with no egress host and no Secret, serving the owner's browser and a health endpoint on the plugin port, and the demo must serve it at / through its Ingress while /v1 stays the core's
    role: normative
    status: modeled
    span:
      start: 0
      end: 282
origin:
  kind: agent
  recorded_at: '2026-10-28T19:50:19.106Z'
id: REQ-ui-plugin-packaging
type: req
---
The built-in UI must ship as the plugin image ui whose manifest asks for the view capability only, with no egress host and no Secret, serving the owner's browser and a health endpoint on the plugin port, and the demo must serve it at / through its Ingress while /v1 stays the core's.

## Context

D-3 (the built-in UI is the default view plugin), D-11 (manifests), D-42 (manifest label).

## Source

> M0 ships two view plugins: a very simple built-in UI as the default, and Ghostfolio auto-push.

docs/DESIGN.md#D-3
