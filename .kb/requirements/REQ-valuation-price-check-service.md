---
title: The core must serve PriceCheckService on its plugin port to view plugins only, for an owner session, setting or clearing one instrument's threshold as a fraction greater than 0 and less than 1 and acknowledging a flagged check, and must answer each refusal with its own Connect code.
status: open
text_ref: docs/DESIGN.md#D-79
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-79
  - lane:strict
semantic_text: The core must serve PriceCheckService on its plugin port to view plugins only, for an owner session, setting or clearing one instrument's threshold as a fraction greater than 0 and less than 1 and acknowledging a flagged check, and must answer each refusal with its own Connect code.
logic_claims:
  - CLAIM-0977C2379BB998F9
semantic_clauses:
  - The core must serve PriceCheckService on its plugin port to view plugins only, for an owner session, setting or clearing one instrument's threshold as a fraction greater than 0 and less than 1 and acknowledging a flagged check, and must answer each refusal with its own Connect code.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: bf4ed92ff8bf9c2ad08af287baf6b1d62225a8a0360e7d19aff8d48b47d5ce54
semantic_inventory:
  - claim_key: CLAIM-0977C2379BB998F9
    claim_text: The core must serve PriceCheckService on its plugin port to view plugins only, for an owner session, setting or clearing one instrument's threshold as a fraction greater than 0 and less than 1 and acknowledging a flagged check, and must answer each refusal with its own Connect code
    role: normative
    status: modeled
    span:
      start: 0
      end: 282
origin:
  kind: agent
  recorded_at: '2026-10-30T02:59:00.179Z'
id: REQ-valuation-price-check-service
type: req
---
The core must serve PriceCheckService on its plugin port to view plugins only, for an owner session, setting or clearing one instrument's threshold as a fraction greater than 0 and less than 1 and acknowledging a flagged check, and must answer each refusal with its own Connect code.

## Context

D-110 and D-112 give the owner a per-instrument threshold and acknowledgements; D-79 has each kind of hand entry reach the core through its own owner-only service, and D-115 serves those on the plugin port.

## Source

> What the owner enters by hand reaches the core through one service per kind of entry: bond lots, the model portfolio, accounts, statement uploads and any later kind. Each is owner-only, versioned on its own, and called only by a view plugin acting for an owner session (D-69).

docs/DESIGN.md#D-79
