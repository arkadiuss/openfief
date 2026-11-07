---
title: The install lets the operator get exactly the CA bundle Secret of an instance that expects one
status: active
sourceFile: internal/plugins/kube/cabundle_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-07T10:22:57.147Z'
id: TEST-plugins-kube-ca-bundle-access
type: test
---
TestCABundleAccess checks the Role grants get on the single Secret <instance>--ca-bundle in openfief and the RoleBinding binds the operator ServiceAccount; TestInstancesCABundleAccess drives kube.Instances against a fake API server through install with extraCA, an upgrade without it, an upgrade with it again, uninstall and an install without it, checking the Role and RoleBinding exist exactly when expected, and that the --print YAML holds them only for an instance that expects a bundle.