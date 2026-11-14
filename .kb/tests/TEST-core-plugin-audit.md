---
title: The core enters the operator's installs, upgrades and removals, and refuses to audit what it can't
status: active
sourceFile: cmd/core/pluginaudit_test.go
verification_scope: integration
origin:
  kind: agent
  ref: the S8 brief, from D-211, 2026-11-14
  recorded_at: '2026-11-14T02:39:39.843Z'
id: TEST-core-plugin-audit
type: test
---
Under task test:pg, runs the core's observation loop over a scripted plugin cache and the real store: a cache that has not synced is asked again, then the install, the upgrade and the removal are entries of the operator, once each. TestOwnerSessionsRefuseToAuditWhatTheyCannot checks the owner sessions' Audited refuses a demo session, no session, an unknown action or reference kind, and an auth that can't record.