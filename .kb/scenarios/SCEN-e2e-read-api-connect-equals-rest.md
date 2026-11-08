---
title: The Go SDK's generated Connect client returns the same data as the REST API
status: active
tags:
  - acceptance:p10-7
origin:
  kind: agent
  recorded_at: '2026-11-08T04:07:02.217Z'
id: SCEN-e2e-read-api-connect-equals-rest
type: scenario
---
Given the data path's record, when the suite makes every call of the conformance scenario both through the Go SDK's generated ReadServiceClient over the Connect protocol and over REST, with the same script token and the same request fields, then the REST body, decoded strictly into the response message, equals the Connect response field for field, a pair that differs being made again up to five times while the record changes; the model portfolio, off, is not_found over Connect and 404 over REST.