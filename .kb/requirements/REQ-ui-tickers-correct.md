---
title: The owner's UI must list the confirmed tickers once per institution with a form to correct each one's ISIN, and hand the owner's correction to the core.
status: open
text_ref: docs/DESIGN.md#D-153
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-153
  - lane:strict
semantic_text: The owner's UI must list the confirmed tickers once per institution with a form to correct each one's ISIN, and hand the owner's correction to the core.
logic_claims:
  - CLAIM-CE04C22259D7C5F8
semantic_clauses:
  - The owner's UI must list the confirmed tickers once per institution with a form to correct each one's ISIN, and hand the owner's correction to the core.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: acb34628eb8743fc6cdbe79a39fd8018badb3fa4deda0ff1ffa147d5cbce5b69
semantic_inventory:
  - claim_key: CLAIM-CE04C22259D7C5F8
    claim_text: The owner's UI must list the confirmed tickers once per institution with a form to correct each one's ISIN, and hand the owner's correction to the core
    role: normative
    status: modeled
    span:
      start: 0
      end: 151
origin:
  kind: agent
  recorded_at: '2026-11-02T07:00:58.637Z'
id: REQ-ui-tickers-correct
type: req
---
The owner's UI must list the confirmed tickers once per institution with a form to correct each one's ISIN, and hand the owner's correction to the core.

## Context

D-153 lets the owner correct a confirmed ISIN, which D-116 has the owner confirm in the UI; the page layout is this implementation's choice.

## Source

> **[D-153]** A ticker's confirmed ISIN (D-116) holds for every account at the same institution. The owner can correct a wrong one, which supersedes the transactions it booked (D-18).

docs/DESIGN.md#D-153
