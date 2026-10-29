---
title: CoreAccess finds the running, ready core pod and reads the origins it allows
status: active
sourceFile: internal/identity/kube/firstpasskey_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T12:01:52.609Z'
id: TEST-identity-kube-core-access
type: test
---
With the client-go fake clientset: CorePod picks the first running, ready, not deleted pod labelled openfief-core by name, ignoring other workloads, and fails with ErrNoCorePod when there is none. AuthOrigins (origins_test.go) reads --auth-origin values written with = or as the next argument, with one dash or two, in order, ignores flags that only start alike or lack a value, and falls back to the core's default http://localhost:8080; AllowedOrigins reads them from the core container of the ready core pod, not from another container, and fails without a core pod.