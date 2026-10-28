---
title: Invalid index readings are refused as a whole
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T08:55:02.073Z'
id: SCEN-ingest-index-reading-valid
type: scenario
---
Given a delivery with index readings, when every reading names CPI or the reference rate, has a date and, for CPI, is dated on the first of its month, then all are stored with the instance as source; when any reading is invalid, the batch fails and nothing of it is imported or stored.