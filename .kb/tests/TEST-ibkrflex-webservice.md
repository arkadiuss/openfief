---
title: The Flex Web Service client fetches a statement, retries while it is generated, and never leaks the token
status: active
sourceFile: internal/ingest/ibkrflex/webservice_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T06:34:02.229Z'
id: TEST-ibkrflex-webservice
type: test
---
Replay tests with hypert play back synthetic, hand-written recordings in testdata/TestFetchStatement*: SendRequest answers a reference code, GetStatement first answers code 1019 (generation in progress) and then a FlexQueryResponse that maps to one buy and one cash assertion; an invalid-token answer (1015) fails without polling. The t query parameter is sanitized, and requests are validated against the recordings including the User-Agent. Fake transports cover a statement never ready within the attempt limit, a non-retryable code, a network failure, an HTTP error status, an unreadable body, a statement over the size cap and a cancelled context, asserting that no error message contains the token.