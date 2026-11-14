---
title: A sync's audit entry must count the FX rates, index readings and index levels it wrote that were new or different, as it counts prices and bond terms, so a connector that re-sends the same market data writes nothing and leaves no entry.
status: open
firmness: default
required_evidence: unverified
origin:
  kind: agent
  ref: the S9b brief, author decisions of 2026-11-14 on D-211, D-198, D-196 and D-18
  recorded_at: '2026-11-14T06:30:13.805Z'
semantic_text: A sync's audit entry must count the FX rates, index readings and index levels it wrote that were new or different, as it counts prices and bond terms, so a connector that re-sends the same market data writes nothing and leaves no entry.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 168728f641c8fae52b1406888136c6c8189ab5ccdb3ec1864696f58149b31ed5
semantic_inventory:
  - claim_key: CLAIM-237BD644B675E7B5
    claim_text: A sync's audit entry must count the FX rates, index readings and index levels it wrote that were new or different, as it counts prices and bond terms, so a connector that re-sends the same market data writes nothing and leaves no entry
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 235
logic_claims:
  - CLAIM-237BD644B675E7B5
id: REQ-ingest-sync-counts-all-market-data
type: req
---
A sync's audit entry must count the FX rates, index readings and index levels it wrote that were new or different, as it counts prices and bond terms, so a connector that re-sends the same market data writes nothing and leaves no entry.

## Context

The stores' upserts skip a row that already holds the incoming values (IS DISTINCT FROM) and count the rows they affected; an index level counts when its source changed too. The entry's counts are fx_rates, index_readings and index_levels beside transactions, prices and bond_terms, and never an amount. Cross-check prices are not counted.

## Source

> connector_synced ALSO COUNTS FX RATES AND INDEX LEVELS (all market data a sync writes), via the same changed-rows upsert (IS DISTINCT FROM, count affected rows) so identical re-sends write nothing.

Source: the S9b brief, author decisions of 2026-11-14 on D-211, D-198, D-196 and D-18