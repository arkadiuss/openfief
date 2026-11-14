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

The author decided it with D-211, and chose in the design session that a change and its entry are written together, so a change whose entry can't be written fails. Built: the identity events (sign-ins and sign-outs, passkeys, enrollment links, script tokens, grants, viewers), written in the same transaction as the change; model calls and the maintainer's runs and decisions; and, since S8 and S9, the changes of every other context, which enqueue their entry as a River job in the transaction of the change itself, to be written to the log once by a worker with the change's own time (REQ-identity-audit-outbox): the owner's changes (REQ-identity-audit-owner-changes), the models' settings, a sync the core started that changed state as the system (REQ-identity-audit-sync-entries), the administrator's changes at the command line as the admin actor (REQ-identity-audit-admin-actor), and the operator's plugin installs (REQ-identity-audit-plugin-installs). MCP calls are not built, since D-213 is not, so the requirement stays open. Read API calls are not entries.

## Source

> An append-only audit log records every change of state and every access event: sign-ins and sign-outs, passkeys, enrollment links, script tokens, grants, settings, the owner's entries and corrections, decisions on attention items, plugin installs and upgrades the operator applies, and model and MCP calls.

OpenFief author decisions (docs/DESIGN.md), design-decisions:D-211#2