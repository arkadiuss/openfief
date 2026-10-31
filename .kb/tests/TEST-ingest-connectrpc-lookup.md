---
title: The core's client asks a connector to look tickers up and checks its proposals
status: active
sourceFile: internal/ingest/connectrpc/lookup_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-10-31T12:54:26.551Z'
id: TEST-ingest-connectrpc-lookup
type: test
---
Serves fake connectors over HTTP/2 in process and calls connectrpc.Client.LookupInstruments: a proposal for an asked ticker with a valid ISIN comes back, a repeated one keeps the first; a connector without the RPC wraps ErrNoEntries, an unreachable instance ErrConnectorUnavailable, and a proposal for a ticker not asked or with a wrong check digit is ErrInvalidMessage.