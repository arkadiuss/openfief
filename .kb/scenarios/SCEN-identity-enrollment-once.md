---
title: An enrollment code creates one passkey on one origin before it expires
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T19:26:53.767Z'
id: SCEN-identity-enrollment-once
type: scenario
---
Given a one-time enrollment code issued for an origin, when the owner creates a passkey with it on that origin before it expires, then the passkey is stored and the code is used up; a second use, another origin, or use after expiry is refused.