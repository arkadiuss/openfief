---
title: The operator runs the built-in UI on k3d and kind, and the core proxies its login page, also through Traefik
status: active
sourceFile: test/e2e/ui_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T20:20:54.926Z'
id: TEST-e2e-ui
type: test
---
On kind and k3d the demo's stack is installed from this checkout's images; the operator deploys the demo's ui instance from the UI's image in the suite's registry, signed with the run's key. Waits with a deadline for the Plugin ui to be Ready with the effective browser path /, then from the client pod asks the core's Service for / and /login, which the core proxies to the UI (303 then 200 with the login page), and on k3d the same through Traefik. A pass with a broken proxy would need the core to render the UI's login page itself.