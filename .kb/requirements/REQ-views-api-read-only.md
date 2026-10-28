---
title: 'The read API must only read: every method must be unary, declare no side effects and map to a GET under /v1 without a body, and REST must answer any other method with 405 and Allow: GET before authentication.'
status: open
text_ref: docs/DESIGN.md#D-7
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-7
  - lane:strict
semantic_text: 'The read API must only read: every method must be unary, declare no side effects and map to a GET under /v1 without a body, and REST must answer any other method with 405 and Allow: GET before authentication.'
logic_claims:
  - CLAIM-F1E769F213D5A77F
semantic_clauses:
  - 'The read API must only read: every method must be unary, declare no side effects and map to a GET under /v1 without a body, and REST must answer any other method with 405 and Allow: GET before authentication.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 5e477a10a6c161d47953d7eeee9b7aadbfaefceae60d3c3a516d0a5fd89aaff6
semantic_inventory:
  - claim_key: CLAIM-F1E769F213D5A77F
    claim_text: 'The read API must only read: every method must be unary, declare no side effects and map to a GET under /v1 without a body, and REST must answer any other method with 405 and Allow: GET before authentication'
    role: normative
    status: modeled
    span:
      start: 0
      end: 207
origin:
  kind: agent
  recorded_at: '2026-10-28T08:44:01.543Z'
id: REQ-views-api-read-only
type: req
---
The read API must only read: every method must be unary, declare no side effects and map to a GET under /v1 without a body, and REST must answer any other method with 405 and Allow: GET before authentication.

## Context

Refines the firm requirements REQ-AUTO-69B52C44712F8813 (D-6: read-only) and REQ-AUTO-41A1F4D687DAA771 (D-7: data enters only through connectors' ingest paths, never the public API). Vanguard refuses a method the HTTP rules don't name before the handler, so the refusal reveals nothing. Modeled by Claude Code for the M0 read API task; the author has not reviewed it yet.

## Source

> OpenFief's public API is read-only REST/JSON with an OpenAPI spec, versioned `/v1`, covering accounts, transactions, tax lots, holdings and valuations. Every view plugin, the built-in UI included, uses it.

docs/DESIGN.md D-6

> Data enters only through each connector's ingest path (upload, CLI, watched folder), never through the public API.

docs/DESIGN.md D-7
