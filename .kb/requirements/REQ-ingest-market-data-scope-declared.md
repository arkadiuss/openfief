---
title: The core must ask every market-data connector instance, one whose manifest declares it and whose grant approves that, to price each instrument held today in any account, other than treasury bonds with stored series terms, and the model portfolio's instrument and index targets while it is on, each once, and must ask every other instance for nothing.
status: open
text_ref: docs/DESIGN.md#D-123
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-123
  - lane:strict
semantic_text: The core must ask every market-data connector instance, one whose manifest declares it and whose grant approves that, to price each instrument held today in any account, other than treasury bonds with stored series terms, and the model portfolio's instrument and index targets while it is on, each once, and must ask every other instance for nothing.
logic_claims:
  - CLAIM-34AFDD3C89B5B844
semantic_clauses:
  - The core must ask every market-data connector instance, one whose manifest declares it and whose grant approves that, to price each instrument held today in any account, other than treasury bonds with stored series terms, and the model portfolio's instrument and index targets while it is on, each once, and must ask every other instance for nothing.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 38e6415b484b25185d8d7111f83a4935ed2ecce757f7574321804aa580b9999d
semantic_inventory:
  - claim_key: CLAIM-34AFDD3C89B5B844
    claim_text: The core must ask every market-data connector instance, one whose manifest declares it and whose grant approves that, to price each instrument held today in any account, other than treasury bonds with stored series terms, and the model portfolio's instrument and index targets while it is on, each once, and must ask every other instance for nothing
    role: normative
    status: modeled
    span:
      start: 0
      end: 349
origin:
  kind: agent
  recorded_at: '2026-11-01T10:07:06.316Z'
id: REQ-ingest-market-data-scope-declared
type: req
---
The core must ask every market-data connector instance, one whose manifest declares it and whose grant approves that, to price each instrument held today in any account, other than treasury bonds with stored series terms, and the model portfolio's instrument and index targets while it is on, each once, and must ask every other instance for nothing.

## Context

D-93 has each sync of a market-data connector carry what the owner holds and the model's targets; D-123 makes a connector one only when its manifest declares it and its grant approves that, so the core no longer treats an instance bound to no account as one. Replaces REQ-ingest-market-data-scope. Modeled by Claude Code for the read-access work package (impl/read-access); the author has not reviewed it yet.

## Source

> **[D-123]** Only a connector whose manifest declares it a market-data connector, and whose grant approves that, receives D-93's list of held instruments and model targets, so the install shows the owner that the plugin will learn what they hold.

docs/DESIGN.md#D-123
