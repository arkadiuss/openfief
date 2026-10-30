---
title: A withdrawal from one account and a deposit into another account of exactly the same amount and currency, dated at most five days apart in either order, must be matched as a transfer that is not an external cash flow, each withdrawal taking the nearest unmatched deposit, and opening cash must never be a leg.
status: closed
text_ref: docs/DESIGN.md#D-61
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-61
  - lane:strict
semantic_text: A withdrawal from one account and a deposit into another account of exactly the same amount and currency, dated at most five days apart in either order, must be matched as a transfer that is not an external cash flow, each withdrawal taking the nearest unmatched deposit, and opening cash must never be a leg.
logic_claims:
  - CLAIM-21A98D574F5D9E1B
semantic_clauses:
  - A withdrawal from one account and a deposit into another account of exactly the same amount and currency, dated at most five days apart in either order, must be matched as a transfer that is not an external cash flow, each withdrawal taking the nearest unmatched deposit, and opening cash must never be a leg.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 354bcc791e1402639a5359c85e197839f3f34cc599a84bc3dad66b479084ec26
semantic_inventory:
  - claim_key: CLAIM-21A98D574F5D9E1B
    claim_text: A withdrawal from one account and a deposit into another account of exactly the same amount and currency, dated at most five days apart in either order, must be matched as a transfer that is not an external cash flow, each withdrawal taking the nearest unmatched deposit, and opening cash must never be a leg
    role: normative
    status: modeled
    span:
      start: 0
      end: 308
origin:
  kind: agent
  recorded_at: '2026-10-28T20:55:07.137Z'
id: REQ-views-performance-transfers
type: req
---
A withdrawal from one account and a deposit into another account of exactly the same amount and currency, dated at most five days apart in either order, must be matched as a transfer that is not an external cash flow, each withdrawal taking the nearest unmatched deposit, and opening cash must never be a leg.

## Context

Closed: superseded by REQ-views-performance-confirmed-transfers. D-105 leaves out of the time-weighted return's cash flows only the transfers the owner confirmed, not every pair this matching rule finds; the rule now only suggests (REQ-views-transfer-suggestions).
