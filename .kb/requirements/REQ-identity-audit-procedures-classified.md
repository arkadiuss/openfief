---
title: Every procedure of an owner service that changes something must write an audit entry, and every one that does not must be listed with the reason, so that a procedure added without an entry fails a test.
status: open
firmness: default
required_evidence: unverified
origin:
  kind: agent
  ref: the S8 brief, from D-211, 2026-11-14
  recorded_at: '2026-11-14T02:39:50.520Z'
semantic_text: Every procedure of an owner service that changes something must write an audit entry, and every one that does not must be listed with the reason, so that a procedure added without an entry fails a test.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: c9ebd8764456ce6de827676e65e3ef498d01199f762f9915a5821e859a76aa2c
semantic_inventory:
  - claim_key: CLAIM-8CBA9F91DC2BE866
    claim_text: Every procedure of an owner service that changes something must write an audit entry
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 84
  - claim_key: CLAIM-9914D28F94F851A7
    claim_text: every one that does not must be listed with the reason, so that a procedure added without an entry fails a test
    role: normative
    status: ontology_gap
    span:
      start: 90
      end: 201
logic_claims:
  - CLAIM-8CBA9F91DC2BE866
  - CLAIM-9914D28F94F851A7
id: REQ-identity-audit-procedures-classified
type: req
---
Every procedure of an owner service that changes something must write an audit entry, and every one that does not must be listed with the reason, so that a procedure added without an entry fails a test.

## Context

D-211 asks for every change of state. A test lists the procedures of every owner service from the contracts and fails for one that is in no list of audited, covered, unaudited-by-design and read-only procedures, and drives each change through a real core to check for exactly one entry.

## Source

> An append-only audit log records every change of state and every access event: sign-ins and sign-outs, passkeys, enrollment links, script tokens, grants, settings, the owner's entries and corrections, decisions on attention items, plugin installs and upgrades the operator applies, and model and MCP calls. Each entry names who, what, which records by ID, when and the outcome, never an amount, a secret or a question's text (GR-12). Settings shows it to the owner, filterable.

Source: the S8 brief, from D-211, 2026-11-14