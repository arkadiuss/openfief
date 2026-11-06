---
title: The owner's UI must list the tickers waiting for their ISIN with the proposal to confirm or a field to type it, hand the owner's confirmation to the core, and link the page from an upload whose rows wait.
status: closed
text_ref: docs/DESIGN.md#D-116
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-116
  - lane:strict
semantic_text: The owner's UI must list the tickers waiting for their ISIN with the proposal to confirm or a field to type it, hand the owner's confirmation to the core, and link the page from an upload whose rows wait.
logic_claims:
  - CLAIM-B0E83A38579F23CE
semantic_clauses:
  - The owner's UI must list the tickers waiting for their ISIN with the proposal to confirm or a field to type it, hand the owner's confirmation to the core, and link the page from an upload whose rows wait.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 9ae689319bee1246f027f51f20f1d70e28bab06b30d6fb79c55ff0dc2df9ebd6
semantic_inventory:
  - claim_key: CLAIM-B0E83A38579F23CE
    claim_text: The owner's UI must list the tickers waiting for their ISIN with the proposal to confirm or a field to type it, hand the owner's confirmation to the core, and link the page from an upload whose rows wait
    role: normative
    status: modeled
    span:
      start: 0
      end: 203
origin:
  kind: agent
  recorded_at: '2026-10-31T12:56:58.079Z'
id: REQ-ui-tickers-to-confirm
type: req
---
The owner's UI must list the tickers waiting for their ISIN with the proposal to confirm or a field to type it, hand the owner's confirmation to the core, and link the page from an upload whose rows wait.

## Context

D-116 has the owner confirm in the UI, on the upload result or a tickers page as the work package asked. Recorded by the XTB implementation agent (impl/xtb) for the author's XTB work package; default firmness, not reviewed by the author yet.

## Source

> XTB's export names instruments by ticker and market suffix and carries no ISIN (D-35). For each new ticker the market-data connector looks up the ISIN by ticker and market and proposes it, and the owner confirms it once in the UI, or enters it when the lookup finds none. Rows of a ticker not confirmed yet wait as import problems.

docs/DESIGN.md#D-116
