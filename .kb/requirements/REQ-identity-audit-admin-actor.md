---
title: A change of state made through the core's command line, such as core token issue and revoke (the break-glass way to a script token, D-67), core enroll-link and core connector bind, must be written to the audit log as the admin actor, whose ID is admin, and not as the core's own work.
status: open
firmness: default
required_evidence: unverified
origin:
  kind: agent
  ref: the S9 brief, from D-211, D-198 and D-67, 2026-11-14
  recorded_at: '2026-11-14T04:22:37.726Z'
semantic_text: A change of state made through the core's command line, such as core token issue and revoke (the break-glass way to a script token, D-67), core enroll-link and core connector bind, must be written to the audit log as the admin actor, whose ID is admin, and not as the core's own work.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 63a8cc62a6042d07f35d728d19418a7deea9e48aa9c5769f2cf3a68f9fb0a75f
semantic_inventory:
  - claim_key: CLAIM-C2DEA1E0493C661A
    claim_text: A change of state made through the core's command line, such as core token issue and revoke (the break-glass way to a script token, D-67), core enroll-link and core connector bind, must be written to the audit log as the admin actor, whose ID is admin, and not as the core's own work
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 283
logic_claims:
  - CLAIM-C2DEA1E0493C661A
id: REQ-identity-audit-admin-actor
type: req
---
A change of state made through the core's command line, such as core token issue and revoke (the break-glass way to a script token, D-67), core enroll-link and core connector bind, must be written to the audit log as the admin actor, whose ID is admin, and not as the core's own work.

## Context

The author decided on 2026-11-14 that the administrator at the command line is recorded as such: a person ran the command in the core's pod. The core's own work, such as its scheduled sync, the grants cutover and the sweeps, stays the system actor. core connector bind enters its account_bound entry through the outbox, in the transaction of the binding.

## Source

> An append-only audit log records every change of state and every access event: sign-ins and sign-outs, passkeys, enrollment links, script tokens, grants, settings, the owner's entries and corrections, decisions on attention items, plugin installs and upgrades the operator applies, and model and MCP calls. Each entry names who, what, which records by ID, when and the outcome, never an amount, a secret or a question's text (GR-12). Settings shows it to the owner, filterable.

Source: the S9 brief, from D-211, 2026-11-14