---
title: Every sync attempt is recorded with an error class and none of its text
status: active
origin:
  kind: agent
  recorded_at: '2026-11-05T23:38:59.322Z'
id: SCEN-ingest-sync-attempt-history
type: scenario
---
Given an instance whose first sync attempt fails because the connector can't be reached and whose retry succeeds, when both attempts end, then the history holds the failure with class unavailable, one failure in a row and the retry's due time, and the success with no class; an error is classified from the known errors in its chain, never from its text, and the history refuses updates and deletes.