---
title: The demo's Traefik serves the UI's login page
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T20:20:59.306Z'
id: SCEN-e2e-ui-login-through-traefik
type: scenario
---
Given the demo's Ingress on k3s's Traefik, which routes every path to the core (D-81), and the operator-deployed UI with its browser path /, when a client in another namespace asks / through Traefik, then the core proxies it to the UI, which redirects to /login, and /login answers 200 with the login page. No policy admits Traefik to the UI.