---
title: 'The maintainer''s use cases over fakes: ask, work, weekly, decide and read'
status: active
sourceFile: internal/portfolio/app/maintainer_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-13T21:07:18.570Z'
id: TEST-portfolio-app-maintainer
type: test
---
Use-case tests with in-memory fakes of the store, the gate, the holdings, the asker, the Inbox items and the audit log: one run at a time, every ending as a state with its audit entry, fail-closed audit, two weekly failures, apply and dismiss and who may read.