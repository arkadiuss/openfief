---
title: The demo's stack runs on kind and k3d from this checkout's images and serves the read API
status: active
sourceFile: test/e2e/core_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T18:00:54.167Z'
id: TEST-e2e-demo-stack
type: test
---
Go tests with the e2e build tag under test/e2e, run by task e2e on a local kind or k3d cluster after task e2e:images has loaded images built from the checkout. TestMain installs CloudNativePG from its release manifest at the demo's pinned operator version, deploy/crds, the demo's database deploy/demo/database, the demo's app layer deploy/demo/openfief through an overlay that only swaps the images and admits the suite's client pod, on k3s the demo's Ingress through Traefik, and two curl pods in their own namespace. The tests then check from the client pod, through the core's Service and real NetworkPolicies, that the core becomes available and answers /readyz, also through Traefik on k3s; that a pod no policy admits cannot connect; that the demo-seed Job completes; that a script token issued with /core token issue inside the core pod lists the three demo accounts, each valued through today; and that a call without a token gets 401. Every wait has a deadline and fails fast on a pod that cannot start; a failing test logs pods, events, policies, endpoints, pod specs and logs. A false pass would be a check that reached the core around its policies, so calls go through the Service from a pod the policies must admit, and the outsider pod proves the policies are enforced.