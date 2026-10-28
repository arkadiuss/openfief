---
title: The core's command line, token listing and health answers before it is ready
status: active
sourceFile: cmd/core/main_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T14:32:44.309Z'
id: TEST-core-unit
type: test
---
Table tests parse every command and flag with a fake environment: serve is the default, with PLN, Europe/Warsaw, end of day 23:30, two workers, net if redeemed and default sleeves on; token issue, list and revoke and migrate take their arguments; a missing OPENFIEF_DSN, unknown commands and flags, extra arguments, a market or unknown bond basis, an unknown time zone and malformed end-of-day times fail. The token table shows ID, name, scope, times and active, expired or revoked status for a fixed time. The health handlers answer /healthz 200 and /readyz and the API 503 while starting, and /readyz 503 while draining, all uncacheable. A false pass would be a parser that ignores flags: each flag's value is compared.