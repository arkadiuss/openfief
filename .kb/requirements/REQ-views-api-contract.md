---
title: 'The read API must be the openfief.views.v1.ReadService protobuf service with HTTP annotations, served from one set of handlers as Connect and, transcoded by Vanguard, as REST under /v1, whose REST responses match the OpenAPI document generated from the same definitions: money and quantities as decimal strings, dates as YYYY-MM-DD, and errors mapped to not found, invalid argument or internal.'
status: open
text_ref: docs/DESIGN.md#D-50
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-50
  - lane:strict
semantic_text: 'The read API must be the openfief.views.v1.ReadService protobuf service with HTTP annotations, served from one set of handlers as Connect and, transcoded by Vanguard, as REST under /v1, whose REST responses match the OpenAPI document generated from the same definitions: money and quantities as decimal strings, dates as YYYY-MM-DD, and errors mapped to not found, invalid argument or internal.'
logic_claims:
  - CLAIM-F8D92AA3C4096CAA
semantic_clauses:
  - 'The read API must be the openfief.views.v1.ReadService protobuf service with HTTP annotations, served from one set of handlers as Connect and, transcoded by Vanguard, as REST under /v1, whose REST responses match the OpenAPI document generated from the same definitions: money and quantities as decimal strings, dates as YYYY-MM-DD, and errors mapped to not found, invalid argument or internal.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: e06372d81adef0c8e8c17dabdd7ab058f0918ff0ae0385faca1503db3920cd95
semantic_inventory:
  - claim_key: CLAIM-F8D92AA3C4096CAA
    claim_text: 'The read API must be the openfief.views.v1.ReadService protobuf service with HTTP annotations, served from one set of handlers as Connect and, transcoded by Vanguard, as REST under /v1, whose REST responses match the OpenAPI document generated from the same definitions: money and quantities as decimal strings, dates as YYYY-MM-DD, and errors mapped to not found, invalid argument or internal'
    role: normative
    status: modeled
    span:
      start: 0
      end: 393
origin:
  kind: agent
  recorded_at: '2026-10-28T08:43:45.250Z'
id: REQ-views-api-contract
type: req
---
The read API must be the openfief.views.v1.ReadService protobuf service with HTTP annotations, served from one set of handlers as Connect and, transcoded by Vanguard, as REST under /v1, whose REST responses match the OpenAPI document generated from the same definitions: money and quantities as decimal strings, dates as YYYY-MM-DD, and errors mapped to not found, invalid argument or internal.

## Context

Refines the firm requirements REQ-views-read-api-protobuf (D-50), REQ-AUTO-F7E15E59C985C20D and REQ-AUTO-6BEF2198D156F4F0: D-6 names REST/JSON with an OpenAPI spec under /v1 and D-50 makes protobuf the source, served as Connect and REST through Vanguard, with OpenAPI generated (protoc-gen-connect-openapi, pinned in the Taskfile; task proto:check fails when the committed document or Go code is stale). Every view plugin uses this API (REQ-AUTO-A04740EC43D6F6BA). Decimal strings keep amounts exact (D-20, no float). Modeled by Claude Code for the M0 read API task; the author has not reviewed it yet.

## Source

> The read API (D-6) is defined in protobuf with HTTP annotations, like the plugin contracts (D-11). The core serves it both as Connect, for generated Go and TypeScript clients, and as REST under `/v1` through transcoding (Buf's Vanguard). The OpenAPI spec is generated from the same definitions.

docs/DESIGN.md D-50

> OpenFief's public API is read-only REST/JSON with an OpenAPI spec, versioned `/v1`, covering accounts, transactions, tax lots, holdings and valuations. Every view plugin, the built-in UI included, uses it.

docs/DESIGN.md D-6
