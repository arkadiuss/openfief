---
title: Every performance period is the one period card
status: active
origin:
  kind: agent
  ref: 'Design session decision, on behalf of the author: every performance period answers with the period card and one convention, a gap against D-61, D-194 and D-203, 2026-11-15'
  recorded_at: '2026-11-15T15:00:00.000Z'
id: SCEN-ui-ask-every-period-one-card
type: scenario
---
Given a core with its latest valued day 2026-03-31 and tracking begun on 2026-03-30, when the box is asked, typed and through the performance tool, for this year, this month, the last 12 months, last year, 2025, 2026, since 2026-01-15, everything, November 2025 and 2025-11-01 to 2025-11-30, then each asks GetPeriodChange for the days its period resolves to (this year from 2026-01-01, the last 12 months from 2025-04-01, everything from 2026-03-30) and answers with the same card, the return, the start and end values and the deposits and withdrawals, from the end of the day before its first day; and with nothing recorded the period of everything says it has no valued days and asks the core nothing.