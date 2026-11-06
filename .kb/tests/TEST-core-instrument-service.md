---
title: The core's plugin port serves the owner's renames
status: active
sourceFile: cmd/core/instruments_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-06T06:59:01.663Z'
id: TEST-core-instrument-service
type: test
---
Under task test:pg against the real core: the owner's view renames a synthetic ISIN; another caller, an unknown session and a malformed ISIN are refused with their codes.