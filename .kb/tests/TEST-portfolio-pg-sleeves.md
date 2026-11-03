---
title: PostgreSQL keeps sleeve kinds, sleeves and which account is in which sleeve in the portfolio schema
status: active
sourceFile: internal/portfolio/postgres/sleeves_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-03T11:48:54.904Z'
id: TEST-portfolio-pg-sleeves
type: test
---
Runs against a fresh database with the portfolio migrations under task test:pg. It checks that the built-in kinds are seeded, a kind keeps its default role even against a direct UPDATE, a sleeve needs a stored kind and an override exactly when its role differs, an account moves between sleeves and a sleeve not stored is refused by the foreign key, a sleeve with accounts and a kind in use are not deleted, and concurrent EnsureDefaults store the defaults once and keep the owner's sleeve.