---
title: The public API must be defined in protobuf with HTTP annotations and served both as Connect and as REST under /v1, with its OpenAPI spec generated from the same definitions.
status: open
text_ref: design-decisions:D-50#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-50-1
  - lane:strict
semantic_text: The public API must be defined in protobuf with HTTP annotations and served both as Connect and as REST under /v1, with its OpenAPI spec generated from the same definitions.
logic_claims:
  - CLAIM-9349E75B04C5A204
semantic_clauses:
  - The public API must be defined in protobuf with HTTP annotations and served both as Connect and as REST under /v1, with its OpenAPI spec generated from the same definitions.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: f172b08124241fd1701b726bd2c4c62245e17e0041d40cf131b215245dbd9c2f
semantic_inventory:
  - claim_key: CLAIM-9349E75B04C5A204
    claim_text: The public API must be defined in protobuf with HTTP annotations and served both as Connect and as REST under /v1, with its OpenAPI spec generated from the same definitions
    role: normative
    status: modeled
    span:
      start: 0
      end: 172
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-50#1
  recorded_at: '2026-10-28T06:08:50.745Z'
id: REQ-views-read-api-protobuf
type: req
---
The public API must be defined in protobuf with HTTP annotations and served both as Connect and as REST under /v1, with its OpenAPI spec generated from the same definitions.

## Context

Existing trackers can't hold what OpenFief has to get exactly right: FIFO tax lots priced in PLN, statutory bond values, and sleeve attribution that the later lease ledger reconciles against. So OpenFief owns the record, and trackers only display it.

## Source

> The read API (D-6) is defined in protobuf with HTTP annotations, like the plugin contracts (D-11). The core serves it both as Connect, for generated Go and TypeScript clients, and as REST under `/v1` through transcoding (Buf's Vanguard). The OpenAPI spec is generated from the same definitions.

OpenFief author decisions (docs/DESIGN.md, section 3.1), design-decisions:D-50#1
