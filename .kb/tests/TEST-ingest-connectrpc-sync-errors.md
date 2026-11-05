---
title: The Connect client wraps a failed Sync so it is classified as unreachable or as the connector's own error
status: active
sourceFile: internal/ingest/connectrpc/syncerror_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-05T23:38:56.902Z'
id: TEST-ingest-connectrpc-sync-errors
type: test
---
Starts a Connect server over HTTP/2 without TLS with connectors that answer UNAVAILABLE, UNIMPLEMENTED or PERMISSION_DENIED, and an instance with no address, and checks the class app.SyncErrorClassOf gives each error the client returns. The connector's reason is synthetic and never inspected.