---
title: Each EODHD sync delivers closes and index levels after each series' mark
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T21:37:14.306Z'
id: SCEN-plugins-eodhd-sync
type: scenario
---
Given an eodhd instance configured with two instruments and the FAKE20 index, start 2025-01-02 and an API key in its Secret, when the core syncs on 2025-01-03 without a watermark, then the delivery from institution eodhd holds both days' prices of each instrument in its listing's currency and FAKE20's two levels, and the watermark marks each series on 2025-01-03; when a series is unlisted, ambiguous or unknown, then it is an unsupported problem and the rest is delivered; when EODHD fails, then the sync is Unavailable and the core retries.