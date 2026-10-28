---
title: The identity context must keep script tokens in its own PostgreSQL schema, migrated by its own runner and lock, with a row per token holding its ID, name, scope, times and the 32-byte SHA-256 hash of its key, and must refuse rows with a malformed ID, name or hash or an expiry not after creation.
status: open
text_ref: docs/DESIGN.md#D-14
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-14
  - lane:strict
semantic_text: The identity context must keep script tokens in its own PostgreSQL schema, migrated by its own runner and lock, with a row per token holding its ID, name, scope, times and the 32-byte SHA-256 hash of its key, and must refuse rows with a malformed ID, name or hash or an expiry not after creation.
logic_claims:
  - CLAIM-95DD9512447C8C06
semantic_clauses:
  - The identity context must keep script tokens in its own PostgreSQL schema, migrated by its own runner and lock, with a row per token holding its ID, name, scope, times and the 32-byte SHA-256 hash of its key, and must refuse rows with a malformed ID, name or hash or an expiry not after creation.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 9543af84affd9c056c3ae470c0e3ef072faa1d8915d1ec1987b29d38c3b1ca24
semantic_inventory:
  - claim_key: CLAIM-95DD9512447C8C06
    claim_text: The identity context must keep script tokens in its own PostgreSQL schema, migrated by its own runner and lock, with a row per token holding its ID, name, scope, times and the 32-byte SHA-256 hash of its key, and must refuse rows with a malformed ID, name or hash or an expiry not after creation
    role: normative
    status: modeled
    span:
      start: 0
      end: 295
origin:
  kind: agent
  recorded_at: '2026-10-28T14:11:37.635Z'
id: REQ-identity-script-token-store
type: req
---
The identity context must keep script tokens in its own PostgreSQL schema, migrated by its own runner and lock, with a row per token holding its ID, name, scope, times and the 32-byte SHA-256 hash of its key, and must refuse rows with a malformed ID, name or hash or an expiry not after creation.

## Context

D-14 puts state in PostgreSQL; each context owns its schema (docs/dev/code-organization.md), so identity copies the ledger's migration runner with its own version table and lock key. Database checks back up the domain's validation. Modeled by Claude Code for the core task: the author asked for minimal script tokens so the read API can run on the demo; the author has not reviewed this requirement yet.

## Source

> PostgreSQL via CloudNativePG, and no NATS in M0.

docs/DESIGN.md D-14
