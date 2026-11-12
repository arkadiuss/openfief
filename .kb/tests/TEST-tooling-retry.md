---
title: retry.sh retries a download only when a later try can fix it
status: active
sourceFile: scripts/tests/retry.test.sh
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-12T15:11:36.927Z'
id: TEST-tooling-retry
type: test
---
Runs scripts/retry.sh around a fake flaky command that fails with scripted error text, and a fake sleep that records the waits: a 429, any 5xx, a curl connection, DNS or timeout error, go's proxy answering 503, npm's connection reset and gh's HTTP 502 are retried with waits of 2, 4, 8, 16 and 32 seconds and give up after six tries; a 400, 401, 403, 404, 410, a certificate problem and a failure that is not the network's are final at once with the command's own exit status; arguments with spaces arrive whole. A false pass would be a rule that retries a 404 or never gives up, which the call counts catch.
