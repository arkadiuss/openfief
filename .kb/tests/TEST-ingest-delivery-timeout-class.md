---
title: The class delivery-timeout is named, parsed and picked from ErrDeliveryTimeout
status: active
sourceFile: internal/ingest/app/syncerror_test.go
verification_scope: unit
origin:
  kind: agent
  ref: the S9b brief, author decisions of 2026-11-14 on D-211, D-198, D-196 and D-18
  recorded_at: '2026-11-14T06:29:53.404Z'
id: TEST-ingest-delivery-timeout-class
type: test
---
TestSyncErrorClassOf checks that an error wrapping ErrDeliveryTimeout is the class delivery-timeout whatever else it wraps; TestSyncErrorClassNames in internal/ingest checks its name, its parsing and the contract's enum.