---
title: The core's browser port, which the Service's http port targets, must serve only the front door, the read API and the health endpoints.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-115-1
  - lane:strict
text_ref: design-decisions:D-115#1
origin:
  kind: import
  ref: design-decisions:D-115#1
  recorded_at: '2026-10-29T16:10:51.000Z'
semantic_text: The core's browser port, which the Service's http port targets, must serve only the front door, the read API and the health endpoints.
logic_claims:
  - CLAIM-56B90E86E4B73A4D
semantic_clauses:
  - The core's browser port, which the Service's http port targets, must serve only the front door, the read API and the health endpoints.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 670f0385bbd82eda1c93e302d874c7869aa5101e42969fadfe5064dcfdd7af97
semantic_inventory:
  - claim_key: CLAIM-56B90E86E4B73A4D
    claim_text: The core's browser port, which the Service's http port targets, must serve only the front door, the read API and the health endpoints
    role: normative
    status: modeled
    span:
      start: 0
      end: 133
id: REQ-core-browser-port-scope
type: req
---
The core's browser port, which the Service's http port targets, must serve only the front door, the read API and the health endpoints.

## Context

The core's plugin-only Connect services, the AuthService (D-69) and the owner services (D-79), shared one port with what browsers reach, so whatever route the owner points at the core (D-28, REQ-platform-httproute-component) reached them too. The author decided D-115 while answering the implementation's question about that shared port: the core listens on two ports, and the browser port a route points at carries only the front door (D-81), the read API (D-50) and the health endpoints, so no route reaches a plugin-only service. REQ-core-plugin-port-scope records the plugin port.

## Source

> The core listens on two ports. The browser port, which the Service's `http` port targets and a route points at, serves the front door (D-81), the read API (D-50) and the health endpoints. The plugin port, which the Service's `connect` port targets and only plugin pods may reach, serves the read API and the services only plugins call: the AuthService (D-69) and the owner services (D-79). No route reaches a plugin-only service.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-115#1
