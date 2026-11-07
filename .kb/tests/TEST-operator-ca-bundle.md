---
title: The operator reports a missing CA bundle, stamps its checksum and rolls the pod, and never blocks the plugin
status: active
sourceFile: internal/plugins/operator/ca_bundle_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-07T10:22:57.996Z'
id: TEST-operator-ca-bundle
type: test
---
Runs the PluginReconciler against envtest, a real API server with the CRDs: a Plugin with extraCa and a grant approving the bundle Secret is reconciled while the Secret is missing (CABundleMissing True SecretMissing, Deployment applied, a one minute requeue), after it is created and changed (False SecretPresent with the checksum of exactly that data on the pod template), after deletion (True again, no checksum), with a reader the API refuses (Unknown SecretUnreadable) and with extraCa turned off (no condition, no volume). Each step checks the condition message never holds certificate text. TestStampCABundle covers the annotation helper.