---
title: The read API refuses compressed and oversized requests before reading or authenticating them
status: active
sourceFile: internal/views/connectapi/limits_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-02T19:45:08.400Z'
id: TEST-views-connectapi-request-limits
type: test
---
TestReadAPIBoundsBodies serves the read API handler in process through httptest.NewRecorder with no bearer token: a small gzip bomb that inflates to 8 MiB over Connect unary and REST, and Connect-Content-Encoding and Grpc-Encoding requests, must get 415 having read none of a lazily made body; a declared length over the cap 413 without a read; and a body of unknown length must stop being read within 2 MiB and answer resource_exhausted. A response of 401, or unauthenticated, would mean the token was checked first. Removing the bounds makes every case fail, the read count rising to the whole 8 MiB body.