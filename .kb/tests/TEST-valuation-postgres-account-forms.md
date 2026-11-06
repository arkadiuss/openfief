---
title: Account forms and earlier payments round-trip in PostgreSQL and go with a deleted account
status: active
sourceFile: internal/valuation/postgres/accountforms_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-01T20:10:18.283Z'
id: TEST-valuation-postgres-account-forms
type: test
---
Against a fresh database with every valuation migration applied (task test:pg), an account has no form and no earlier payments at first, each saved value reads back and a second save replaces the first, and ForgetAccount removes both. It runs real SQL, so a missing column or table fails it.