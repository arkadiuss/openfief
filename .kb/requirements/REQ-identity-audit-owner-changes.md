---
title: Every change of state the owner makes through an owner service in the ledger, ingest, valuation, portfolio and attention contexts must be written to the audit log, once, naming the owner's session and the records changed by ID and never an amount, a name, a secret or a question's text, so that a change the log can't record is not made.
status: open
firmness: default
required_evidence: unverified
origin:
  kind: agent
  ref: the S8 brief, from D-211, 2026-11-14
  recorded_at: '2026-11-14T02:39:48.059Z'
semantic_text: Every change of state the owner makes through an owner service in the ledger, ingest, valuation, portfolio and attention contexts must be written to the audit log, once, naming the owner's session and the records changed by ID and never an amount, a name, a secret or a question's text, so that a change the log can't record is not made.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 8187b310fcbef3a59569aae68b73bcd763d47caf8e7c34dabc26ed8cfe070306
semantic_inventory:
  - claim_key: CLAIM-03884D298DACB13C
    claim_text: Every change of state the owner makes through an owner service in the ledger, ingest, valuation, portfolio and attention contexts must be written to the audit log, once, naming the owner's session and the records changed by ID and never an amount, a name, a secret or a question's text, so that a change the log can't record is not made
    role: normative
    status: missing
    span:
      start: 0
      end: 336
logic_claims:
  - CLAIM-03884D298DACB13C
id: REQ-identity-audit-owner-changes
type: req
---
Every change of state the owner makes through an owner service in the ledger, ingest, valuation, portfolio and attention contexts must be written to the audit log, once, naming the owner's session and the records changed by ID and never an amount, a name, a secret or a question's text, so that a change the log can't record is not made.

## Context

D-211 lists the owner's entries and corrections, settings and decisions on attention items. Each change is entered through the identity context's RecordOwnerChange, which since S9 enqueues the entry in the outbox (REQ-identity-audit-outbox): the change runs in one transaction with its stores joined and the entry is a job inserted in it, so a change whose entry can't be enqueued is not made, and a change that fails leaves no entry. A worker writes the entry to the log afterwards, once, with the change's time; a log that is down delays it and loses none.

## Source

> An append-only audit log records every change of state and every access event: sign-ins and sign-outs, passkeys, enrollment links, script tokens, grants, settings, the owner's entries and corrections, decisions on attention items, plugin installs and upgrades the operator applies, and model and MCP calls. Each entry names who, what, which records by ID, when and the outcome, never an amount, a secret or a question's text (GR-12). Settings shows it to the owner, filterable.

Source: the S8 brief, from D-211, 2026-11-14