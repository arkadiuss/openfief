---
title: SyncService.RetrySync answers each refusal with the code its contract names
status: active
sourceFile: internal/ingest/connectapi/syncs_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-06T04:22:54.912Z'
id: TEST-ingest-connectapi-syncs
type: test
---
Starts the SyncService over httptest with the view-plugin interceptor and calls it with a generated client: the owner's retry succeeds; another caller and a read session get PERMISSION_DENIED, an unknown session UNAUTHENTICATED, an unlisted instance NOT_FOUND, a running sync FAILED_PRECONDITION, no scheduler UNAVAILABLE, any other failure INTERNAL, and a second retry within the minute RESOURCE_EXHAUSTED.