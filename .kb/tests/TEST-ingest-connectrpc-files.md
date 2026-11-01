---
title: The core's file import tells a connector's refusal, missing support and unavailability apart, and stops on a broken file
status: active
sourceFile: internal/ingest/connectrpc/files_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T04:58:01.167Z'
id: TEST-ingest-connectrpc-files
type: test
verification_scope: integration
---
Calls a real in-process Connect server over h2c: a connector answering INVALID_ARGUMENT gives ErrEntryRefused with its reason, UNIMPLEMENTED ErrNoEntries, UNAVAILABLE or no address ErrConnectorUnavailable. A file that breaks off after three chunks returns the read's error and the connector never completes the call, so it maps nothing. Synthetic tokens and bytes only.