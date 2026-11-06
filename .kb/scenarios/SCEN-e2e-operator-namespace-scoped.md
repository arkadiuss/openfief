---
title: The operator's ServiceAccount may act only in the openfief namespace, and there only as its Role says
status: active
tags:
  - acceptance:p10-12
origin:
  kind: agent
  recorded_at: '2026-11-06T19:44:39.258Z'
id: SCEN-e2e-operator-namespace-scoped
type: scenario
---
Given OpenFief installed from the release bundle, when the API server is asked with SubjectAccessReviews what the operator's ServiceAccount and its groups may do, verb by verb, on Secrets, pods, workloads, policies, RBAC, leases, events and the OpenFief resources in openfief and in other namespaces, and on cluster-scoped kinds, then every answer matches deploy/operator/role.yaml: its verbs on its resources in openfief, and nothing anywhere else, so the operator cannot create or change anything outside openfief nor read a Secret. Runs on both kind and k3d.