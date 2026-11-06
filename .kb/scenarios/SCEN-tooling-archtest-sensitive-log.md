---
title: Logging money or a raw error fails the architecture test
status: active
origin:
  kind: agent
  recorded_at: '2026-11-06T17:49:54.129Z'
id: SCEN-tooling-archtest-sensitive-log
type: scenario
---
Given slog calls that log a Money value, a Money's String(), an error and an error's Error() text, when the privacy rule runs, then each function is reported with what it logs; an error passed through kernel.LogSafe and an internal ID are not.
