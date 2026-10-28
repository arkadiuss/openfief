---
title: A sleeve's role defaults from its kind, and only an explicit override makes a defensive kind offensive
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T09:34:30.171Z'
id: SCEN-portfolio-sleeve-role-override
type: scenario
---
Given the built-in kinds, when a sleeve of kind core is saved with no role, then it becomes defensive, and of kind agent-pool, offensive; when a satellite sleeve is saved as defensive, then it is accepted; when a retirement sleeve is saved as offensive without the explicit offensive override, then it is refused, and with the override it is accepted; when the override is set on any other sleeve, then it is refused, so the flag always means one thing.