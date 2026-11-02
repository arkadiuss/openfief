---
title: The authorization model must be written as a SpiceDB schema.
status: open
text_ref: design-decisions:D-164#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-164-1
  - lane:strict
semantic_text: The authorization model must be written as a SpiceDB schema.
logic_claims:
  - CLAIM-60D98FFF2E6E0236
semantic_clauses:
  - The authorization model must be written as a SpiceDB schema.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: be099bfcfdcc554b482c96097cd8c11dd2c19209b16867168c61190bd8ee1b6b
semantic_inventory:
  - claim_key: CLAIM-60D98FFF2E6E0236
    claim_text: The authorization model must be written as a SpiceDB schema
    role: normative
    status: modeled
    span:
      start: 0
      end: 59
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-164#1
  recorded_at: '2026-11-02T00:57:15.297Z'
id: REQ-identity-authz-spicedb-schema
type: req
---
The authorization model must be written as a SpiceDB schema.

## Context

The author recorded it on 2026-11-01, when planning viewers for the second alpha. Reason not stated. SpiceDB, by Authzed, is an open-source authorization database modeled on Google's Zanzibar. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> The authorization model is written as a SpiceDB schema, whose assertions CI checks. From the second alpha, SpiceDB answers the interface (D-163), with its datastore a database in the core's PostgreSQL cluster (D-14), and script tokens and view plugins are granted accounts as viewers are (D-65).

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-164#1
