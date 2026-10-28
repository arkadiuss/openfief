---
title: The NBP FX plugin continues from its watermark
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T17:06:56.729Z'
id: SCEN-plugins-nbp-fx-watermark
type: scenario
---
Given the NBP FX plugin configured to start on 2025-01-01, when the core syncs it on 2025-01-03 without a watermark, then it delivers the rates of 2 and 3 January and returns watermark 2025-01-03; when the core syncs again on 2025-01-06 with that watermark and NBP published nothing since, it delivers nothing and keeps the watermark; when NBP is unreachable the call is Unavailable and the core retries the batch.