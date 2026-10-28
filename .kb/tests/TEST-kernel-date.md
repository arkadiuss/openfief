---
title: Date is a calendar date in ISO 8601 form with no time zone
status: active
sourceFile: internal/kernel/date_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T04:48:23.475Z'
id: TEST-kernel-date
type: test
---
Table tests parse ISO dates (including the 2024 leap day) and reject 2026-02-29, a timestamp with a time, day-first text and the empty string; NewDate rejects 2026-04-31, month 13 and day 0 instead of normalizing them; AddDays and DaysUntil agree across a month end, a leap day, a year end and the March 2026 European daylight-saving change; Compare, Before, After, Equal and == agree. A false pass would be arithmetic in a local time zone: the DST case would then lose or gain a day.