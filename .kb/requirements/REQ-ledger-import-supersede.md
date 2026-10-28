---
title: Importing changed content for a known source and institution ID must append the next version, superseding the current one, and earlier versions must never be removed or changed.
status: open
text_ref: docs/DESIGN.md#D-18
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-18
  - lane:strict
semantic_text: Importing changed content for a known source and institution ID must append the next version, superseding the current one, and earlier versions must never be removed or changed.
logic_claims:
  - CLAIM-D6D64A687DAA5728
semantic_clauses:
  - Importing changed content for a known source and institution ID must append the next version, superseding the current one, and earlier versions must never be removed or changed.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 0582190c9412b2151bb77729d0b98c356688e960bbfb7816146531028667b3e6
semantic_inventory:
  - claim_key: CLAIM-D6D64A687DAA5728
    claim_text: Importing changed content for a known source and institution ID must append the next version, superseding the current one, and earlier versions must never be removed or changed
    role: normative
    status: modeled
    span:
      start: 0
      end: 176
origin:
  kind: agent
  recorded_at: '2026-10-28T05:32:25.781Z'
id: REQ-ledger-import-supersede
type: req
---
Importing changed content for a known source and institution ID must append the next version, superseding the current one, and earlier versions must never be removed or changed.

## Context

Refines the firm requirements REQ-AUTO-22E70B6530E994B1 and REQ-AUTO-CACBF41A8E14837D (D-18): the log is append-only and versioned, and a changed row supersedes the old one, which is never deleted. Only the current version is compared, so a row that changes back to an earlier content is still a new version, which keeps the history of what the broker reported. Modeled by Claude Code for the M0 ledger task; the author has not reviewed it yet.

## Source

> The source of truth is an append-only, versioned transaction log. ... a changed row creates a new version and supersedes the old one, which is never deleted.

docs/DESIGN.md D-18
