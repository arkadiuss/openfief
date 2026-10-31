---
title: For each new XTB ticker, the market-data connector must look up the ISIN by ticker and market and propose it.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-116-1
  - lane:strict
text_ref: design-decisions:D-116#1
origin:
  kind: import
  ref: design-decisions:D-116#1
  recorded_at: '2026-10-31T12:05:38.000Z'
semantic_text: For each new XTB ticker, the market-data connector must look up the ISIN by ticker and market and propose it.
logic_claims:
  - CLAIM-3D0649DCC644C464
semantic_clauses:
  - For each new XTB ticker, the market-data connector must look up the ISIN by ticker and market and propose it.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 717937a190763f89d4d51a161ec10d6c5e82bfa9370fafa8e8f39c570b578964
semantic_inventory:
  - claim_key: CLAIM-3D0649DCC644C464
    claim_text: For each new XTB ticker, the market-data connector must look up the ISIN by ticker and market and propose it
    role: normative
    status: modeled
    span:
      start: 0
      end: 108
id: REQ-ingest-xtb-isin-proposed
type: req
---
For each new XTB ticker, the market-data connector must look up the ISIN by ticker and market and propose it.

## Context

XTB's xStation export names each instrument by a ticker with a market suffix, such as `.PL` or `.UK`, and a display name, and carries no ISIN anywhere (R-24), while every instrument needs an ISIN with a valid check digit (D-35, REQ-decision-d35-1-instrument-isin-symbol). The implementation could not build the XTB import without the export's real layout (question #1 on docs/implementation-questions.md); once the author had read a real export, they decided on 2026-10-31 that the market-data connector, which already finds instruments (D-51), proposes the ISIN, and the owner confirms it (REQ-ingest-xtb-isin-owner-confirms).

## Source

> XTB's export names instruments by ticker and market suffix and carries no ISIN (D-35). For each new ticker the market-data connector looks up the ISIN by ticker and market and proposes it, and the owner confirms it once in the UI, or enters it when the lookup finds none. Rows of a ticker not confirmed yet wait as import problems.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-116#1
