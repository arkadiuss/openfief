---
title: Instruments to classify are a check of the attention store that resolves itself
status: active
sourceFile: internal/attention/app/unclassified_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-01T20:10:00.859Z'
id: TEST-attention-app-unclassified
type: test
---
UnclassifiedInstruments over in-memory fakes of the ledger, the values and the stored valuations: two OKI accounts holding the same synthetic funds give one condition per ISIN, while a treasury bond, a taxable account's fund, a fund held only in an earlier tax year and an account never revalued give none, and the owner's class removes it; a sweep raises the item and, after the owner classifies the fund, the next sweep resolves it. A false pass would need the fakes to hide a holding from the check.