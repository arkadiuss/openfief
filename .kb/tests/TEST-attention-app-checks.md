---
title: Each check of the core finds its conditions
status: active
sourceFile: internal/attention/app/checks_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-10-31T18:30:36.128Z'
id: TEST-attention-app-checks
type: test
---
Feeds each app.Check fixed synthetic records through fakes of its ports: a flagged and a matched balance assertion, a confirmed transfer whose deposit was corrected and one that still matches, the latest price check per instrument, mapped and unmapped broker accounts, a pending ticker, a batch with unsupported rows, and IKZE or taxable accounts with and without the marginal rate. Each must return exactly the conditions listed, valid and of the check's kind.