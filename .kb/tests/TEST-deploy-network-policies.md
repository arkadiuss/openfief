---
title: The database, egress proxy and operator policies admit only their callers
status: active
sourceFile: scripts/tests/network-policies.test.sh
verification_scope: integration
origin:
  kind: agent
  recorded_at: '2026-11-03T00:10:25.960Z'
id: TEST-deploy-network-policies
type: test
---
The script test builds deploy/core, deploy/egress-proxy, deploy/operator and deploy/demo-data with kustomize and checks the built NetworkPolicies with gojq: the database's admits exactly the core on 5432, its own instances on 5432 and 8000 and CloudNativePG's operator on 8000, and no plugin; the demo loader brings its own admission; the proxy admits plugin pods on 4750 alone and its egress leaves every private, shared, link-local and loopback range out; the operator admits nothing. Any extra rule fails a case.