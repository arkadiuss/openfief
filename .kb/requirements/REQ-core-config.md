---
title: The core must read its database connection string from OPENFIEF_DSN and from flags with documented defaults its browser port's and plugin port's listen addresses, reporting currency, time zone, end-of-day time, revaluation concurrency, bond basis and shutdown timeout, and refuse to start on a missing or invalid value.
status: open
text_ref: docs/DESIGN.md#D-115
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-115
  - lane:strict
semantic_text: The core must read its database connection string from OPENFIEF_DSN and from flags with documented defaults its browser port's and plugin port's listen addresses, reporting currency, time zone, end-of-day time, revaluation concurrency, bond basis and shutdown timeout, and refuse to start on a missing or invalid value.
logic_claims:
  - CLAIM-88D1458C01FB9511
semantic_clauses:
  - The core must read its database connection string from OPENFIEF_DSN and from flags with documented defaults its browser port's and plugin port's listen addresses, reporting currency, time zone, end-of-day time, revaluation concurrency, bond basis and shutdown timeout, and refuse to start on a missing or invalid value.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 849adf95a09a28997a6e0cb912a3b12793cfa8eef866284f809c15c9c810ff38
semantic_inventory:
  - claim_key: CLAIM-88D1458C01FB9511
    claim_text: The core must read its database connection string from OPENFIEF_DSN and from flags with documented defaults its browser port's and plugin port's listen addresses, reporting currency, time zone, end-of-day time, revaluation concurrency, bond basis and shutdown timeout, and refuse to start on a missing or invalid value
    role: normative
    status: modeled
    span:
      start: 0
      end: 318
origin:
  kind: agent
  recorded_at: '2026-10-28T14:33:18.485Z'
id: REQ-core-config
type: req
---
The core must read its database connection string from OPENFIEF_DSN and from flags with documented defaults its browser port's and plugin port's listen addresses, reporting currency, time zone, end-of-day time, revaluation concurrency, bond basis and shutdown timeout, and refuse to start on a missing or invalid value.

## Context

D-36 makes the end of day, concurrency and bond basis configurable and D-62 the reporting currency; D-115 gives the core a second listener, so --listen names the browser port's address and --plugin-listen the plugin port's (defaults :8080 and :8081, which deploy/core's Service ports http and connect target).

## Source

> The core listens on two ports. The browser port, which the Service's `http` port targets and a route points at, serves the front door (D-81), the read API (D-50) and the health endpoints. The plugin port, which the Service's `connect` port targets and only plugin pods may reach, serves the read API and the services only plugins call: the AuthService (D-69) and the owner services (D-79). No route reaches a plugin-only service.

docs/DESIGN.md#D-115
