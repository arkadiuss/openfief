---
title: Instances whose names and roles once spelled one Secret both run, and the API server refuses -- in names and roles
status: active
sourceFile: internal/plugins/operator/shared_secret_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-03T08:22:04.909Z'
id: TEST-operator-shared-secret
type: test
---
Runs against envtest's real API server with the CRDs from deploy/crds. It installs instance clash-a, then instance clash with role a-token, which under the old hyphen naming both spelled clash-a-token, and checks both run, each mounting its own <instance>--<role> Secret through the reconciler's ownSecrets check. It then creates a Plugin and a PluginGrant named clash--b and a Plugin with role b--token and expects the API server to refuse each as Invalid, so no pair of instance and role can come to name one Secret (D-89). It does not reach the SecretHeld refusal itself, which TestMayMountSecrets covers at the decision level.