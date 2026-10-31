---
title: The IBKR Flex plugin maps pulled and uploaded Flex statements in the requested mode
status: active
sourceFile: plugins/ibkr-flex/connector_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T17:06:48.443Z'
id: TEST-plugins-ibkr-flex
type: test
verification_scope: integration
---
Tests drive the connector with a fake Flex Web Service fetch and the ibkrflex adapter's synthetic activity statement, with synthetic credentials written to a temporary Secret directory. Sync in opening and transactions modes delivers the mapped transactions for the batch, fetched with the Secret's token and query ID, and returns the given watermark; no mode is InvalidArgument; a missing token or empty query ID FailedPrecondition; IBKR failing Unavailable; an oversized or unreadable statement InvalidArgument; no error contains the token. Only opening mode yields opening lots. ImportFile through the SDK server maps a chunked upload and refuses a missing mode or non-Flex file. TestManifest checks the manifest's capability, egress host and Secret mount. A false pass would be credentials read from the wrong place; the recorded fetch arguments guard it.