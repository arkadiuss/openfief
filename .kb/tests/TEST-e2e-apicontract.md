---
title: The contract walker lists the OpenAPI operations and protobuf routes and builds REST targets, Connect requests and error codes
status: active
sourceFile: test/e2e/apicontract/apicontract_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-08T04:06:55.938Z'
id: TEST-e2e-apicontract
type: test
---
Runs offline under task test. It compares apicontract.Operations of the generated OpenAPI document with apicontract.Routes of the real ReadService descriptor, so the walker the end-to-end suite relies on agrees with both; maps synthetic services built from protobuf text (GET, POST, PUT, PATCH, DELETE, streaming, missing rule, custom verb, unknown, nested and unclosed variables); builds REST targets with escaped path variables and sorted queries and refuses missing or unknown fields; fills request messages with enums and integers as strings and refuses another method's input; and reads Connect error codes from unary bodies and streaming end-of-stream envelopes, refusing bodies without one. For the error-response walk, RESTStatus decodes a REST error body strictly as a google.rpc.Status and refuses Connect's string code, a missing or null details list, a missing message, another field and a body that is no JSON; ErrorResponse, walked over every route of the generated document, names BadRequest and Unauthorized for each, NotFound for each with a path parameter and ModelPortfolioOff for GetModelPortfolio, refuses 403, 500 and other undeclared statuses that only the default response covers, and refuses an operation the document lacks.