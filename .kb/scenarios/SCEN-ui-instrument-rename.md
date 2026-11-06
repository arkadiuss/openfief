---
title: The owner renames an instrument from its page
status: active
origin:
  kind: agent
  recorded_at: '2026-11-06T10:42:57.465Z'
id: SCEN-ui-instrument-rename
type: scenario
---
Given the owner on an instrument's page, when they save a name, then it reaches the core trimmed and the page says it is saved; a refused name shows the core's reason; an empty or overlong name never reaches the core; a read session is refused.