---
title: The UI's config sets the Ask box's limits
status: active
origin:
  kind: agent
  ref: 'M5 brief from the design session: the demo''s spend guards (D-208), plus the M4 follow-ups, 2026-11-14'
  recorded_at: '2026-11-14T20:38:11.931Z'
id: SCEN-ui-ask-loop-limits-config
type: scenario
---
Given a UI config with askToolRounds, askToolCalls, askOutputTokens and askDeadline, when the plugin starts then the Ask box runs with those limits, each within its maximum, and a value out of range or a deadline that is no duration stops it; with no config the limits are D-208's: five rounds, ten tool calls, 800 output tokens and two minutes.