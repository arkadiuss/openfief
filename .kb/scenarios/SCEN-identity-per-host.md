---
title: Passkeys and sessions work only on their host
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T19:26:55.664Z'
id: SCEN-identity-per-host
type: scenario
---
Given allowed origins localhost and a dev host, when a passkey created on localhost is used to log in on the dev host, then the login is refused; a session issued on localhost is refused on the dev host; an origin off the allow-list is refused for every call.