---
title: Every account must belong to exactly one sleeve, and an account without a sleeve, ID, institution or name must be rejected.
status: open
text_ref: docs/DESIGN.md#33-ledger-and-valuation
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-33-ledger-and-valuation
  - lane:strict
semantic_text: Every account must belong to exactly one sleeve, and an account without a sleeve, ID, institution or name must be rejected.
logic_claims:
  - CLAIM-DF8F1E0EF8D711C2
semantic_clauses:
  - Every account must belong to exactly one sleeve, and an account without a sleeve, ID, institution or name must be rejected.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 585fd4a46d0e76b3251ace654df19b321dbebc31232e3fbb5de0f6d1aeb11bfb
semantic_inventory:
  - claim_key: CLAIM-DF8F1E0EF8D711C2
    claim_text: Every account must belong to exactly one sleeve, and an account without a sleeve, ID, institution or name must be rejected
    role: normative
    status: modeled
    span:
      start: 0
      end: 122
origin:
  kind: agent
  recorded_at: '2026-10-28T05:24:28.107Z'
id: REQ-ledger-account-sleeve
type: req
---
Every account must belong to exactly one sleeve, and an account without a sleeve, ID, institution or name must be rejected.

## Context

The ledger diagram in DESIGN section 3.3 has every sleeve backed by accounts, and the glossary says an account belongs to one sleeve. Allocation by sleeve (D-30) and the later lease ledger need every account to have one. The ledger refers to the sleeve by its ID only and does not depend on the ingest context for the institution. Modeled by Claude Code for the M0 ledger task; the author has not reviewed it yet.

## Source

> SLEEVE ||--|{ ACCOUNT : "backed by"

docs/DESIGN.md section 3.3

> One account at an institution, such as one IBKR linked account. Belongs to one sleeve.

docs/glossary.md (Account)
