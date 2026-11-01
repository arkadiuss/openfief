---
title: The charts' default start is the latest seeding day, accounts with their whole history don't move it, and a chosen start is kept
status: active
sourceFile: internal/views/chartstart_test.go
origin:
  kind: agent
  recorded_at: '2026-11-01T13:01:19.832Z'
id: TEST-views-chart-start
type: test
---
Table tests of the pure rules in internal/views/chartstart.go on synthetic transactions and histories: HistoryOf finds an account's first day and the latest day of an opening lot or opening cash; ChartStart returns the latest seeding day of the seeded accounts, ignores accounts with their whole history whether they opened earlier or later, falls back to the first day of any account when none was seeded, and reports no start for no accounts; ChosenStart keeps a chosen day, takes the default for none, and moves a day before the first transaction to it. A false pass would need the fixtures' days to coincide, which the cases avoid.