---
title: The install finds institutions another instance books for
status: active
sourceFile: internal/plugins/app/institutions_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-03T07:22:14.804Z'
id: TEST-plugins-app-institutions
type: test
---
Over an in-memory instance store with three installed instances: SharedInstitutions returns the institution another instance already books for, with that instance, skips the instance being installed or upgraded itself, and returns nothing for a plugin that declares no institution, without listing the store needlessly.