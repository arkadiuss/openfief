---
title: The CLI's Kubernetes store creates, lists and deletes instances and renders them for GitOps
status: active
sourceFile: internal/plugins/kube/instances_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T23:13:56.515Z'
id: TEST-plugins-kube-instances
type: test
---
Tests in package kube_test run kube.Instances on controller-runtime's fake client with the operator's scheme. ResourcesYAML of a synthetic install must equal a golden YAML stream with a Plugin and a PluginGrant and no status or creationTimestamp. Create must apply both and refuse a name that has a Plugin, a grant or both; List must merge Plugins and grants by name with their Secrets, signature reason, readiness and pending approval. Delete must remove the grant, the Plugin and the named Secrets, leave other Secrets, and ignore what is already gone. OwnerSigners must read the openfief-trusted-signers ConfigMap, treat a missing ConfigMap or key as none, and refuse an invalid document. A false pass would be a Delete that removes a shared Secret; the shared Secret is checked.