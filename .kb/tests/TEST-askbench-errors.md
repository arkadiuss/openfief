---
title: Status labels, all calls failed exits 2, partial outage scored with its errors listed, no key or provider text in any output
status: active
sourceFile: tools/askbench/errors_test.go
verification_scope: integration
origin:
  kind: agent
  ref: Brief to the Ask bench status session (a run no model answered must not look like a model result), 2026-11-14
  recorded_at: '2026-11-14T13:26:23.822Z'
id: TEST-askbench-errors
type: test
---
Runs the command against an httptest server that answers 402, 401, 429 or 500 to every call, or to the first five, and checks the labels of the statuses and of a timeout, the exit code 2 and the could-not-run text in stdout, summary.md and report.json, the per-status counts, the errored runs left unscored with no failure reason, and that neither the key nor the provider's words are in any output.
