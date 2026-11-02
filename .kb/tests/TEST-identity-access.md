---
title: 'The rule until SpiceDB answers: the owner views and changes, every other holder of read access only views'
status: active
sourceFile: internal/identity/access_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-01T22:45:52.605Z'
id: TEST-identity-access
type: test
---
TestMayAct runs identity.MayAct over every holder, the zero caller and an unknown holder, with view, change and no action, and TestActionAndHolderNames checks the names the schema uses. A rule that let a script token change, or the zero caller view, would fail it.