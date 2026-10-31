---
title: The connector contract must carry the tickers a delivery names without an ISIN and their transactions, and the core must refuse a ticker transaction that names no ticker of its delivery or books no valid transaction.
status: open
text_ref: docs/DESIGN.md#D-116
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-116
  - lane:strict
semantic_text: The connector contract must carry the tickers a delivery names without an ISIN and their transactions, and the core must refuse a ticker transaction that names no ticker of its delivery or books no valid transaction.
logic_claims:
  - CLAIM-1F6746954BA09933
semantic_clauses:
  - The connector contract must carry the tickers a delivery names without an ISIN and their transactions, and the core must refuse a ticker transaction that names no ticker of its delivery or books no valid transaction.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 1ad30bda6d65723b7901ade952776260d3385257aa411088b79f1f7d70c3e15e
semantic_inventory:
  - claim_key: CLAIM-1F6746954BA09933
    claim_text: The connector contract must carry the tickers a delivery names without an ISIN and their transactions, and the core must refuse a ticker transaction that names no ticker of its delivery or books no valid transaction
    role: normative
    status: modeled
    span:
      start: 0
      end: 215
origin:
  kind: agent
  recorded_at: '2026-10-31T12:56:33.346Z'
id: REQ-ingest-ticker-contract
type: req
---
The connector contract must carry the tickers a delivery names without an ISIN and their transactions, and the core must refuse a ticker transaction that names no ticker of its delivery or books no valid transaction.

## Context

D-116 needs the connector to report tickers and their rows before an ISIN exists; D-49 has the core call connectors over the contract. Recorded by the XTB implementation agent (impl/xtb) for the author's XTB work package; default firmness, not reviewed by the author yet.

## Source

> XTB's export names instruments by ticker and market suffix and carries no ISIN (D-35). For each new ticker the market-data connector looks up the ISIN by ticker and market and proposes it, and the owner confirms it once in the UI, or enters it when the lookup finds none. Rows of a ticker not confirmed yet wait as import problems.

docs/DESIGN.md#D-116
