---
title: A change of state in the ledger, ingest, valuation, portfolio, attention or models context must enqueue the audit entry of it as a job in the transaction of the change itself, and a worker must write it to the audit log once, keyed by the job, with the time of the change, so that the change and its pending entry commit or roll back together, no transaction is held open across two contexts' tables, and a change whose entry can't be enqueued is not made.
status: open
firmness: default
required_evidence: unverified
origin:
  kind: agent
  ref: the S9 brief, from D-211, D-198 and D-67, 2026-11-14
  recorded_at: '2026-11-14T04:22:34.246Z'
semantic_text: A change of state in the ledger, ingest, valuation, portfolio, attention or models context must enqueue the audit entry of it as a job in the transaction of the change itself, and a worker must write it to the audit log once, keyed by the job, with the time of the change, so that the change and its pending entry commit or roll back together, no transaction is held open across two contexts' tables, and a change whose entry can't be enqueued is not made.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 49899f8732764bc5831a8bae763c0ad35670b02fd1bc1e3b9431b7f93333ba5c
semantic_inventory:
  - claim_key: CLAIM-361BFCB1544273C7
    claim_text: A change of state in the ledger, ingest, valuation, portfolio, attention or models context must enqueue the audit entry of it as a job in the transaction of the change itself, and a worker must write it to the audit log once, keyed by the job, with the time of the change, so that the change and its pending entry commit or roll back together, no transaction is held open across two contexts' tables, and a change whose entry can't be enqueued is not made
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 455
logic_claims:
  - CLAIM-361BFCB1544273C7
id: REQ-identity-audit-outbox
type: req
---
A change of state in the ledger, ingest, valuation, portfolio, attention or models context must enqueue the audit entry of it as a job in the transaction of the change itself, and a worker must write it to the audit log once, keyed by the job, with the time of the change, so that the change and its pending entry commit or roll back together, no transaction is held open across two contexts' tables, and a change whose entry can't be enqueued is not made.

## Context

The author's decision of 2026-11-14 replaces S8's seam, which kept the entry in a transaction of identity's own while the change ran on the other contexts' connections. The change's stores join one transaction carried in the context, the entry is a River job inserted in it (InsertTx), and a worker writes it to identity's log, idempotently by the job's ID (a unique column), with the change's own time. River retries a job the worker can't write for ever, so a log that is down delays entries and loses none. The identity context's own changes keep their direct write, which is in their own transaction already. A consequence of the one transaction: every store of a change shares one connection, so a statement that fails and is not handled inside a savepoint rolls the whole change back.

## Source

> An append-only audit log records every change of state and every access event: sign-ins and sign-outs, passkeys, enrollment links, script tokens, grants, settings, the owner's entries and corrections, decisions on attention items, plugin installs and upgrades the operator applies, and model and MCP calls. Each entry names who, what, which records by ID, when and the outcome, never an amount, a secret or a question's text (GR-12). Settings shows it to the owner, filterable.

Source: the S9 brief, from D-211, 2026-11-14