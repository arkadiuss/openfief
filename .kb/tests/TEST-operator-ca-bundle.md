---
title: The operator reports the CA bundle from the spec alone and copies the stamp to the pod template
status: active
sourceFile: internal/plugins/operator/ca_bundle_test.go
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-07T10:22:57.996Z'
id: TEST-operator-ca-bundle
type: test
---
Runs the PluginReconciler against envtest, a real API server with the CRDs: a Plugin with extraCa and a grant approving the bundle Secret is reconciled with nothing stamped (CABundleMissing True ChecksumNotStamped, Secret mounted, Deployment applied, no requeue), after a labelled Secret is created by hand (unchanged, since the operator never reads it), after a stamp and a new stamp (False ChecksumStamped with exactly that checksum on the pod template), after clearing it (True again, no annotation) and with extraCa off (no condition, no volume). Each step checks the message never holds the Secret data. TestStampCABundle covers the annotation helper.