---
title: An oversized connector answer or delivery is refused
status: active
origin:
  kind: agent
  recorded_at: '2026-11-02T19:45:17.170Z'
id: SCEN-ingest-connector-answer-bounded
type: scenario
---
Given a connector instance, when its answer to a sync is longer than 32 MiB, then the call fails with RESOURCE_EXHAUSTED before the answer is decoded; when its delivery holds more than 500,000 items, postings counted with the rest, then the whole delivery is refused with ErrDeliveryTooLarge before any item is translated, and nothing of it is imported.