---
title: The core must translate a connector delivery's index levels into index levels keyed by the index name the connector sends, refuse one without a name, a date or a positive level, and store them with the connector instance as their source.
status: open
text_ref: docs/DESIGN.md#D-51
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-51
  - lane:strict
semantic_text: The core must translate a connector delivery's index levels into index levels keyed by the index name the connector sends, refuse one without a name, a date or a positive level, and store them with the connector instance as their source.
logic_claims:
  - CLAIM-BBDB965E01F00324
semantic_clauses:
  - The core must translate a connector delivery's index levels into index levels keyed by the index name the connector sends, refuse one without a name, a date or a positive level, and store them with the connector instance as their source.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: fd48915cfb0c55ebcccc42ee12f55277cdbe9dbfe090ff8f9d22b57462de380d
semantic_inventory:
  - claim_key: CLAIM-BBDB965E01F00324
    claim_text: The core must translate a connector delivery's index levels into index levels keyed by the index name the connector sends, refuse one without a name, a date or a positive level, and store them with the connector instance as their source
    role: normative
    status: modeled
    span:
      start: 0
      end: 236
origin:
  kind: agent
  recorded_at: '2026-10-28T22:42:35.865Z'
id: REQ-ingest-index-levels
type: req
---
The core must translate a connector delivery's index levels into index levels keyed by the index name the connector sends, refuse one without a name, a date or a positive level, and store them with the connector instance as their source.

## Context

The connector contract's IndexLevel (D-49, D-51) was delivered but dropped; the orchestrator asked to store it in the style of prices and to key index targets by the same name. Modeled by Claude Code for the model portfolio task (index targets need stored levels); the author has not reviewed it yet.

## Source

> **[D-51]** The first market-data connector uses EODHD. EODHD covers the Warsaw Stock Exchange, LSE, Xetra and Euronext, finds instruments by ISIN, includes end-of-day index history, and its terms allow storing the data for private use.

docs/DESIGN.md#D-51

> **[D-52]** each target is an instrument (such as an index ETF or a stock) or an index, with a weight

docs/DESIGN.md#D-52
