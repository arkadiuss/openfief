---
title: The day a period opens on the charts is the close of the day before its first day, for every period, and the year's opening comes from the performance series
status: active
origin:
  kind: agent
  ref: 'Design session, on the author''s behalf: D-223, a period''s return runs from the close of the day before its first day and the chart opens there, 2026-11-15'
  recorded_at: '2026-11-15T18:11:58.364Z'
sourceFile: internal/views/ui/home_unit_test.go
verification_scope: unit
id: TEST-ui-period-opening
type: test
---
Table tests of periodOpening for this year, this month, the last 12 months, last year, a calendar year, a month, since a day and everything (no opening: the charts' default start), and of yearReturn's opening day: the last day dated on or before 31 December, the first day when the series starts later, none without days, with the return in the owner's number format. A false pass would need periodRange to be wrong in the same way, but its own table test fixes each period's days.