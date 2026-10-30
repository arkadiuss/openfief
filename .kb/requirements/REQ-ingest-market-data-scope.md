---
title: The core must ask every connector instance bound to no account to price each instrument held today in any account, other than treasury bonds with stored series terms, and the model portfolio's instrument and index targets while it is on, each once, and must ask an instance bound to an account for nothing.
status: open
text_ref: docs/DESIGN.md#D-93
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-93
  - lane:strict
semantic_text: The core must ask every connector instance bound to no account to price each instrument held today in any account, other than treasury bonds with stored series terms, and the model portfolio's instrument and index targets while it is on, each once, and must ask an instance bound to an account for nothing.
logic_claims:
  - CLAIM-EF14AC4F99088435
semantic_clauses:
  - The core must ask every connector instance bound to no account to price each instrument held today in any account, other than treasury bonds with stored series terms, and the model portfolio's instrument and index targets while it is on, each once, and must ask an instance bound to an account for nothing.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 210bb1f1229ac3ac7f47814f18f10b0127048e76627925496a1f540a68cc43f4
semantic_inventory:
  - claim_key: CLAIM-EF14AC4F99088435
    claim_text: The core must ask every connector instance bound to no account to price each instrument held today in any account, other than treasury bonds with stored series terms, and the model portfolio's instrument and index targets while it is on, each once, and must ask an instance bound to an account for nothing
    role: normative
    status: modeled
    span:
      start: 0
      end: 305
origin:
  kind: agent
  recorded_at: '2026-10-29T16:58:01.770Z'
id: REQ-ingest-market-data-scope
type: req
---
The core must ask every connector instance bound to no account to price each instrument held today in any account, other than treasury bonds with stored series terms, and the model portfolio's instrument and index targets while it is on, each once, and must ask an instance bound to an account for nothing.

## Context

D-93 has each sync request to a market-data connector carry what is held and what the model portfolio targets. No manifest field marks a market-data connector, so the narrowest reading treats an instance that feeds no account as one, and a broker connector never learns what is held elsewhere (implementation questions). Bonds have no market price (D-9). Modeled by Claude Code for the market-data work package (D-93, D-94, D-95, impl/market-data); the author has not reviewed it yet.

## Source

> **[D-93]** Each sync request to a market-data connector carries the instruments the owner holds and the model portfolio's targets, so a newly bought instrument is priced without configuration. An optional list in the instance's config adds instruments to watch.

docs/DESIGN.md#D-93
