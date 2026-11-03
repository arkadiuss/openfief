---
title: Every context's Connect services, the built-in UI and the core's attention sweep must log a failure only by the known errors in its chain, or else by its innermost error's type, never with the amounts, holdings, ISINs or account numbers its text may carry.
status: open
text_ref: docs/DESIGN.md#GR-12
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-gr-12
  - lane:strict
semantic_text: Every context's Connect services, the built-in UI and the core's attention sweep must log a failure only by the known errors in its chain, or else by its innermost error's type, never with the amounts, holdings, ISINs or account numbers its text may carry.
logic_claims:
  - CLAIM-4B500FB65A6E517B
semantic_clauses:
  - Every context's Connect services, the built-in UI and the core's attention sweep must log a failure only by the known errors in its chain, or else by its innermost error's type, never with the amounts, holdings, ISINs or account numbers its text may carry.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 708e97bbeb88c4d8231c31f70b894e9ece31a473a75502e0c5f04c2b5449815c
semantic_inventory:
  - claim_key: CLAIM-4B500FB65A6E517B
    claim_text: Every context's Connect services, the built-in UI and the core's attention sweep must log a failure only by the known errors in its chain, or else by its innermost error's type, never with the amounts, holdings, ISINs or account numbers its text may carry
    role: normative
    status: modeled
    span:
      start: 0
      end: 255
origin:
  kind: agent
  recorded_at: '2026-11-03T03:24:27.541Z'
id: REQ-core-logs-carry-no-values
type: req
---
Every context's Connect services, the built-in UI and the core's attention sweep must log a failure only by the known errors in its chain, or else by its innermost error's type, never with the amounts, holdings, ISINs or account numbers its text may carry.

## Context

GR-12 puts privacy first, and the code organization's rule is never to log amounts, holdings, account numbers or tokens. REQ-ingest-logs-carry-no-values did this for the ingest context; the author asked, before the first alpha, for every other sink to follow, sharing one helper (kernel.LogSafe) that each context gives its own known errors.

## Source

> **[GR-12]** The privacy and security model comes first, designed per use case.

docs/DESIGN.md#GR-12
