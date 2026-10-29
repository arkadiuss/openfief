---
title: The core must serve the AuthService and the owner services only on its plugin port, which the Service's connect port targets and only plugin pods may reach.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-115-2
  - lane:strict
text_ref: design-decisions:D-115#2
origin:
  kind: import
  ref: design-decisions:D-115#2
  recorded_at: '2026-10-29T16:10:51.000Z'
semantic_text: The core must serve the AuthService and the owner services only on its plugin port, which the Service's connect port targets and only plugin pods may reach.
logic_claims:
  - CLAIM-33987154DC912BCE
semantic_clauses:
  - The core must serve the AuthService and the owner services only on its plugin port, which the Service's connect port targets and only plugin pods may reach.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 5e09d3b64295640c52127af4fc66d7744037cf788e296f31118632213e079ed7
semantic_inventory:
  - claim_key: CLAIM-33987154DC912BCE
    claim_text: The core must serve the AuthService and the owner services only on its plugin port, which the Service's connect port targets and only plugin pods may reach
    role: normative
    status: modeled
    span:
      start: 0
      end: 155
id: REQ-core-plugin-port-scope
type: req
---
The core must serve the AuthService and the owner services only on its plugin port, which the Service's connect port targets and only plugin pods may reach.

## Context

Only view plugins call the AuthService (D-69, REQ-identity-auth-callers) and the owner services (D-79, REQ-core-owner-entry-service-callers), but they shared one port with what browsers reach, so whatever route the owner points at the core (D-28, REQ-platform-httproute-component) reached them too. The author decided D-115 while answering the implementation's question about the core's plugin-only Connect services sharing the browsers' port: they move to a plugin port that only plugin pods may reach, which also serves the read API, so no route reaches a plugin-only service. REQ-core-browser-port-scope records the browser port.

## Source

> The core listens on two ports. The browser port, which the Service's `http` port targets and a route points at, serves the front door (D-81), the read API (D-50) and the health endpoints. The plugin port, which the Service's `connect` port targets and only plugin pods may reach, serves the read API and the services only plugins call: the AuthService (D-69) and the owner services (D-79). No route reaches a plugin-only service.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-115#2
