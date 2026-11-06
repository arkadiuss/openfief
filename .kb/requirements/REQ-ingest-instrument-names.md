---
title: 'Connectors must deliver instruments'' display names: IBKR''s trade and position descriptions in the delivery''s instrument names, and a ticker''s institution name, or else the lookup''s name for the proposed ISIN, when the owner confirms it.'
status: open
text_ref: docs/DESIGN.md#D-197
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-197
  - lane:strict
semantic_text: 'Connectors must deliver instruments'' display names: IBKR''s trade and position descriptions in the delivery''s instrument names, and a ticker''s institution name, or else the lookup''s name for the proposed ISIN, when the owner confirms it.'
logic_claims:
  - CLAIM-54CCCB0B2979D1AA
semantic_clauses:
  - 'Connectors must deliver instruments'' display names: IBKR''s trade and position descriptions in the delivery''s instrument names, and a ticker''s institution name, or else the lookup''s name for the proposed ISIN, when the owner confirms it.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 0fcc607316f05b1869415915e07b468957aeb4ff2cb339754647b9b9145460d3
semantic_inventory:
  - claim_key: CLAIM-54CCCB0B2979D1AA
    claim_text: 'Connectors must deliver instruments'' display names: IBKR''s trade and position descriptions in the delivery''s instrument names, and a ticker''s institution name, or else the lookup''s name for the proposed ISIN, when the owner confirms it'
    role: normative
    status: modeled
    span:
      start: 0
      end: 235
origin:
  kind: agent
  recorded_at: '2026-11-06T06:59:21.467Z'
id: REQ-ingest-instrument-names
type: req
---
Connectors must deliver instruments' display names: IBKR's trade and position descriptions in the delivery's instrument names, and a ticker's institution name, or else the lookup's name for the proposed ISIN, when the owner confirms it.

## Context

Refines D-197's sources of names: XTB's display name, IBKR's description and the market-data lookup's. Modeled by Claude Code for the UI v2 overview slice; the author has not reviewed the wording.

## Source

> **[D-197]** An instrument has a display name: the first one a connector delivers for it, such as XTB's display name, IBKR's description or the market-data lookup's, which the owner can change. Pages show it before the ticker and ISIN.

docs/DESIGN.md#D-197
