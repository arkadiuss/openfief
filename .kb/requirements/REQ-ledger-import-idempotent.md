---
title: Re-importing a transaction whose source and institution ID are already in the log with the same content as its current version must change nothing, and content equality must not depend on posting order or on how a number is written.
status: open
text_ref: docs/DESIGN.md#D-18
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-18
  - lane:strict
semantic_text: Re-importing a transaction whose source and institution ID are already in the log with the same content as its current version must change nothing, and content equality must not depend on posting order or on how a number is written.
logic_claims:
  - CLAIM-5913080FD43262D2
semantic_clauses:
  - Re-importing a transaction whose source and institution ID are already in the log with the same content as its current version must change nothing, and content equality must not depend on posting order or on how a number is written.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: f6c8e60d63cb47707e1403aef00f39d6a5ec26833864e235e280c0efe4cf6848
semantic_inventory:
  - claim_key: CLAIM-5913080FD43262D2
    claim_text: Re-importing a transaction whose source and institution ID are already in the log with the same content as its current version must change nothing, and content equality must not depend on posting order or on how a number is written
    role: normative
    status: modeled
    span:
      start: 0
      end: 231
origin:
  kind: agent
  recorded_at: '2026-10-28T05:32:20.421Z'
id: REQ-ledger-import-idempotent
type: req
---
Re-importing a transaction whose source and institution ID are already in the log with the same content as its current version must change nothing, and content equality must not depend on posting order or on how a number is written.

## Context

Refines the firm requirement REQ-AUTO-F19F621B67F14F4F (D-18): re-imports are idempotent, keyed by source plus the institution's own ID. Connectors may re-send overlapping periods (a daily Flex pull, a re-uploaded XTB export), so the same row must not create duplicates or versions. Equality is a canonical SHA-256 fingerprint over every field, with postings sorted and decimals in canonical form, so it is deterministic and storable. Modeled by Claude Code for the M0 ledger task; the author has not reviewed it yet.

## Source

> Re-imports are idempotent, keyed by source plus the institution's own ID; a changed row creates a new version and supersedes the old one, which is never deleted.

docs/DESIGN.md D-18
