---
title: The core must serve BondLotService on its plugin port to view plugins, reach the bound connector instance through the client its connector discovery builds, answering unavailable until then, and record delivered bond series terms through valuation so imported lots are valued by the bond rules.
status: open
text_ref: docs/DESIGN.md#D-115
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-115
  - lane:strict
semantic_text: The core must serve BondLotService on its plugin port to view plugins, reach the bound connector instance through the client its connector discovery builds, answering unavailable until then, and record delivered bond series terms through valuation so imported lots are valued by the bond rules.
logic_claims:
  - CLAIM-49EF810348D6F835
semantic_clauses:
  - The core must serve BondLotService on its plugin port to view plugins, reach the bound connector instance through the client its connector discovery builds, answering unavailable until then, and record delivered bond series terms through valuation so imported lots are valued by the bond rules.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: f9a20987d1cf11004588c74d2bcd65e553b67f8facbcaa814e212415c8d0ec37
semantic_inventory:
  - claim_key: CLAIM-49EF810348D6F835
    claim_text: The core must serve BondLotService on its plugin port to view plugins, reach the bound connector instance through the client its connector discovery builds, answering unavailable until then, and record delivered bond series terms through valuation so imported lots are valued by the bond rules
    role: normative
    status: modeled
    span:
      start: 0
      end: 293
origin:
  kind: agent
  recorded_at: '2026-10-29T01:54:50.264Z'
id: REQ-core-bond-lot-service
type: req
---
The core must serve BondLotService on its plugin port to view plugins, reach the bound connector instance through the client its connector discovery builds, answering unavailable until then, and record delivered bond series terms through valuation so imported lots are valued by the bond rules.

## Context

D-49 has the core call connectors and D-79 gives each kind of hand-entered data its own owner service; the built-in UI calls it for the owner. D-115 serves the owner services on the plugin port only, so no route reaches them.

## Source

> The core listens on two ports. The browser port, which the Service's `http` port targets and a route points at, serves the front door (D-81), the read API (D-50) and the health endpoints. The plugin port, which the Service's `connect` port targets and only plugin pods may reach, serves the read API and the services only plugins call: the AuthService (D-69) and the owner services (D-79). No route reaches a plugin-only service.

docs/DESIGN.md#D-115
