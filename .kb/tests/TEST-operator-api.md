---
title: The Plugin and PluginGrant API types register with a scheme and deep-copy independently
status: active
sourceFile: internal/plugins/operator/v1alpha1/types_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T06:43:42.602Z'
id: TEST-operator-api
type: test
---
TestAddToScheme registers the openfief.io/v1alpha1 types on a fresh runtime.Scheme and checks it recognizes Plugin, PluginList, PluginGrant and PluginGrantList. TestDeepCopyIsIndependent deep-copies a synthetic Plugin with every field set, a PluginGrant and both lists, compares each copy with go-cmp, then edits slices and raw config bytes of the copy and checks the original did not change. A false pass would be generated deepcopy code sharing a slice or the raw config; the edits target exactly those.