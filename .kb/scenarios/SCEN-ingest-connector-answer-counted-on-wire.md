---
title: A connector answer with too many messages is refused before it is decoded
status: active
origin:
  kind: agent
  recorded_at: '2026-11-03T05:55:04.832Z'
id: SCEN-ingest-connector-answer-counted-on-wire
type: scenario
---
Given a connector instance, when its answer is longer than 16 MiB, then the call fails with RESOURCE_EXHAUSTED; when it is shorter but holds more than 640,000 messages and repeated elements, even empty ones of two bytes each, then the core counts them on the wire without allocating and refuses the answer with ErrDeliveryTooLarge before decoding it; when a decoded delivery holds more than 200,000 items, postings counted, then it is refused whole before any item is translated; and in every case the core's heap grows far less than its 256 MiB memory limit.