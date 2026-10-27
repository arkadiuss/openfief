---
title: Data must enter OpenFief only through a connector's ingest path (upload, CLI or watched folder), never through the public API.
status: open
text_ref: design-decisions:D-7#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-7-1
  - lane:strict
semantic_text: Data must enter OpenFief only through a connector's ingest path (upload, CLI or watched folder), never through the public API.
logic_claims:
  - CLAIM-6AD528F80096DDDE
semantic_clauses:
  - Data must enter OpenFief only through a connector's ingest path (upload, CLI or watched folder), never through the public API.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 1469e0aec21b6c133f424b2cd73c8de72b6c217a5230af7c265414db26b2e0f2
semantic_inventory:
  - claim_key: CLAIM-6AD528F80096DDDE
    claim_text: Data must enter OpenFief only through a connector's ingest path (upload, CLI or watched folder), never through the public API
    role: normative
    status: modeled
    span:
      start: 0
      end: 125
origin:
  kind: import
  ref: design-decisions:D-7#1
  recorded_at: '2026-10-27T20:13:00.359Z'
firmness: firm
required_evidence: asserted
id: REQ-AUTO-41A1F4D687DAA771
type: req
---
Data must enter OpenFief only through a connector's ingest path (upload, CLI or watched folder), never through the public API.

## Source

> GR-3 reading: connectors get data from each institution by whatever works (a broker API where one exists, files or other means otherwise; not necessarily REST), and OpenFief exposes its own read API. Data enters only through each connector's ingest path (upload, CLI, watched folder), never through the public API.

Source: OpenFief author decisions (DESIGN.md section 10) - D-7#1
