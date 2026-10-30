---
title: The built-in UI must ship as the plugin image ui whose manifest asks for the view capability only, with no egress host and no Secret, serving the owner's browser and a health endpoint on the plugin port, and the demo must serve it at / through its Ingress to the core's http port while /v1 stays the core's.
status: open
text_ref: docs/DESIGN.md#D-28
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-28
  - lane:strict
semantic_text: The built-in UI must ship as the plugin image ui whose manifest asks for the view capability only, with no egress host and no Secret, serving the owner's browser and a health endpoint on the plugin port, and the demo must serve it at / through its Ingress to the core's http port while /v1 stays the core's.
logic_claims:
  - CLAIM-A7CDB767C046AB1B
semantic_clauses:
  - The built-in UI must ship as the plugin image ui whose manifest asks for the view capability only, with no egress host and no Secret, serving the owner's browser and a health endpoint on the plugin port, and the demo must serve it at / through its Ingress to the core's http port while /v1 stays the core's.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 191c90ac4ebd5a4dfc9e3cafbe005232c09bf4f8e775c466816374f93d373142
semantic_inventory:
  - claim_key: CLAIM-A7CDB767C046AB1B
    claim_text: The built-in UI must ship as the plugin image ui whose manifest asks for the view capability only, with no egress host and no Secret, serving the owner's browser and a health endpoint on the plugin port, and the demo must serve it at / through its Ingress to the core's http port while /v1 stays the core's
    role: normative
    status: modeled
    span:
      start: 0
      end: 306
origin:
  kind: agent
  recorded_at: '2026-10-28T19:50:19.106Z'
id: REQ-ui-plugin-packaging
type: req
---
The built-in UI must ship as the plugin image ui whose manifest asks for the view capability only, with no egress host and no Secret, serving the owner's browser and a health endpoint on the plugin port, and the demo must serve it at / through its Ingress to the core's http port while /v1 stays the core's.

## Context

D-3 ships a built-in UI as a view plugin; D-81 makes the core the only entry point for browsers, so the demo's Ingress, the installation's own way in (D-28), points at the core's browser port, Service port http, and the core proxies the UI's paths to it.

## Source

> OpenFief exposes nothing by default: every Service it creates is ClusterIP, the core's being the only way in (D-81), and it ships no tunnel. An optional kustomize component adds a Gateway API HTTPRoute to the core, which the owner points at their own Gateway and hostname.

docs/DESIGN.md#D-28
