---
title: The operator runs nothing for an instance naming another running instance's Secret
status: active
sourceFile: internal/plugins/operator/shared_secret_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-03T08:22:04.909Z'
id: TEST-operator-shared-secret
type: test
---
testSharedSecret runs under TestReconcile against envtest's API server: Plugin clash-a with role token runs; Plugin clash with role a-token and a grant for clash-a-token, as GitOps would apply it, gets Ready=False/SecretHeld with a deny-all NetworkPolicy and no Deployment while clash-a keeps running; after clash-a is deleted, clash runs.