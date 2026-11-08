---
title: The read API's OpenAPI document must declare, each with a google.rpc.Status body, a 400 for invalid arguments and a 401 for a missing or invalid bearer token on every operation, a 404 on every operation with an ID in its path, and a 404 on GetModelPortfolio while the model portfolio is off, and the read API's REST error responses must match those declarations.
status: open
text_ref: design-decisions:D-200#7
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-200-7
  - lane:strict
semantic_text: The read API's OpenAPI document must declare, each with a google.rpc.Status body, a 400 for invalid arguments and a 401 for a missing or invalid bearer token on every operation, a 404 on every operation with an ID in its path, and a 404 on GetModelPortfolio while the model portfolio is off, and the read API's REST error responses must match those declarations.
logic_claims:
  - CLAIM-C46183F2A2C2A063
semantic_clauses:
  - The read API's OpenAPI document must declare, each with a google.rpc.Status body, a 400 for invalid arguments and a 401 for a missing or invalid bearer token on every operation, a 404 on every operation with an ID in its path, and a 404 on GetModelPortfolio while the model portfolio is off, and the read API's REST error responses must match those declarations.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 846706fb595b1a39f11545fe90bf77f9c14188d0c4b1ed994ea08da8d7cf5463
semantic_inventory:
  - claim_key: CLAIM-C46183F2A2C2A063
    claim_text: The read API's OpenAPI document must declare, each with a google.rpc.Status body, a 400 for invalid arguments and a 401 for a missing or invalid bearer token on every operation, a 404 on every operation with an ID in its path, and a 404 on GetModelPortfolio while the model portfolio is off, and the read API's REST error responses must match those declarations
    role: normative
    status: modeled
    span:
      start: 0
      end: 361
origin:
  kind: agent
  recorded_at: '2026-11-08T05:43:53.228Z'
id: REQ-views-read-api-declared-errors
type: req
---
The read API's OpenAPI document must declare, each with a google.rpc.Status body, a 400 for invalid arguments and a 401 for a missing or invalid bearer token on every operation, a 404 on every operation with an ID in its path, and a 404 on GetModelPortfolio while the model portfolio is off, and the read API's REST error responses must match those declarations.

## Context

D-200's criterion 7 has /v1 serve the record as its OpenAPI spec describes and reject unauthenticated calls, and D-6 makes the spec the public contract. The generated document declared only each operation's success, so a 401, a 404 or a 400 had no declared shape and the end-to-end suite could not validate one. The design session decided that every operation declares 401 and 400, every operation taking an ID declares 404, and GetModelPortfolio declares 404 while the model portfolio is off (D-52), all with the error body the REST mapping returns, generated deterministically and kept fresh by proto:check, and that the unit and end-to-end tests validate those responses against the spec. Over REST Vanguard writes a google.rpc.Status with the numeric gRPC code, not Connect's error JSON, so that is the declared schema. Modeled by Claude Code for the read API error responses; the author has not reviewed the wording.

## Source

> 7. **Read API.** `/v1` serves accounts, transactions, tax lots, holdings and valuations as its OpenAPI spec describes, rejects unauthenticated calls, and has no write endpoint. The Connect clients generated from the same definitions return the same data.

design-decisions:D-200#7 (docs/DESIGN.md, D-200, criterion 7)
