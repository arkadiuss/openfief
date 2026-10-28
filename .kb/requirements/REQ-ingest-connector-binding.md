---
title: A connector instance's delivery must be refused whole, before anything is imported, when any of its transactions or balance assertions is in an account the instance has no connector binding to, and a connector binding must only be added for a stored account and a valid instance name.
status: open
text_ref: docs/DESIGN.md#D-48
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-48
  - lane:strict
semantic_text: A connector instance's delivery must be refused whole, before anything is imported, when any of its transactions or balance assertions is in an account the instance has no connector binding to, and a connector binding must only be added for a stored account and a valid instance name.
logic_claims:
  - CLAIM-11CFCF371D6CF259
semantic_clauses:
  - A connector instance's delivery must be refused whole, before anything is imported, when any of its transactions or balance assertions is in an account the instance has no connector binding to, and a connector binding must only be added for a stored account and a valid instance name.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 4b08b4f88121eb6faed5d044198b1fa66f79f158202f6f62a0a6feccfbfae9c0
semantic_inventory:
  - claim_key: CLAIM-11CFCF371D6CF259
    claim_text: A connector instance's delivery must be refused whole, before anything is imported, when any of its transactions or balance assertions is in an account the instance has no connector binding to, and a connector binding must only be added for a stored account and a valid instance name
    role: normative
    status: modeled
    span:
      start: 0
      end: 283
origin:
  kind: agent
  recorded_at: '2026-10-28T16:50:24.282Z'
id: REQ-ingest-connector-binding
type: req
---
A connector instance's delivery must be refused whole, before anything is imported, when any of its transactions or balance assertions is in an account the instance has no connector binding to, and a connector binding must only be added for a stored account and a valid instance name.

## Context

Key scoping (D-48) keeps one instance from overwriting another's transactions, but a connector could still name any account. The orchestrator asked for a minimal binding of instance to allowed accounts, managed by the owner with core connector bind, so a compromised or misconfigured connector can't write into accounts it doesn't serve. Modeled by Claude Code for the core-connectors task (orchestrator brief: the core scheduler, watermarks, connector bindings); the author has not reviewed this requirement yet.

## Source

> Each install is a separately named instance with its own config, secrets, grant and egress allow-list.

docs/DESIGN.md D-48
