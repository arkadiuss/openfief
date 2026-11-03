---
title: The connector client's codec counts an answer's messages on the wire and refuses one over the cap
status: active
sourceFile: internal/ingest/connectrpc/codec_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-03T05:55:00.022Z'
id: TEST-ingest-connectrpc-codec
type: test
---
TestBoundedCodecCounts unmarshals SyncResponse wire bytes built with protowire holding one transaction of empty postings: exactly MaxDeliveryMessages messages decode with every posting, one more fails with ErrDeliveryTooLarge and leaves the message undecoded. TestBoundedCodecRoundTrip checks that what the codec marshals it decodes unchanged with an unknown field skipped, and that truncated bytes and non-messages are refused. TestCountElementsRepeatedScalars counts unpacked strings and packed varints and fixed values.