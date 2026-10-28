---
title: The model portfolio service admits view plugins and owner sessions and maps errors to Connect codes
status: active
sourceFile: internal/portfolio/connectapi/service_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T23:10:15.094Z'
id: TEST-portfolio-connectapi
type: test
---
An in-memory Connect server with a stand-in view-plugin interceptor and the real use cases over a memory store: the owner saves an instrument and an index target and reads them back unchanged; a caller that is not a view plugin and a read session get permission_denied, an unknown session unauthenticated, weights not summing to 100, a malformed weight, an index without a currency, no kind and a bad ISIN invalid_argument, and a failing store internal.