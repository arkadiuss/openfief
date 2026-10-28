---
title: Only one sync of an instance runs at a time, within a timeout
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T16:50:14.729Z'
id: SCEN-ingest-single-flight
type: scenario
---
Given a sync of instance xtb-main that is running, when a tick finds it due again or the owner asks for a sync now, then no second sync starts and the request fails with ErrSyncRunning; when another core replica holds an unexpired claim, then this one doesn't sync; when an attempt exceeds its timeout, then it fails and is recorded as a failure.