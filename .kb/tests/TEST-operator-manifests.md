---
title: The product's install bases ship no Ingress and the operator no cluster-wide RBAC
status: active
sourceFile: internal/plugins/operator/manifests_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T07:08:28.920Z'
id: TEST-operator-manifests
type: test
verification_scope: unit
---
TestManifests walks deploy/ and lists the kind of every YAML document. The product bases (deploy/crds, deploy/operator, deploy/egress-proxy and deploy/core) must declare no Ingress, IngressClass, Gateway or HTTPRoute (D-28), and deploy/operator no ClusterRole or ClusterRoleBinding (D-15); deploy/operator/role.yaml must hold exactly one Role. deploy/demo is deliberately outside the product list: the demo installation adds its own routes. A false pass would be a walk that finds no manifests, which fails the test.