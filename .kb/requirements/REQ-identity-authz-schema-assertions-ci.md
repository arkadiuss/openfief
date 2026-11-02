---
title: CI must check the authorization schema's assertions.
status: open
text_ref: design-decisions:D-164#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-164-2
  - lane:strict
semantic_text: CI must check the authorization schema's assertions.
logic_claims:
  - CLAIM-83AA735081C06821
semantic_clauses:
  - CI must check the authorization schema's assertions.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 36677af512cca1ad8df1b639c8b11455f24576f3f675ede9015d6f0ced66eff0
semantic_inventory:
  - claim_key: CLAIM-83AA735081C06821
    claim_text: CI must check the authorization schema's assertions
    role: normative
    status: modeled
    span:
      start: 0
      end: 51
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-164#2
  recorded_at: '2026-11-02T00:57:21.005Z'
id: REQ-identity-authz-schema-assertions-ci
type: req
---
CI must check the authorization schema's assertions.

## Context

The author recorded it on 2026-11-01, when planning viewers for the second alpha. Reason not stated. zed, SpiceDB's command-line tool, validates a schema against relationships and assertions offline. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> The authorization model is written as a SpiceDB schema, whose assertions CI checks. From the second alpha, SpiceDB answers the interface (D-163), with its datastore a database in the core's PostgreSQL cluster (D-14), and script tokens and view plugins are granted accounts as viewers are (D-65).

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-164#2
