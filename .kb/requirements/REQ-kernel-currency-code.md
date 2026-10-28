---
title: A currency must be a three-letter upper-case ISO 4217 alphabetic code, and any other code must be rejected.
status: open
text_ref: docs/dev/code-organization.md#bounded-contexts
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-code-organization-md-bounded-contexts
  - lane:strict
semantic_text: A currency must be a three-letter upper-case ISO 4217 alphabetic code, and any other code must be rejected.
logic_claims:
  - CLAIM-7C3AA9BF247BDE73
semantic_clauses:
  - A currency must be a three-letter upper-case ISO 4217 alphabetic code, and any other code must be rejected.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 107a81c97ab4d8de2b75fcc55b1c2d3d08acdb6559c04ce2674d3c86053d4680
semantic_inventory:
  - claim_key: CLAIM-7C3AA9BF247BDE73
    claim_text: A currency must be a three-letter upper-case ISO 4217 alphabetic code, and any other code must be rejected
    role: normative
    status: modeled
    span:
      start: 0
      end: 106
origin:
  kind: agent
  recorded_at: '2026-10-28T04:48:52.212Z'
id: REQ-kernel-currency-code
type: req
---
A currency must be a three-letter upper-case ISO 4217 alphabetic code, and any other code must be rejected.

## Context

Every amount carries its currency (D-20), so the currency must be a real code rather than free text from a broker file; brokers and NBP table A use ISO 4217 alphabetic codes. Only the format is checked, not membership in the current ISO list, so a new or historic code in a broker statement does not stop an import. ISO 4217 was named in the orchestrating agent's task; reason otherwise not stated. Modeled by Claude Code for the M0 foundation task (shared kernel); the author has not reviewed it yet.

## Source

> A tiny shared kernel `internal/kernel` holds money, currency, quantity and date types.

docs/dev/code-organization.md#bounded-contexts

> Every amount keeps its original currency.

docs/DESIGN.md D-20
