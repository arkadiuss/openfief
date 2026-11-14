---
title: The core must enter in the audit log, as the operator's, each install, upgrade and removal of a plugin it observes from the plugin's name and generation, once per generation and retried until it is written.
status: open
firmness: default
required_evidence: unverified
origin:
  kind: agent
  ref: the S8 brief, from D-211, 2026-11-14
  recorded_at: '2026-11-14T02:39:49.294Z'
semantic_text: The core must enter in the audit log, as the operator's, each install, upgrade and removal of a plugin it observes from the plugin's name and generation, once per generation and retried until it is written.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 17c5d7fbee039ac74091af9ed1c3d4030a767b96493e6f15f1cbf59ff6b33a06
semantic_inventory:
  - claim_key: CLAIM-AD6505C66793FB3E
    claim_text: The core must enter in the audit log, as the operator's, each install, upgrade and removal of a plugin it observes from the plugin's name and generation, once per generation and retried until it is written
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 205
logic_claims:
  - CLAIM-AD6505C66793FB3E
id: REQ-identity-audit-plugin-installs
type: req
---
The core must enter in the audit log, as the operator's, each install, upgrade and removal of a plugin it observes from the plugin's name and generation, once per generation and retried until it is written.

## Context

D-211 asks for the plugin installs and upgrades the operator applies. The operator is a separate process that can't be made to wait for an entry, so the core records what it observes in the Plugin resources, with the generation of each, in one transaction with the generation it remembers, which keeps the same observation from being entered twice. A failed write is retried.

## Source

> An append-only audit log records every change of state and every access event: sign-ins and sign-outs, passkeys, enrollment links, script tokens, grants, settings, the owner's entries and corrections, decisions on attention items, plugin installs and upgrades the operator applies, and model and MCP calls. Each entry names who, what, which records by ID, when and the outcome, never an amount, a secret or a question's text (GR-12). Settings shows it to the owner, filterable.

Source: the S8 brief, from D-211, 2026-11-14