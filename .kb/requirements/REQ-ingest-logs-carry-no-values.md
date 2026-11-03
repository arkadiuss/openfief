---
title: The ingest context must log a failure only by the known errors in its chain, or else by its innermost error's type, never with the amounts, holdings, ISINs or account numbers its text may carry, and the ledger's refusals of a transaction must not quote its amount or quantity.
status: open
text_ref: docs/DESIGN.md#GR-12
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-gr-12
  - lane:strict
semantic_text: The ingest context must log a failure only by the known errors in its chain, or else by its innermost error's type, never with the amounts, holdings, ISINs or account numbers its text may carry, and the ledger's refusals of a transaction must not quote its amount or quantity.
logic_claims:
  - CLAIM-0A05AB1AE74B1EA8
semantic_clauses:
  - The ingest context must log a failure only by the known errors in its chain, or else by its innermost error's type, never with the amounts, holdings, ISINs or account numbers its text may carry, and the ledger's refusals of a transaction must not quote its amount or quantity.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: c0f09d7f3bb7ec05f264fa4ba651c8360c403e72dfe8a9c14fec7f53bfe2aee2
semantic_inventory:
  - claim_key: CLAIM-0A05AB1AE74B1EA8
    claim_text: The ingest context must log a failure only by the known errors in its chain, or else by its innermost error's type, never with the amounts, holdings, ISINs or account numbers its text may carry, and the ledger's refusals of a transaction must not quote its amount or quantity
    role: normative
    status: modeled
    span:
      start: 0
      end: 275
origin:
  kind: agent
  recorded_at: '2026-11-03T01:05:26.128Z'
id: REQ-ingest-logs-carry-no-values
type: req
---
The ingest context must log a failure only by the known errors in its chain, or else by its innermost error's type, never with the amounts, holdings, ISINs or account numbers its text may carry, and the ledger's refusals of a transaction must not quote its amount or quantity.

## Context

Errors wrap what failed, and their text is built for the owner, who may see his own values; a log line is kept and shipped elsewhere, so it carries only a value-free description. Other contexts' services log their own errors and are not covered here. Recorded with default firmness by the implementation agent of the alpha security fixes (fix/p1-isolation-net), from the security review the author put before the alpha tag: errors that carry amounts, ISINs or account numbers reached the logs through the sync scheduler and the ingest services.

## Source

> Never log amounts, holdings, account numbers or tokens. Log internal IDs only (GR-12, UC-0).

docs/dev/code-organization.md#rules
