---
title: 'Exported spans must never carry money amounts, quantities, account numbers, ISINs, tokens, session IDs, secrets, SQL statements or their arguments, or error messages: only an allow-list of attributes may leave the process.'
status: open
text_ref: docs/DESIGN.md#GR-12
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-gr-12
  - lane:strict
semantic_text: 'Exported spans must never carry money amounts, quantities, account numbers, ISINs, tokens, session IDs, secrets, SQL statements or their arguments, or error messages: only an allow-list of attributes may leave the process.'
logic_claims:
  - CLAIM-1706FF3D39230191
semantic_clauses:
  - 'Exported spans must never carry money amounts, quantities, account numbers, ISINs, tokens, session IDs, secrets, SQL statements or their arguments, or error messages: only an allow-list of attributes may leave the process.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 8afbaa7d1bfb6b9c1052f794abce528d2321a9c6d6bc143ed0503a16b01c86b5
semantic_inventory:
  - claim_key: CLAIM-1706FF3D39230191
    claim_text: 'Exported spans must never carry money amounts, quantities, account numbers, ISINs, tokens, session IDs, secrets, SQL statements or their arguments, or error messages: only an allow-list of attributes may leave the process'
    role: normative
    status: modeled
    span:
      start: 0
      end: 221
origin:
  kind: agent
  recorded_at: '2026-11-02T03:07:15.537Z'
id: REQ-telemetry-private-spans
type: req
---
Exported spans must never carry money amounts, quantities, account numbers, ISINs, tokens, session IDs, secrets, SQL statements or their arguments, or error messages: only an allow-list of attributes may leave the process.

## Context

The repository owner asked for it directly ("add tracing and make pushing to otel configurable"). No decision in docs/DESIGN.md covers observability, so this is the narrowest sound version; docs/implementation-questions.md#tracing (items 71 to 74) parks the choices it leaves open for the author. Modeled by Claude Code; the author has not reviewed it yet. Traces leave the cluster for a collector the owner may not run, so they follow the rule for logs: internal IDs only, never amounts, holdings, account numbers or tokens (docs/dev/code-organization.md, AGENTS.md). Spans are named by operation, queries record no statement, and telemetry.Redact enforces an allow-list on export.

## Source

> [GR-12] The privacy and security model comes first, designed per use case.

docs/DESIGN.md#GR-12; the owner's request: "add tracing and make pushing to otel configurable"
