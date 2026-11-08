---
title: Every read API operation answers as its OpenAPI document describes, and only with a valid token
status: active
tags:
  - acceptance:p10-7
origin:
  kind: agent
  recorded_at: '2026-11-08T04:06:59.764Z'
id: SCEN-e2e-read-api-openapi-conformance
type: scenario
---
Given the data path's record, when a script token calls every operation of the read API's OpenAPI document over REST under /v1 through a port-forward to the core's browser port, with the data path's three account IDs, three ISINs, its IBKR connector instance, the IBKR week's dates and every allocation dimension and value kind, then each answers 200 with a body kin-openapi validates against the document's schema for that operation and that decodes strictly into the protobuf response message, except the model portfolio, which is off and answers 404; and each operation answers 401 without a token and with a token the core never issued.