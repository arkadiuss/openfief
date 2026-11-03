---
title: htmx keeps no page in its history cache
status: active
origin:
  kind: agent
  recorded_at: '2026-11-03T16:30:43.722Z'
id: SCEN-ui-htmx-keeps-no-history
type: scenario
---
Given any page of the UI, when htmx pushes it into the history, then its config has the history cache off and refreshes a page missing from the history, so no snapshot of the owner's record stays in sessionStorage (D-151).