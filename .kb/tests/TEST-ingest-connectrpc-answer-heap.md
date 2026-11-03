---
title: A hostile or the largest allowed connector answer grows the core's heap a bounded amount
status: active
sourceFile: internal/ingest/connectrpc/heap_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-03T05:55:02.415Z'
id: TEST-ingest-connectrpc-answer-heap
type: test
---
TestAnswerHeap reruns its own test binary once per case, so no other test allocates, and the child TestAnswerHeapChild serves the answer from an httptest server to Client.Sync and reports how far HeapSys grew: the byte cap filled with empty postings must stay within 64 MiB, the most empty postings the wire count lets through within 96 MiB, and 200,000 valid prices decoded and translated within 192 MiB, against the core's 256 MiB limit; before the count, the first case grew the heap by 1.8 GiB. TestClientTooLargeIsNotRefusal checks an entry call's oversized answer stays ErrDeliveryTooLarge and is never reported as the connector refusing the entries.