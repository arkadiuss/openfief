---
title: The core's guard flags are off by default, validated together and become the router's guards
status: active
sourceFile: cmd/core/guards_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-14T20:37:54.011Z'
id: TEST-core-guards
type: test
---
The serve flags: none set limits nothing, the demo's values parse, and a budget that is no number or negative, a negative limit, more questions an hour than a day, negative calls at once or wait, a header without proxies, proxies without a header and a proxy that is no network are refused; the budget and limits become the router's.