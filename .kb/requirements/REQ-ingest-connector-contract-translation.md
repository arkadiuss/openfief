---
title: The core must translate connector contract messages into domain values in its Connect adapter, refusing, with the field named, a decimal with an exponent or grouping, a date that is not YYYY-MM-DD, a malformed currency or ISIN, a missing required field, an unspecified enum or a delivery for another batch; no contract type may reach the domain or the use cases.
status: open
text_ref: docs/dev/code-organization.md
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-code-organization-md
  - lane:strict
semantic_text: The core must translate connector contract messages into domain values in its Connect adapter, refusing, with the field named, a decimal with an exponent or grouping, a date that is not YYYY-MM-DD, a malformed currency or ISIN, a missing required field, an unspecified enum or a delivery for another batch; no contract type may reach the domain or the use cases.
logic_claims:
  - CLAIM-A7EAE16BB38054D2
semantic_clauses:
  - The core must translate connector contract messages into domain values in its Connect adapter, refusing, with the field named, a decimal with an exponent or grouping, a date that is not YYYY-MM-DD, a malformed currency or ISIN, a missing required field, an unspecified enum or a delivery for another batch; no contract type may reach the domain or the use cases.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: b0ea0c3a251316d8bb1267c38a8fe6a7de84a57a2e4a001ae51bb403df184a42
semantic_inventory:
  - claim_key: CLAIM-A7EAE16BB38054D2
    claim_text: The core must translate connector contract messages into domain values in its Connect adapter, refusing, with the field named, a decimal with an exponent or grouping, a date that is not YYYY-MM-DD, a malformed currency or ISIN, a missing required field, an unspecified enum or a delivery for another batch; no contract type may reach the domain or the use cases
    role: normative
    status: modeled
    span:
      start: 0
      end: 361
origin:
  kind: agent
  recorded_at: '2026-10-28T08:55:58.328Z'
id: REQ-ingest-connector-contract-translation
type: req
---
The core must translate connector contract messages into domain values in its Connect adapter, refusing, with the field named, a decimal with an exponent or grouping, a date that is not YYYY-MM-DD, a malformed currency or ISIN, a missing required field, an unspecified enum or a delivery for another batch; no contract type may reach the domain or the use cases.

## Context

docs/dev/code-organization.md requires anti-corruption at adapters: wire and vendor formats are translated there and the domain never sees them. Plain decimal strings keep amounts exact (no float64 for money). Supports REQ-AUTO-BC327C2200FC820F. Modeled by Claude Code for the M0 plugin contracts task; the author has not reviewed it yet.

## Source

> Anti-corruption. Broker and vendor formats are parsed and translated in their adapter or plugin. Domain code never sees an IBKR, XTB or Ghostfolio shape.

docs/dev/code-organization.md
