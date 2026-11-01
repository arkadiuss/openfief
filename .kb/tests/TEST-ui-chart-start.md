---
title: The overview and an account's page start their time charts on the start the owner chose, by default where tracking began
status: active
sourceFile: internal/views/ui/chartstart_test.go
origin:
  kind: agent
  recorded_at: '2026-11-01T13:01:22.561Z'
id: TEST-ui-chart-start
type: test
---
Handler tests of the built-in UI against in-memory Connect fakes of the core: the overview with start in its URL asks GetPerformance and GetModelPortfolio from that day, draws the value chart from it, offers a link back to the default and carries the start and the tracking mark into the value chart's switch; without it there is no such link; an account's page asks GetPerformance with the account's ID and the chosen start, draws the account's value and performance with the start control, whose help names the account's first day as the default; a malformed start is refused with 400 on both pages.