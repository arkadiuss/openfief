---
title: The operator waits for a grant to approve a connector's institutions
status: active
sourceFile: internal/plugins/operator/institutions_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-03T07:22:12.156Z'
id: TEST-operator-institutions
type: test
---
Against envtest's API server with the real reconciler: a connector declaring ibkr under a grant naming an empty list waits with ibkr in status.pendingApproval and nothing deployed, then runs with ibkr in status.effective once the grant approves it; a grant without the institutions field, as one made before D-185, keeps the running version and its institutions, and a new version declaring xtb as well waits with xtb pending while the old version keeps running.