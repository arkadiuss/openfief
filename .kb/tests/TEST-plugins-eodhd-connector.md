---
title: The EODHD plugin syncs closes and index levels per series watermark through the connector contract
status: active
sourceFile: plugins/eodhd/connector_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T21:37:01.060Z'
id: TEST-plugins-eodhd-connector
type: test
---
Table tests call connector.Sync with a fake EODHD and a temporary Secret directory holding a synthetic API key: a first sync from the configured start delivers FAKEW.WAR closes in PLN, SYNT.LSE closes converted from pence to GBP under the configured symbol, and FAKE20 index levels, with a JSON watermark per ISIN and index; a later sync asks only after each mark; nothing new keeps the watermark; no start asks for the whole history; ambiguous, unlisted and unknown-ticker series become unsupported problems while the rest is delivered; EODHD failing is Unavailable, a bad watermark InvalidArgument, a missing key FailedPrecondition, none carrying the key. TestNextFrom covers the due-day rule; TestServedConnector serves it with the SDK, refuses a caller without the core's token and leaves ImportFile Unimplemented.