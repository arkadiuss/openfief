---
title: Dates are calendar dates without a time zone
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T04:48:31.680Z'
id: SCEN-kernel-calendar-date
type: scenario
---
Given the texts 2024-02-29, 2026-02-29 and 2026-03-31T10:00:00Z, when each is parsed as a date, then the leap day is accepted and printed back unchanged and the other two are rejected as invalid dates; and adding one day to 2026-03-31 gives 2026-04-01 whatever the local time zone of the machine.