---
title: Each change of the models settings is audited, and the providers are listed
status: active
sourceFile: internal/modelrouter/app/settings_audit_test.go
verification_scope: unit
origin:
  kind: agent
  ref: MS1 brief, 2026-11-13
  recorded_at: '2026-11-14T00:24:05.965Z'
id: TEST-models-app-settings-audit
type: test
---
Runs the settings use cases over a fake store, owner sessions, audit log and providers: every change writes its entry with the owner's session, a change the log can't record is not made, a refused change writes none, and the providers and broken route entries as the owner sees them.