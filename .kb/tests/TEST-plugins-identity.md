---
title: Plugin identity maps to and from the ServiceAccount plugin-<name> in openfief
status: active
sourceFile: internal/plugins/identity_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T05:28:02.776Z'
id: TEST-plugins-identity
type: test
verification_scope: unit
---
Table tests derive ServiceAccount names and map Kubernetes usernames back to plugin names: system:serviceaccount:openfief:plugin-sample maps to sample and round-trips; other namespaces (including the prefix trick openfief-x), non-plugin accounts, empty or invalid plugin names, extra segments, user names and the empty string do not map to any plugin. A false pass would be accepting another namespace; two cases cover it.