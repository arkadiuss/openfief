---
title: A compiled plugin workload maps onto exactly the expected Kubernetes objects
status: active
sourceFile: internal/plugins/operator/objects_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T06:50:13.360Z'
id: TEST-operator-objects
type: test
verification_scope: unit
---
Table tests compile the synthetic sample plugin under grants with and without an egress host and compare every object BuildObjects returns (ServiceAccount, config ConfigMap, Service, NetworkPolicy, Deployment) field by field with go-cmp against hand-written expectations: no API token, only the projected token and a read-only Secret mounted, default-deny NetworkPolicy with core and proxy pod selectors on pod ports, hardened security contexts, and the operator's environment. Refusal cases cover a too-long name and core or proxy selectors that are empty or select plugin pods. DenyAllNetworkPolicy and AllowListConfigMap are compared in full. A false pass would be an extra egress rule; the whole policy is compared.