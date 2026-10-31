---
title: For each ticker reported for the first time, the core must ask the market-data connectors to propose its ISIN, and the eodhd connector must propose the one ISIN EODHD lists for the ticker's code on its market's supported exchange.
status: open
text_ref: docs/DESIGN.md#D-116
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-116
  - lane:strict
semantic_text: For each ticker reported for the first time, the core must ask the market-data connectors to propose its ISIN, and the eodhd connector must propose the one ISIN EODHD lists for the ticker's code on its market's supported exchange.
logic_claims:
  - CLAIM-94432C2BAE3F37CE
semantic_clauses:
  - For each ticker reported for the first time, the core must ask the market-data connectors to propose its ISIN, and the eodhd connector must propose the one ISIN EODHD lists for the ticker's code on its market's supported exchange.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: c905c5cbe2573d02da7bd46ea79ab0d80ec920c4f1cc1bdb259061aafcbb2141
semantic_inventory:
  - claim_key: CLAIM-94432C2BAE3F37CE
    claim_text: For each ticker reported for the first time, the core must ask the market-data connectors to propose its ISIN, and the eodhd connector must propose the one ISIN EODHD lists for the ticker's code on its market's supported exchange
    role: normative
    status: modeled
    span:
      start: 0
      end: 229
origin:
  kind: agent
  recorded_at: '2026-10-31T12:56:40.137Z'
id: REQ-ingest-ticker-isin-lookup
type: req
---
For each ticker reported for the first time, the core must ask the market-data connectors to propose its ISIN, and the eodhd connector must propose the one ISIN EODHD lists for the ticker's code on its market's supported exchange.

## Context

D-116 has the market-data connector look the ISIN up by ticker and market; D-51 makes EODHD that connector and names its exchanges. Recorded by the XTB implementation agent (impl/xtb) for the author's XTB work package; default firmness, not reviewed by the author yet.

## Source

> XTB's export names instruments by ticker and market suffix and carries no ISIN (D-35). For each new ticker the market-data connector looks up the ISIN by ticker and market and proposes it, and the owner confirms it once in the UI, or enters it when the lookup finds none. Rows of a ticker not confirmed yet wait as import problems.

docs/DESIGN.md#D-116
