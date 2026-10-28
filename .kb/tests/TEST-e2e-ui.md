---
title: The built-in UI runs on k3d and kind and answers its login page through Traefik
status: active
sourceFile: test/e2e/ui_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T20:20:54.926Z'
id: TEST-e2e-ui
type: test
---
Runs on a local k3d or kind cluster after the demo's stack is installed from this checkout's images, with the UI from the stand-in of the operator's workload in overlays/openfief. Waits with a deadline for Deployment plugin-ui to be available, failing at once on a stuck pod; on k3s, waits for Traefik and checks that / through the demo's Ingress answers 303 and /login answers 200 with the passkey button. A false pass would need another backend answering the login page, which nothing else serves.