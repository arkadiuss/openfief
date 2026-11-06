---
title: The connector contract must carry the holdings a delivery reports of instruments named by ticker only, and the core must refuse one that names no ticker of its delivery, has no date or has a negative quantity.
status: open
text_ref: docs/DESIGN.md#D-155
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-155
  - lane:strict
semantic_text: The connector contract must carry the holdings a delivery reports of instruments named by ticker only, and the core must refuse one that names no ticker of its delivery, has no date or has a negative quantity.
logic_claims:
  - CLAIM-2AF0B9208DEBE76D
semantic_clauses:
  - The connector contract must carry the holdings a delivery reports of instruments named by ticker only, and the core must refuse one that names no ticker of its delivery, has no date or has a negative quantity.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 26f1ad0e3957a5880b3e59d6a33e71d1752ed1c5c07d70fcd7d585a2bf65fa6d
semantic_inventory:
  - claim_key: CLAIM-2AF0B9208DEBE76D
    claim_text: The connector contract must carry the holdings a delivery reports of instruments named by ticker only, and the core must refuse one that names no ticker of its delivery, has no date or has a negative quantity
    role: normative
    status: modeled
    span:
      start: 0
      end: 208
origin:
  kind: agent
  recorded_at: '2026-11-02T07:00:36.034Z'
id: REQ-ingest-ticker-assertion-contract
type: req
---
The connector contract must carry the holdings a delivery reports of instruments named by ticker only, and the core must refuse one that names no ticker of its delivery, has no date or has a negative quantity.

## Context

D-155 asserts an XTB export's holdings, which the xtb connector plugin reports by ticker since the export names no ISIN; the TickerAssertion message and its refusals are this implementation's reading.

## Source

> **[D-155]** An XTB export's `Open Positions` sheet also asserts the account's holdings (D-18): for each ticker with a confirmed ISIN (D-116), the summed volume of its open positions on the report's date. A ticker not confirmed yet is asserted once the owner confirms it.

docs/DESIGN.md#D-155
