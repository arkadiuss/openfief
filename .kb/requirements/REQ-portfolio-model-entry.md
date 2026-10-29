---
title: The core must let only an owner-scoped session on an allowed origin read or replace the model portfolio, through a service it serves to view plugins only, refusing a read session, another origin, an invalid session and a model that breaks a rule or names a malformed ISIN, and storing nothing on a refusal.
status: open
text_ref: docs/DESIGN.md#D-52
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-52
  - lane:strict
semantic_text: The core must let only an owner-scoped session on an allowed origin read or replace the model portfolio, through a service it serves to view plugins only, refusing a read session, another origin, an invalid session and a model that breaks a rule or names a malformed ISIN, and storing nothing on a refusal.
logic_claims:
  - CLAIM-AEA4AB0E9E637834
semantic_clauses:
  - The core must let only an owner-scoped session on an allowed origin read or replace the model portfolio, through a service it serves to view plugins only, refusing a read session, another origin, an invalid session and a model that breaks a rule or names a malformed ISIN, and storing nothing on a refusal.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 480ea1114472498448430def49f86a20fa4880e0411685e75974664391a17007
semantic_inventory:
  - claim_key: CLAIM-AEA4AB0E9E637834
    claim_text: The core must let only an owner-scoped session on an allowed origin read or replace the model portfolio, through a service it serves to view plugins only, refusing a read session, another origin, an invalid session and a model that breaks a rule or names a malformed ISIN, and storing nothing on a refusal
    role: normative
    status: modeled
    span:
      start: 0
      end: 305
origin:
  kind: agent
  recorded_at: '2026-10-28T23:10:31.811Z'
id: REQ-portfolio-model-entry
type: req
---
The core must let only an owner-scoped session on an allowed origin read or replace the model portfolio, through a service it serves to view plugins only, refusing a read session, another origin, an invalid session and a model that breaks a rule or names a malformed ISIN, and storing nothing on a refusal.

## Context

D-52 has the owner enter the model in the built-in UI, which hands it to the core like manual bond lots (D-49); the read API stays read-only (D-6). The separate view-plugin service with the AuthService's caller check and owner scope is this context's narrowest reading, parked for the author. Modeled by Claude Code for the model portfolio task; the author has not reviewed it yet.

## Source

> **[D-52]** The model portfolio is optional and off by default; while it is off, it appears nowhere in the UI or the read API. It is a weighted basket for the whole portfolio: each target is an instrument (such as an index ETF or a stock) or an index, with a weight, and the weights sum to 100%. The owner enters it in the built-in UI, which hands it to the core like manual bond lots (D-49). When it is on, the read API serves it with its backtest and drawdowns.

docs/DESIGN.md#D-52
