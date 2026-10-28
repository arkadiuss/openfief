---
title: The EODHD plugin must answer Sync with the end-of-day closes of each configured instrument and the levels of each configured market index after that series' own watermark mark, or from its configured start, up to today, from institution eodhd; must report an unlisted, ambiguous or unknown series as an ingest problem and still deliver the rest; must answer Unavailable when EODHD fails and FailedPrecondition without an API key; and must refuse ImportFile.
status: open
text_ref: docs/DESIGN.md#D-51
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-51
  - lane:strict
semantic_text: The EODHD plugin must answer Sync with the end-of-day closes of each configured instrument and the levels of each configured market index after that series' own watermark mark, or from its configured start, up to today, from institution eodhd; must report an unlisted, ambiguous or unknown series as an ingest problem and still deliver the rest; must answer Unavailable when EODHD fails and FailedPrecondition without an API key; and must refuse ImportFile.
logic_claims:
  - CLAIM-A90B0A001F621EC3
semantic_clauses:
  - The EODHD plugin must answer Sync with the end-of-day closes of each configured instrument and the levels of each configured market index after that series' own watermark mark, or from its configured start, up to today, from institution eodhd; must report an unlisted, ambiguous or unknown series as an ingest problem and still deliver the rest; must answer Unavailable when EODHD fails and FailedPrecondition without an API key; and must refuse ImportFile.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: ccf82268d043ad0ed9c7bac8aa2af8fe6a567ead0be7e61f943e40328c0a93f0
semantic_inventory:
  - claim_key: CLAIM-A90B0A001F621EC3
    claim_text: The EODHD plugin must answer Sync with the end-of-day closes of each configured instrument and the levels of each configured market index after that series' own watermark mark, or from its configured start, up to today, from institution eodhd; must report an unlisted, ambiguous or unknown series as an ingest problem and still deliver the rest; must answer Unavailable when EODHD fails and FailedPrecondition without an API key; and must refuse ImportFile
    role: normative
    status: modeled
    span:
      start: 0
      end: 456
origin:
  kind: agent
  recorded_at: '2026-10-28T21:37:44.418Z'
id: REQ-plugins-eodhd-sync
type: req
---
The EODHD plugin must answer Sync with the end-of-day closes of each configured instrument and the levels of each configured market index after that series' own watermark mark, or from its configured start, up to today, from institution eodhd; must report an unlisted, ambiguous or unknown series as an ingest problem and still deliver the rest; must answer Unavailable when EODHD fails and FailedPrecondition without an API key; and must refuse ImportFile.

## Context

D-49 has the core call connectors and own retries, so a failed fetch fails the batch for a retry, while a series that can't be priced is a problem, not an error, so one bad ISIN doesn't block the others. The core's SyncRequest names no instruments, so the instance's config lists them (D-48); a per-series watermark lets an instrument added later start from the configured start. D-22 and D-51 make this the market-data connector on the owner's key. Refines REQ-AUTO-0C514DFCDB36FBEB and REQ-valuation-eodhd-connector. Modeled by Claude Code for the EODHD market-data connector task; the author has not reviewed it yet.

## Source

> The first market-data connector uses EODHD. EODHD covers the Warsaw Stock Exchange, LSE, Xetra and Euronext, finds instruments by ISIN, includes end-of-day index history, and its terms allow storing the data for private use. Its closes are aggregated rather than each exchange's official close, which is one more reason for D-22's cross-check.

docs/DESIGN.md D-51

> The core calls connectors: a connector serves a Connect service for syncing and for importing files, and the core owns the schedules, retries and import batches.

docs/DESIGN.md D-49
