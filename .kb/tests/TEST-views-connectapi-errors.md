---
title: The read API's REST errors are the responses its OpenAPI document declares, with a google.rpc.Status body
status: active
sourceFile: internal/views/connectapi/errors_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-08T05:43:33.902Z'
id: TEST-views-connectapi-errors
type: test
---
Runs offline under task test through the real Vanguard transcoder over in-memory fakes. TestRESTErrorsMatchContract sends requests without a token and with one the verifier rejects (401 with the Bearer challenge), for an unknown account, instrument and connector instance and for the model portfolio while it is off (404), and with a reversed date range, a malformed date and a page token the API never returned (400); each response validates with kin-openapi's openapi3filter against the generated document and decodes strictly into code, message and an empty details list with the expected gRPC code and fixed text. TestOpenAPIDeclaresErrors walks every operation of the document for its 400, 401 and, with a path parameter, 404 references, GetModelPortfolio's being ModelPortfolioOff, each a google.rpc.Status. TestOpenAPIRefusesWrongErrors shows the check can fail: a 404, 401 or 400 body carrying another code, a string code or details that are no list does not validate.