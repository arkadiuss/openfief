---
title: A session and an address may ask the model only so often
status: active
origin:
  kind: agent
  ref: 'M5 brief from the design session: the demo''s spend guards (D-208), plus the M4 follow-ups, 2026-11-14'
  recorded_at: '2026-11-14T20:37:58.029Z'
id: SCEN-models-ask-question-limits
type: scenario
---
Given limits of 2 an hour and 30 a day, when a session asks twice then a third question is refused with the moment the first leaves the hour, a second session at the same address is refused by the address, another address is free, an unknown address leaves only the session's limit, a refused question is not counted, thirty questions at once let exactly the limit through, and the database holds a salted hash of the address and nothing older than a day.