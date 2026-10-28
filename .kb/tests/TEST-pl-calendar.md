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
Table tests compute maturity from synthetic purchase dates (EDO ten years on, OTS three months, a leap-day TOS ending Feb 28), reject a purchase outside the month the series code names with ErrSeriesMismatch, step interest periods across month ends as the ROR1027 issue letter's schedule prints them (bought 31.10.2026: periods start 30.11.2026, 28.02.2027, 31.03.2027) and across leap days, and map a period start to the CPI month two months earlier across a year boundary. A false pass would be date arithmetic that chains months: the 31st cases catch it.