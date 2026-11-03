---
title: The core must translate a connector delivery's index levels into index levels keyed by the index name the connector sends, refuse one without a name, a date or a positive level, and store those of an instance that may deliver them with the instance as their source.
status: open
text_ref: docs/DESIGN.md#D-179
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-179
  - lane:strict
semantic_text: The core must translate a connector delivery's index levels into index levels keyed by the index name the connector sends, refuse one without a name, a date or a positive level, and store those of an instance that may deliver them with the instance as their source.
logic_claims:
  - CLAIM-E7D3A6DFC45B2F17
semantic_clauses:
  - The core must translate a connector delivery's index levels into index levels keyed by the index name the connector sends, refuse one without a name, a date or a positive level, and store those of an instance that may deliver them with the instance as their source.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: a17195a48b0ab6acda7fcd6242a830a3cc77f6e2ee4463b9a6c3dcc813694b2c
semantic_inventory:
  - claim_key: CLAIM-E7D3A6DFC45B2F17
    claim_text: The core must translate a connector delivery's index levels into index levels keyed by the index name the connector sends, refuse one without a name, a date or a positive level, and store those of an instance that may deliver them with the instance as their source
    role: normative
    status: modeled
    span:
      start: 0
      end: 264
origin:
  kind: agent
  recorded_at: '2026-11-02T22:25:11.752Z'
id: REQ-ingest-index-levels-allowed-instances
type: req
---
The core must translate a connector delivery's index levels into index levels keyed by the index name the connector sends, refuse one without a name, a date or a positive level, and store those of an instance that may deliver them with the instance as their source.

## Why this changed

On 2026-11-02 the author recorded D-179: market index levels come only from a connector whose signed manifest declares them and whose grant approves that. REQ-ingest-index-levels had the core store every delivery's levels; this requirement replaces it and keeps its translation rules.

## Context

D-51 brings end-of-day index history for the model portfolio's index targets (D-52, D-53); no holding is valued from a level. The eodhd connector declares index levels.

## Source

> Only a connector whose signed manifest declares bond series terms, or market index levels, and whose grant approves that may deliver them; anything else it delivers of that kind is dropped with an attention item (D-130), as D-141 does for cross-check prices. Bond terms need a Polish (`PL`) ISIN, and every stored term and index level records the instance that delivered it.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-179#1

