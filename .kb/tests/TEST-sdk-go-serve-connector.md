---
title: ServeConnector serves only the core's token and stops cleanly
status: active
sourceFile: sdk/go/openfief/run_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T17:06:41.115Z'
id: TEST-sdk-go-serve-connector
type: test
---
Tests in package openfief_test set the plugin's environment (listen port, instance, core subject and a key directory with a locally generated ECDSA key), run ServeConnector in the background and call Sync over HTTP: a token signed for the instance's audience is served, another token gets Unauthenticated, and cancelling the context makes ServeConnector return nil after a graceful shutdown. Without a listen port or an instance name it returns the matching environment error before serving. A false pass would be a server that ignores the verifier; the refused token guards it.