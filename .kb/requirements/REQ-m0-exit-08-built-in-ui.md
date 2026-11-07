---
title: Passkey login must work on localhost, a passkey created for a second hostname through a one-time link must work there, and the allocation and value-over-time charts must be drawn using only the read API.
status: open
text_ref: design-decisions:D-200#8
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-200-8
  - lane:strict
semantic_text: Passkey login must work on localhost, a passkey created for a second hostname through a one-time link must work there, and the allocation and value-over-time charts must be drawn using only the read API.
logic_claims:
  - CLAIM-89B4CDB0EDD5FA63
semantic_clauses:
  - Passkey login must work on localhost, a passkey created for a second hostname through a one-time link must work there, and the allocation and value-over-time charts must be drawn using only the read API.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 8585789a1e1124c9eaba4b6efffed338e3bb689af162071808200e97aab4e26f
semantic_inventory:
  - claim_key: CLAIM-89B4CDB0EDD5FA63
    claim_text: Passkey login must work on localhost, a passkey created for a second hostname through a one-time link must work there, and the allocation and value-over-time charts must be drawn using only the read API
    role: normative
    status: modeled
    span:
      start: 0
      end: 202
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-200#8
  recorded_at: '2026-11-07T03:31:59.786Z'
id: REQ-m0-exit-08-built-in-ui
type: req
---
Passkey login must work on localhost, a passkey created for a second hostname through a one-time link must work there, and the allocation and value-over-time charts must be drawn using only the read API.

## Context

Author decision D-200: M0 is done when every one of its nineteen criteria holds. Each criterion is an end-to-end scenario on the firm requirement it checks, and the whole set passes on both kind and k3d (criterion 15); the scenarios that verify this criterion carry the tag acceptance:p10-8.

## Source

> **Built-in UI.** Passkey login works on `localhost`, and a passkey created for a second hostname through a one-time link works there. The allocation and value-over-time charts are drawn using only the read API.

design-decisions:D-200#8 (docs/DESIGN.md, D-200, criterion 8)
