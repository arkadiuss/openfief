---
title: The read API's OpenAPI document lists exactly the read service's routes, each with a client method
status: active
tags:
  - acceptance:p10-7
origin:
  kind: agent
  recorded_at: '2026-11-08T04:06:58.429Z'
id: SCEN-e2e-read-api-every-resource
type: scenario
---
Given the core installed from the demo's manifests and the data path's record, when the suite compares the OpenAPI document this checkout publishes (gen/openapi) with the google.api.http routes of the protobuf ReadService, then both list the same GET operations with the same paths and operation IDs, the Go SDK's generated ReadServiceClient has a method for every one of them, and accounts, transactions, tax lots, holdings and valuations list the data path's IBKR, XTB and bond accounts' records.