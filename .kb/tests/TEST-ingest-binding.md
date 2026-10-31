---
title: Deliveries are checked against connector bindings and instance names are validated
status: active
sourceFile: internal/ingest/binding_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T16:49:56.510Z'
id: TEST-ingest-binding
type: test
verification_scope: unit
---
Table tests of the pure ingest package: CheckBound accepts an empty mapping and one whose transactions and balance assertions are all in bound accounts, and otherwise returns an UnboundAccountsError wrapping ErrAccountNotBound that names the instance and each unbound account once, sorted, whether a transaction or an assertion names it; the error message lists them. ValidateInstanceName accepts lower-case DNS labels of up to 56 characters and refuses empty, longer, digit-first, hyphen-last, upper-case and dotted names.