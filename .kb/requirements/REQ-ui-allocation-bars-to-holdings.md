---
title: An allocation bar of an asset class or a currency must link to a page listing the valued holdings and cash of that group on its date across the accounts the caller may view, each holding linked to its instrument.
status: open
text_ref: docs/DESIGN.md#D-195
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-195
  - lane:strict
semantic_text: An allocation bar of an asset class or a currency must link to a page listing the valued holdings and cash of that group on its date across the accounts the caller may view, each holding linked to its instrument.
logic_claims:
  - CLAIM-0D8EFE91316C6345
semantic_clauses:
  - An allocation bar of an asset class or a currency must link to a page listing the valued holdings and cash of that group on its date across the accounts the caller may view, each holding linked to its instrument.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 6f87f927def37490b69275f07981a10e12ee5af25723be7339dd734298116f94
semantic_inventory:
  - claim_key: CLAIM-0D8EFE91316C6345
    claim_text: An allocation bar of an asset class or a currency must link to a page listing the valued holdings and cash of that group on its date across the accounts the caller may view, each holding linked to its instrument
    role: normative
    status: modeled
    span:
      start: 0
      end: 211
origin:
  kind: agent
  recorded_at: '2026-11-06T10:43:34.616Z'
id: REQ-ui-allocation-bars-to-holdings
type: req
---
An allocation bar of an asset class or a currency must link to a page listing the valued holdings and cash of that group on its date across the accounts the caller may view, each holding linked to its instrument.

## Context

D-195: an allocation bar links to the holdings behind it; the read API has no filtered holdings call, so the UI lists them itself. Modeled by Claude Code for the UI v2 pages slice; the author has not reviewed the wording.

## Source

> **[D-195]** Every figure links down to what it is made of: the total to its accounts, an account to its holdings, a holding to its instrument, an allocation bar to the holdings behind it. An instrument has its own page: its price history and source, its holdings and lots across accounts, its transactions, its price checks, its ticker and ISIN, and its OKI class.

docs/DESIGN.md#D-195
