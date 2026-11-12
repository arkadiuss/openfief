---
title: test/e2e/chartcheck works out the UI's charts from read API answers and reads the UI's calls from the core's log
status: active
sourceFile: test/e2e/chartcheck/chartcheck_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-08T07:27:23.834Z'
id: TEST-e2e-chartcheck
type: test
---
Runs offline under task test. Table tests of test/e2e/chartcheck, which works out what the built-in UI's charts must draw from read API answers: ChartFrom takes the performance's first day unless it is empty or after the date and never goes more than 3650 days back; ValueChart draws a day per total and a line per value kind, gross, net and liquidation, with an empty value where a total lacks one; PerformanceChart one line of indexes; PriceCharts one chart per listing in order of appearance, the days of every source sorted, the primary source first and the others by name, gaps empty, and the first source by name without a primary; Allocation a bar per group with its value and its share rounded half-even to one place, none for a zero total, refusing amounts that are no decimal. PlainMoney reads amounts written with a narrow-space, comma or point group and a decimal comma or point back as exact decimals and refuses text without a currency or a decimal. Calls parses the core's JSON log for read API requests and plugin calls, skipping other lines and failing on a bad time; PluginCalls and ReadRequests count within a window, both ends included, and Disallowed allows a whole service or one procedure without matching a service named like it.