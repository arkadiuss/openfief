---
title: The audit log must record every change of state and every access event.
status: open
tags:
  - strict-modeling
  - lane:strict
  - provenance:design-decisions-d-211-2
text_ref: design-decisions:D-211#2
origin:
  kind: import
  ref: design-decisions:D-211#2
  recorded_at: '2026-11-12T19:18:16.905Z'
semantic_text: The audit log must record every change of state and every access event.
logic_claims:
  - CLAIM-C808F28A16CF0DAF
semantic_clauses:
  - The audit log must record every change of state and every access event.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 76bda3d68b663230d6c32dc1fc9044d5b5901518976a1dc3b8395be0cced9c11
semantic_inventory:
  - claim_key: CLAIM-C808F28A16CF0DAF
    claim_text: The audit log must record every change of state and every access event
    role: normative
    status: modeled
    span:
      start: 0
      end: 70
id: REQ-identity-audit-every-change
type: req
---
The audit log must record every change of state and every access event.

## Context

The author decided it with D-211, and chose in the design session that a change and its entry commit in one core-database transaction, so a change whose entry can't be written fails. The identity events are built: sign-ins and sign-outs, passkeys, enrollment links and script tokens. Grants, settings, the owner's entries, attention decisions, plugin installs and model and MCP calls are not, so the requirement stays open. Read API calls are not entries.

## Source

> An append-only audit log records every change of state and every access event: sign-ins and sign-outs, passkeys, enrollment links, script tokens, grants, settings, the owner's entries and corrections, decisions on attention items, plugin installs and upgrades the operator applies, and model and MCP calls.

OpenFief author decisions (docs/DESIGN.md), design-decisions:D-211#2