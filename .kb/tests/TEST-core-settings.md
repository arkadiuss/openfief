---
title: The core serves the owner's settings to view plugins, with the Polish defaults until the owner changes them
status: active
sourceFile: cmd/core/settings_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T22:58:56.263Z'
id: TEST-core-settings
type: test
---
PostgreSQL test through the real core (task test:pg) after a passkey login: GetOwnerSettings answers PLN, the space-comma default, four wrappers, seven formats and no marginal rate; after SetNumberFormat comma-dot and SetMarginalRate 0.32 it answers both; an unoffered format and a rate of 32 are INVALID_ARGUMENT, another caller PERMISSION_DENIED and an unknown session UNAUTHENTICATED.