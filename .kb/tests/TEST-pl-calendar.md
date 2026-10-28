---
title: Bond maturity, interest periods and the CPI month follow the purchase date
status: active
sourceFile: internal/jurisdiction/pl/calendar_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T05:21:11.547Z'
id: TEST-pl-calendar
type: test
---
Table tests compute maturity from synthetic purchase dates (EDO ten years on, OTS three months, a leap-day TOS ending Feb 28), reject a purchase outside the month the series code names with ErrSeriesMismatch, step interest periods across month ends (Jan 31 then Feb 28 then Mar 31) and leap days, and map a period start to the CPI month two months earlier across a year boundary. A false pass would be date arithmetic that normalizes Feb 31 into March: the month-end cases would then shift.