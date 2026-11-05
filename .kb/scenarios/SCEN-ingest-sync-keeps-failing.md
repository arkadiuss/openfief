---
title: A sync keeps failing after three failures in a row spanning fifteen minutes
status: active
origin:
  kind: agent
  recorded_at: '2026-11-05T23:39:00.641Z'
id: SCEN-ingest-sync-keeps-failing
type: scenario
---
Given an instance whose latest success was followed by failed attempts, when three of them in a row span at least fifteen minutes from the first one's start to the latest one's end, then its status keeps failing; two failures, or three within fifteen minutes, do not; and one success clears it.