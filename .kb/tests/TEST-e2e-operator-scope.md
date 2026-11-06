---
title: On k3d and kind, SubjectAccessReviews show the bundle's operator may do exactly its Role's verbs in openfief and nothing elsewhere
status: active
sourceFile: test/e2e/scope_test.go
verification_scope: end_to_end
origin:
  kind: agent
  recorded_at: '2026-11-06T19:43:38.383Z'
id: TEST-e2e-operator-scope
type: test
proof_contract:
  version: kibi.proof-contract.v1
  integration: e2e
  required_proofs:
    - symbol_id: SYM-e2e-scope-test-bundle-operator-scope
      target: default
  success_policy: all_required_first_attempt
---
TestBundleOperatorScope runs in the bundle run and sends one SubjectAccessReview per verb for the operator ServiceAccount, with its groups, on Secrets, pods, exec, tokens, ConfigMaps, Services, workloads, policies, RBAC, leases, events, the OpenFief resources and CloudNativePG clusters in openfief, default, kube-system and cnpg-system, and on cluster-scoped kinds such as namespaces, nodes, CRDs, ClusterRoles and webhooks. The expected answer comes from deploy/operator/role.yaml: allowed only in openfief and only where the Role names the group, resource and verb. It also fails when nothing at all is allowed, which would mean the reviews asked about the wrong identity.