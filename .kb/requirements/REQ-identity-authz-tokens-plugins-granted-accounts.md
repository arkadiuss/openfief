---
title: From the second alpha, script tokens and view plugins must be granted accounts as viewers are.
status: open
text_ref: design-decisions:D-164#4
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-164-4
  - lane:strict
semantic_text: From the second alpha, script tokens and view plugins must be granted accounts as viewers are.
logic_claims:
  - CLAIM-1D71B76D17469795
semantic_clauses:
  - From the second alpha, script tokens and view plugins must be granted accounts as viewers are.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 9508bb02404ee3c38eb98beb63d0af153eeadb44bb4cab1a315e1f39f85f99e2
semantic_inventory:
  - claim_key: CLAIM-1D71B76D17469795
    claim_text: From the second alpha, script tokens and view plugins must be granted accounts as viewers are
    role: normative
    status: modeled
    span:
      start: 0
      end: 93
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-164#4
  recorded_at: '2026-11-02T00:57:32.439Z'
id: REQ-identity-authz-tokens-plugins-granted-accounts
type: req
---
From the second alpha, script tokens and view plugins must be granted accounts as viewers are.

## Context

The author recorded it on 2026-11-01, when planning viewers for the second alpha. Reason not stated. D-65's viewers are granted the accounts they may view; script tokens (D-27) and view plugins get the same grants then. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> The authorization model is written as a SpiceDB schema, whose assertions CI checks. From the second alpha, SpiceDB answers the interface (D-163), with its datastore a database in the core's PostgreSQL cluster (D-14), and script tokens and view plugins are granted accounts as viewers are (D-65).

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-164#4
