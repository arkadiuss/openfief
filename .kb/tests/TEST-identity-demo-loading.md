---
title: 'The demo rules: demo sessions with the flag and the marker, demo data loading with the flag alone'
status: active
sourceFile: internal/identity/demo_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-03T19:03:54.688Z'
id: TEST-identity-demo-loading
type: test
---
TestDemoDeployment runs identity.DemoSessionsAvailable and identity.DemoLoading over every pairing of the demo sessions flag and the demo marker as a table: demo sessions only with both, the demo data loading only with the flag and no marker, nothing without the flag. A false pass would need both rules to agree on a wrong pairing.