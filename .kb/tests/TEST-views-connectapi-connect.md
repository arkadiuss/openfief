---
title: The Connect service answers generated clients, is read-only by definition, and its enums match the domain
status: active
sourceFile: internal/views/connectapi/connect_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T08:43:03.633Z'
id: TEST-views-connectapi-connect
type: test
verification_scope: integration
---
A generated viewsv1connect client calls the read API over httptest with a bearer token: accounts, holdings, the valuation status (valued through 2026-03-02, running), the allocation by currency (1569.21 PLN: stored valuations plus USD cash at 4.00) and two transaction pages joined by the next page token. Without a token the call is unauthenticated, an unknown account not found, and a malformed date, an unspecified dimension or an unknown assertion status invalid argument. Connect's GET form answers for a no-side-effects method. TestServiceIsReadOnly walks the ReadService descriptor: every method is unary, declares no side effects and maps to a GET under /v1 without a body. TestEnumsMatch checks that every domain enum value (sleeve role, ledger account kind, transaction kind, assertion status, valuation basis, allocation dimension) has the protobuf value of the same name and number, so the adapter's numeric conversion is safe.