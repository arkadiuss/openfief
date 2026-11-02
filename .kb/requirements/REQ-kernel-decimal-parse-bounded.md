---
title: ParseDecimal must refuse text longer than 256 bytes or with a scientific exponent beyond plus or minus 100, before building the number, and its errors must never quote the text.
status: open
text_ref: docs/DESIGN.md#D-184
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-184
  - lane:strict
semantic_text: ParseDecimal must refuse text longer than 256 bytes or with a scientific exponent beyond plus or minus 100, before building the number, and its errors must never quote the text.
logic_claims:
  - CLAIM-374FFF6BEF80FC12
semantic_clauses:
  - ParseDecimal must refuse text longer than 256 bytes or with a scientific exponent beyond plus or minus 100, before building the number, and its errors must never quote the text.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 576e8aae4dfd762f9b955d70693ee1dd927e95273a7ffaa3cce74ffdd1f34dfc
semantic_inventory:
  - claim_key: CLAIM-374FFF6BEF80FC12
    claim_text: ParseDecimal must refuse text longer than 256 bytes or with a scientific exponent beyond plus or minus 100, before building the number, and its errors must never quote the text
    role: normative
    status: modeled
    span:
      start: 0
      end: 176
origin:
  kind: agent
  recorded_at: '2026-11-02T20:57:46.071Z'
id: REQ-kernel-decimal-parse-bounded
type: req
---
ParseDecimal must refuse text longer than 256 bytes or with a scientific exponent beyond plus or minus 100, before building the number, and its errors must never quote the text.

## Context

1E+2000000000 parsed and then made String or Round write out or allocate gigabytes. Every number from outside goes through ParseDecimal: statements such as XTB's number cells, NBP, EODHD and GUS answers and what the owner types; spreadsheet exponents such as 1.5E-7 stay well within the bound. The text can be an amount, so errors name only the reason (GR-12). Recorded with default firmness by the implementation agent of the alpha security fixes (fix/p0-dos), from the security review's finding P0-3. The author made the fix a decision, D-184. No firm requirement for D-184 existed in kibi on this branch when it was built.

## Source

> The core bounds every request body and uploaded file before authenticating or parsing it, and every parser bounds what it allocates, so no unauthenticated request or small file can exhaust its memory.

docs/DESIGN.md#D-184

> GR-12: never log amounts, holdings, account numbers or tokens (docs/DESIGN.md#GR-12, docs/dev/code-organization.md#rules).
