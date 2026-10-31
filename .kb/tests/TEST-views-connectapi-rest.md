---
title: REST under /v1, transcoded by Vanguard, matches the generated OpenAPI, is GET-only, authenticated and logs no data
status: active
sourceFile: internal/views/connectapi/handler_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T08:42:58.462Z'
id: TEST-views-connectapi-rest
type: test
verification_scope: unit
---
httptest drives connectapi.New over app.Views with fake ports and a fake TokenVerifier. TestREST sends 28 GET requests under /v1 (every operation's 200, unknown accounts and paths 404, malformed dates, page sizes, page tokens, reversed or too long ranges and a missing dimension 400) and validates each 200 response against gen/openapi/openfief/views/v1/read_service.openapi.yaml with kin-openapi's openapi3filter, so the served JSON cannot drift from the document generated from the protos. TestRESTUnauthenticated checks no token, another scheme, an empty token and an unknown token give 401 with the Bearer challenge and no data, an uncheckable token 500, and the scheme in any case. TestRESTReadOnly sends POST, PUT, PATCH and DELETE with and without a token and expects 405 with Allow: GET; TestOpenAPIIsReadOnly checks the document has only GET operations under /v1 and no request body. Other tests check security headers and the request ID on every response, that internal errors name the request ID and not the error, that money is an exact decimal string with its original currency next to the reporting-currency value, and that the log never contains account IDs, institutions, instruments, amounts, currencies, tokens, error texts or paths.