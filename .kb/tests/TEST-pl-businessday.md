---
title: Business days skip Saturdays and each year's statutory days off, and pick the reference rate's day
status: active
sourceFile: internal/jurisdiction/pl/businessday_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T03:45:48.328Z'
id: TEST-pl-businessday
type: test
---
Table tests over real calendar dates: TestIsBusinessDay checks weekdays, Saturdays, Sundays, every fixed holiday, Easter Monday in 2008, 2025, 2026 and 2038 (earliest and latest Easter), Corpus Christi, 6 January before and after 2011 and 24 December before and after 2025, and ErrOutsideCalendar before 1991. TestReferenceRateDay counts ten business days back for months whose count crosses Easter Monday, Corpus Christi and the three Christmas days, each worked out by hand in a comment. TestRateInForce picks from synthetic out-of-order changes, the effective day included. A false pass would need a hand count and the code to share the same holiday mistake.