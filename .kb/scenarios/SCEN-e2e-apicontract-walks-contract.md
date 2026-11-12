---
title: The contract walker maps the read API's protobuf routes to its OpenAPI operations and calls
status: active
origin:
  kind: agent
  recorded_at: '2026-11-08T04:06:57.186Z'
id: SCEN-e2e-apicontract-walks-contract
type: scenario
---
Given the read API's protobuf service and its generated OpenAPI document, when the walker lists both, then they hold the same operations; given a call's request fields by JSON name, the walker builds its REST path and query and fills its Connect request from the same fields, refusing a field the method lacks; given a refused Connect call's body, unary or streaming, it returns the error code.