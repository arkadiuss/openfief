---
title: Only view plugins call the AuthService
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T19:27:01.526Z'
id: SCEN-identity-view-callers
type: scenario
---
Given a plugin calls the AuthService, when its bearer token is a projected ServiceAccount token with audience openfief-core of a plugin whose effective capabilities include view, then the call is served; any other caller gets permission denied.