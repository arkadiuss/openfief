---
title: The maintainer's runs, notes and proposals on PostgreSQL
status: active
sourceFile: internal/portfolio/postgres/maintainer_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-13T21:07:19.936Z'
id: TEST-portfolio-postgres-maintainer
type: test
---
Tests on a disposable PostgreSQL (task test:pg): one active run by a unique index, stale runs abandoned, a run stored with its note and proposal whole or not at all, superseding, apply in one transaction and the table checks.