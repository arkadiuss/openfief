---
title: A sync the core starts, on its schedule or with core connector sync, must write one audit entry of the system actor naming the connector instance, the import batch, the accounts it wrote transactions in and how many transaction versions, prices and bond terms it wrote, never an amount, and only when it changed state, in the transaction of its writes, while a sync that brought nothing new leaves only its attempt record.
status: open
firmness: default
required_evidence: unverified
origin:
  kind: agent
  ref: the S9 brief, from D-211, D-198 and D-67, 2026-11-14
  recorded_at: '2026-11-14T04:22:35.998Z'
semantic_text: A sync the core starts, on its schedule or with core connector sync, must write one audit entry of the system actor naming the connector instance, the import batch, the accounts it wrote transactions in and how many transaction versions, prices and bond terms it wrote, never an amount, and only when it changed state, in the transaction of its writes, while a sync that brought nothing new leaves only its attempt record.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: fd4ecb70d98c484e62288227111fba53fd00021896a3b0512bc48cb82583a4a6
semantic_inventory:
  - claim_key: CLAIM-3F3C97F56C1D59F6
    claim_text: A sync the core starts, on its schedule or with core connector sync, must write one audit entry of the system actor naming the connector instance, the import batch, the accounts it wrote transactions in and how many transaction versions, prices and bond terms it wrote, never an amount, and only when it changed state, in the transaction of its writes, while a sync that brought nothing new leaves only its attempt record
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 421
logic_claims:
  - CLAIM-3F3C97F56C1D59F6
id: REQ-identity-audit-sync-entries
type: req
---
A sync the core starts, on its schedule or with core connector sync, must write one audit entry of the system actor naming the connector instance, the import batch, the accounts it wrote transactions in and how many transaction versions, prices and bond terms it wrote, never an amount, and only when it changed state, in the transaction of its writes, while a sync that brought nothing new leaves only its attempt record.

## Context

D-198 keeps every attempt of a sync, with its error class and never its text; the author asked on 2026-11-14 for a sync that changed state to be an entry too. A price or a bond term counts when it is new or differs from the stored one (the stores report the rows an upsert changed), so a connector that sends the same prices every sync adds no entry. The entry carries counts and IDs only (GR-12).

## Source

> An append-only audit log records every change of state and every access event: sign-ins and sign-outs, passkeys, enrollment links, script tokens, grants, settings, the owner's entries and corrections, decisions on attention items, plugin installs and upgrades the operator applies, and model and MCP calls. Each entry names who, what, which records by ID, when and the outcome, never an amount, a secret or a question's text (GR-12). Settings shows it to the owner, filterable.

Source: the S9 brief, from D-211, 2026-11-14