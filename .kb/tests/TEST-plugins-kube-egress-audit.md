---
title: The egress audit reads every egress proxy pod's log and every PluginGrant's hosts in the openfief namespace
status: active
sourceFile: internal/plugins/kube/egressaudit_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-07T03:30:40.925Z'
id: TEST-plugins-kube-egress-audit
type: test
---
With client-go's fake clientset, EgressProxyLog reads the log of each egress proxy pod in openfief, ending each with a newline, and of no other pod or namespace. With controller-runtime's fake client, GrantedEgress returns each grant's hosts by instance, port 443 when none is written, ignores grants in other namespaces and fails on a wildcard host.