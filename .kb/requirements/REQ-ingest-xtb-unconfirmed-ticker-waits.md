---
title: Rows of an XTB ticker whose ISIN is not confirmed yet must wait as import problems.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-116-3
  - lane:strict
text_ref: design-decisions:D-116#3
origin:
  kind: import
  ref: design-decisions:D-116#3
  recorded_at: '2026-10-31T12:05:38.000Z'
semantic_text: Rows of an XTB ticker whose ISIN is not confirmed yet must wait as import problems.
logic_claims:
  - CLAIM-DC41B322C7F03C65
semantic_clauses:
  - Rows of an XTB ticker whose ISIN is not confirmed yet must wait as import problems.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 8cd6a890dd5f2407c4beeedfe0947103178b55aff585ea90af3448c80880117d
semantic_inventory:
  - claim_key: CLAIM-DC41B322C7F03C65
    claim_text: Rows of an XTB ticker whose ISIN is not confirmed yet must wait as import problems
    role: normative
    status: modeled
    span:
      start: 0
      end: 82
id: REQ-ingest-xtb-unconfirmed-ticker-waits
type: req
---
Rows of an XTB ticker whose ISIN is not confirmed yet must wait as import problems.

## Context

An XTB row names its instrument only by ticker and market (R-24), and nothing is booked without an ISIN (D-35). Until the owner confirms or enters the ticker's ISIN (REQ-ingest-xtb-isin-owner-confirms), its rows are neither booked under a guessed instrument nor dropped: they wait as import problems the owner can see. The author decided this on 2026-10-31 while answering the implementation's question about the XTB import (question #1 on docs/implementation-questions.md).

## Source

> XTB's export names instruments by ticker and market suffix and carries no ISIN (D-35). For each new ticker the market-data connector looks up the ISIN by ticker and market and proposes it, and the owner confirms it once in the UI, or enters it when the lookup finds none. Rows of a ticker not confirmed yet wait as import problems.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-116#3
