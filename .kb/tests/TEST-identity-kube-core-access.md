---
title: openfief enroll finds a running, ready core pod to exec into and forward to
status: active
sourceFile: internal/identity/kube/firstpasskey_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T12:01:52.609Z'
id: TEST-identity-kube-core-access
type: test
---
Table test of CoreAccess.CorePod over the client-go fake clientset: among synthetic pods it picks the first in name order that is a core pod, running, ready and not being deleted; with none it fails with ErrNoCorePod.