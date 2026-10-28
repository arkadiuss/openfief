---
title: A view plugin calls the core's AuthService over Connect
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T18:34:54.847Z'
id: SCEN-identity-auth-contract
type: scenario
---
Given the core serves openfief.identity.v1.AuthService next to the read API, when a view plugin such as the built-in UI calls one of its methods with the generated Go or TypeScript client, then the call reaches the core over Connect with the browser's origin and, where needed, the owner's session token in the request message, and the read API itself stays read-only.