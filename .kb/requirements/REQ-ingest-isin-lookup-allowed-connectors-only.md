---
title: 'A ticker''s ISIN lookup must go only to connectors allowed to learn what the owner holds: those whose manifest declares them market-data connectors and whose grant approves that.'
status: open
text_ref: design-decisions:D-166#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-166-1
  - lane:strict
semantic_text: 'A ticker''s ISIN lookup must go only to connectors allowed to learn what the owner holds: those whose manifest declares them market-data connectors and whose grant approves that.'
logic_claims:
  - CLAIM-2E620D9AB6C98199
semantic_clauses:
  - 'A ticker''s ISIN lookup must go only to connectors allowed to learn what the owner holds: those whose manifest declares them market-data connectors and whose grant approves that.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 4180d7d205a74de53245b9228faaacd9b719a5a15d779f4fb9d00e425608c007
semantic_inventory:
  - claim_key: CLAIM-2E620D9AB6C98199
    claim_text: 'A ticker''s ISIN lookup must go only to connectors allowed to learn what the owner holds: those whose manifest declares them market-data connectors and whose grant approves that'
    role: normative
    status: modeled
    span:
      start: 0
      end: 176
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-166#1
  recorded_at: '2026-11-02T00:55:48.245Z'
id: REQ-ingest-isin-lookup-allowed-connectors-only
type: req
---
A ticker's ISIN lookup must go only to connectors allowed to learn what the owner holds: those whose manifest declares them market-data connectors and whose grant approves that.

## Context

The author recorded it on 2026-11-01. The decision gives the reason: a lookup also tells the connector what the owner holds, which D-123 allows only to declared and approved market-data connectors (REQ-ingest-market-data-list-declared-granted). No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> D-116's ISIN lookup goes only to the connectors D-123 allows, since a lookup also tells the connector what the owner holds.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-166#1
