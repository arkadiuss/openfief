---
title: A sync delivers the CPI and reference-rate readings after the watermark, dated by month
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T03:56:27.771Z'
id: SCEN-treasurybonds-index-sync
type: scenario
---
Given the treasury-bonds connector's watermark and today, when the core syncs it, then it delivers GUS's final CPI for each month after the CPI watermark, at most four years of it per sync, and for each month after the reference-rate watermark whose reference-rate day is over the NBP rate in force on that day, each reading dated on its month's first day, and returns the last month of each as the new watermark; a foreign watermark or an outage delivers nothing.