---
title: Every read API error over REST is a response its operation declares in the OpenAPI document, with a google.rpc.Status body
status: active
origin:
  kind: agent
  recorded_at: '2026-11-08T05:43:44.222Z'
id: SCEN-views-read-api-declared-errors
type: scenario
---
Given the read API's OpenAPI document generated from its protobuf, with the error responses tools/openapierrors adds, when a client calls an operation over REST without a token or with one the core never issued, with an account ID, ISIN or connector instance no record has, for the model portfolio while it is off, or with an invalid argument such as a date range that ends before it starts, then it answers 401, 404 or 400, and that status is a response the operation itself declares, not only its default: 400 and 401 on every operation, 404 on every operation with an ID in its path, and ModelPortfolioOff as GetModelPortfolio's 404. The body is the google.rpc.Status Vanguard writes, code, message and an empty details list, with the declared gRPC code (3 or 9, 16, 5), and it validates against the document; a body with another code would not.