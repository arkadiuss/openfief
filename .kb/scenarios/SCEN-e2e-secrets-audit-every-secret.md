---
title: Every Secret of the openfief namespace is declared, used only by the pods its rule allows, and its values are nowhere else
status: active
tags:
  - acceptance:p10-14
origin:
  kind: agent
  recorded_at: '2026-11-12T19:55:13.204Z'
id: SCEN-e2e-secrets-audit-every-secret
type: scenario
---
Given the demo stack and the data path's five connector instances with their credentials Secrets and CA bundles, when the test lists every Secret of the openfief namespace, then each has a declared rule with every key classified as sensitive or public, the Secret's users among the pods are exactly those its rule allows (the core, the demo-seed Job and CloudNativePG's pods for the database's, each instance's own pods for its Secrets, and never the operator, the UI or the egress proxy) and the pods its rule requires use it; and no sensitive value of eight bytes or more appears in any ConfigMap of the cluster, any pod spec of the namespace, any Plugin or PluginGrant, or, for the connectors' credentials, a dump of the database. An undeclared Secret or key fails the test. Runs on both kind and k3d. It assumes CloudNativePG's Secret names and keys, which the test lists, and that the fixture's own pod, in its namespace, may take the synthetic tokens as arguments, being the institution.