---
title: 'Every owner procedure, SaveThesis and EditThesis and SetOpinionsEnabled included, leaves exactly one audit entry of its '
status: active
sourceFile: cmd/core/ownerchanges_test.go
verification_scope: integration
origin:
  kind: agent
  ref: 'Implement D-225 and D-229: opinions and draft theses (design session brief, 2026-11-21)'
  recorded_at: '2026-11-21T15:00:00.000Z'
id: TEST-thesis-audit-core
type: test
---
Every owner procedure, SaveThesis and EditThesis and SetOpinionsEnabled included, leaves exactly one audit entry of its action and none without the outbox.