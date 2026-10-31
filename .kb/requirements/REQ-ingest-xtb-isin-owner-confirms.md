---
title: The owner must confirm the ISIN proposed for a new XTB ticker once in the UI, or enter one when the lookup finds none.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-116-2
  - lane:strict
text_ref: design-decisions:D-116#2
origin:
  kind: import
  ref: design-decisions:D-116#2
  recorded_at: '2026-10-31T12:05:38.000Z'
semantic_text: The owner must confirm the ISIN proposed for a new XTB ticker once in the UI, or enter one when the lookup finds none.
logic_claims:
  - CLAIM-8A739B345E247717
semantic_clauses:
  - The owner must confirm the ISIN proposed for a new XTB ticker once in the UI, or enter one when the lookup finds none.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 7947c0910bdc24d91ee71fce7dfbf82d3e10a1daa41dbcd1f9b8f7b35284965b
semantic_inventory:
  - claim_key: CLAIM-8A739B345E247717
    claim_text: The owner must confirm the ISIN proposed for a new XTB ticker once in the UI, or enter one when the lookup finds none
    role: normative
    status: modeled
    span:
      start: 0
      end: 117
id: REQ-ingest-xtb-isin-owner-confirms
type: req
---
The owner must confirm the ISIN proposed for a new XTB ticker once in the UI, or enter one when the lookup finds none.

## Context

XTB's export carries only a ticker with a market suffix and no ISIN (R-24), so the market-data connector proposes one by ticker and market (REQ-ingest-xtb-isin-proposed). A lookup can be wrong or find nothing, and every instrument needs an ISIN with a valid check digit (D-35), so the owner confirms each ticker's ISIN once, in the built-in UI, and enters one by hand when the lookup finds none. The author decided this on 2026-10-31 while answering the implementation's question about the XTB import (question #1 on docs/implementation-questions.md).

## Source

> XTB's export names instruments by ticker and market suffix and carries no ISIN (D-35). For each new ticker the market-data connector looks up the ISIN by ticker and market and proposes it, and the owner confirms it once in the UI, or enters it when the lookup finds none. Rows of a ticker not confirmed yet wait as import problems.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-116#2
