---
title: The core sends bond lots over the connector contract exactly and tells refusals from absent support
status: active
sourceFile: internal/ingest/connectrpc/bonds_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T01:13:44.931Z'
id: TEST-ingest-connectrpc-bonds
type: test
---
Calls a real in-process Connect server over h2c: ImportBondLots sends the lot's fields as exact strings (rate 0.068, fee 2.00 PLN) with the instance's core token and reads the delivered bond terms. A connector answering INVALID_ARGUMENT gives ErrEntryRefused carrying its reason, UNIMPLEMENTED gives ErrNoEntries, UNAVAILABLE, no token or no address give ErrConnectorUnavailable. Synthetic tokens and lots only.