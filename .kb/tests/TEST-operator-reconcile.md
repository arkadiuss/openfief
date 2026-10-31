---
title: The operator reconciles Plugins and PluginGrants on a real API server
status: active
sourceFile: internal/plugins/operator/reconciler_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T06:59:49.893Z'
id: TEST-operator-reconcile
type: test
verification_scope: integration
---
Envtest starts a kube-apiserver and etcd (pinned binaries from task tools) with the CRDs from deploy/crds, creates the openfief namespace and synthetic core and egress proxy Services, and calls PluginReconciler.Reconcile directly. Subtests cover a plugin without a grant, a granted plugin and an upgrade that requests a new host before and after approval, revoking a grant, an image the UnimplementedVerifier cannot verify, a config its schema rejects, an invalid grant (also through GrantReconciler), a deleted plugin's allow-list entry, and registering both controllers with a namespace-scoped manager. Each checks the Deployment image, ServiceAccount, NetworkPolicy, allow-list key, conditions and status. A false pass would be a Deployment created for an unverified image; that case asserts there is none.