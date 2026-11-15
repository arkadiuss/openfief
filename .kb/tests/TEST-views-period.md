---
title: A period's change is computed by a pure function from two totals and the performance series
status: active
origin:
  kind: agent
  ref: 'Design session, on the author''s behalf: the Ask box''s performance and value tools answer for a bounded period, a gap against D-194 (D-203: the read API answers first), 2026-11-15'
  recorded_at: '2026-11-15T12:27:09.949Z'
sourceFile: internal/views/period_test.go
verification_scope: unit
id: TEST-views-period
type: test
---
Table tests of views.ChangeOver over synthetic totals and series: a rise from a deposit while the return falls, withdrawals and deposits told apart by day, the opening day's own flow left out, missing net or liquidation left unset, no percentage from a start of nothing, half-even ties, a series that does not run from the opening day to the last refused, another currency refused. A false pass would need the figures computed by the test itself, which they are not: every want is written out.