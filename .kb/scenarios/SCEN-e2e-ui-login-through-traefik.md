---
title: The demo's Traefik serves the UI's login page
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T20:20:59.306Z'
id: SCEN-e2e-ui-login-through-traefik
type: scenario
---
Given the demo's stack on k3d with the demo's Ingress, which points at Service openfief-core port http, the core's browser port, when a client asks Traefik for /login on the route's host, then Traefik reaches the core through the core-from-traefik policy, the core proxies the path to the built-in UI, and the login page comes back; nothing routes to the core's plugin port or to the UI directly.