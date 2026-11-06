---
title: A failing sync shows its timeline and Retry now
status: active
origin:
  kind: agent
  recorded_at: '2026-11-06T04:32:11.152Z'
id: SCEN-ui-inbox-sync-timeline
type: scenario
---
Given a sync of xtb-main that keeps failing, when the owner opens the Inbox, then its item shows the failed attempts with their error classes, the last good sync and the next try, and Retry now, which the core starts; a refused retry shows the core's reason, and a read session sees the timeline without the button.