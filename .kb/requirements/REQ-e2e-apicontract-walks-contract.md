---
title: The end-to-end suite's contract walker must list the read API's OpenAPI operations and the HTTP routes its protobuf service's google.api.http annotations declare, build each call's REST target and Connect request from the same request fields, and read the Connect error code of a refused call.
status: open
text_ref: docs/DESIGN.md#D-200
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-200
  - lane:strict
semantic_text: The end-to-end suite's contract walker must list the read API's OpenAPI operations and the HTTP routes its protobuf service's google.api.http annotations declare, build each call's REST target and Connect request from the same request fields, and read the Connect error code of a refused call.
logic_claims:
  - CLAIM-05B4C19D1D4FAB21
semantic_clauses:
  - The end-to-end suite's contract walker must list the read API's OpenAPI operations and the HTTP routes its protobuf service's google.api.http annotations declare, build each call's REST target and Connect request from the same request fields, and read the Connect error code of a refused call.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: d53e030554265a016728e021d9ccfeea27fc957d7c7ac2c7bbfcef4c303534f0
semantic_inventory:
  - claim_key: CLAIM-05B4C19D1D4FAB21
    claim_text: The end-to-end suite's contract walker must list the read API's OpenAPI operations and the HTTP routes its protobuf service's google.api.http annotations declare, build each call's REST target and Connect request from the same request fields, and read the Connect error code of a refused call
    role: normative
    status: modeled
    span:
      start: 0
      end: 292
origin:
  kind: agent
  recorded_at: '2026-11-08T04:07:11.659Z'
id: REQ-e2e-apicontract-walks-contract
type: req
---
The end-to-end suite's contract walker must list the read API's OpenAPI operations and the HTTP routes its protobuf service's google.api.http annotations declare, build each call's REST target and Connect request from the same request fields, and read the Connect error code of a refused call.

## Context

D-200 criterion 7 is checked end to end (test/e2e/readapi_test.go): every OpenAPI operation called over REST, compared with the protobuf service's routes (D-50), and with the same call through the Go SDK's Connect client. The walker is the pure part of that check, so it is unit-tested offline under task test (docs/dev/testing.md). Modeled by Claude Code for D-200 criterion 7's end-to-end scenarios; the author has not reviewed the wording.

## Source

> 7. Read API. /v1 serves accounts, transactions, tax lots, holdings and valuations as its OpenAPI spec describes, rejects unauthenticated calls, and has no write endpoint. The Connect clients generated from the same definitions return the same data.

docs/DESIGN.md#D-200
