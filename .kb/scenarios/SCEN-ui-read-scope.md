---
title: A read-scoped session sees no owner actions
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T19:49:05.124Z'
id: SCEN-ui-read-scope
type: scenario
---
Given a read-scoped session such as the demo's, when it opens the overview, then no settings link is shown and every settings page and action answers 403 without calling the core; an owner session sees the settings link.