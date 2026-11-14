---
title: The maintainer on a real core with River and the audit log
status: active
sourceFile: cmd/core/maintainer_core_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-13T21:07:21.893Z'
id: TEST-core-maintainer-real-core
type: test
---
Tests on a core over PostgreSQL (task test:pg): with everything off an ask is refused and writes nothing; with the switch, the toggle, a route and the model portfolio on, an ask queues a run that River works, which ends without a note for want of figures, audited once.