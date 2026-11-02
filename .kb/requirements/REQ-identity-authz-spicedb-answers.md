---
title: From the second alpha, SpiceDB must answer the authorization interface, with its datastore a database in the core's PostgreSQL cluster.
status: open
text_ref: design-decisions:D-164#3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-164-3
  - lane:strict
semantic_text: From the second alpha, SpiceDB must answer the authorization interface, with its datastore a database in the core's PostgreSQL cluster.
logic_claims:
  - CLAIM-98FCF70A0FC521C2
semantic_clauses:
  - From the second alpha, SpiceDB must answer the authorization interface, with its datastore a database in the core's PostgreSQL cluster.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 3fa2afbefe1b953ea3ecdb984119d4841eb8350de0c337a64d431cf9c4fcf2b1
semantic_inventory:
  - claim_key: CLAIM-98FCF70A0FC521C2
    claim_text: From the second alpha, SpiceDB must answer the authorization interface, with its datastore a database in the core's PostgreSQL cluster
    role: normative
    status: modeled
    span:
      start: 0
      end: 134
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-164#3
  recorded_at: '2026-11-02T00:57:26.624Z'
id: REQ-identity-authz-spicedb-answers
type: req
---
From the second alpha, SpiceDB must answer the authorization interface, with its datastore a database in the core's PostgreSQL cluster.

## Context

The author recorded it on 2026-11-01, when planning viewers for the second alpha. Reason not stated. D-14 runs the core's PostgreSQL cluster with CloudNativePG. The first alpha uses the rule D-163 states instead. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> The authorization model is written as a SpiceDB schema, whose assertions CI checks. From the second alpha, SpiceDB answers the interface (D-163), with its datastore a database in the core's PostgreSQL cluster (D-14), and script tokens and view plugins are granted accounts as viewers are (D-65).

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-164#3
