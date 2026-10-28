---
title: The operator publishes the cluster's signing keys and the SDK verifies the core's TokenRequest tokens with them
status: active
sourceFile: internal/plugins/operator/signingkeys_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T16:47:23.886Z'
id: TEST-operator-signing-keys
type: test
---
Envtest suites against a real kube-apiserver with default RBAC. Impersonating the operator's ServiceAccount, IssuerDiscovery reads /.well-known/openid-configuration and /openid/v1/jwks with no OpenFief ClusterRole, while a plain authenticated user is forbidden. The publisher applies openfief-sa-jwks with the API server's issuer and requeues after its refresh. Tokens requested through the TokenRequest API for the core's ServiceAccount with audience openfief-plugin:nbp pass the SDK's JWKSVerifier reading only the ConfigMap's files; another instance's audience, another ServiceAccount and a plugin's core token are refused. A deleted ConfigMap stays absent while the source fails and returns on the next reconcile; run in a manager, the publisher publishes at start and puts a deleted ConfigMap back. A false pass would be keys that verify nothing; the TokenRequest round trip guards it.