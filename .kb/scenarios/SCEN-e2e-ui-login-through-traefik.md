---
title: The demo's Traefik serves the UI's login page
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T20:20:59.306Z'
id: SCEN-e2e-ui-login-through-traefik
type: scenario
---
Given the demo's Ingress on k3s's Traefik and the ui-from-traefik policy, when a client in another namespace asks / through Traefik, then the UI redirects it to /login, and /login answers 200 with the login page.