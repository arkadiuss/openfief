---
title: A session expires when idle or old, and on logout
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T19:26:46.249Z'
id: SCEN-identity-session-lifetime
type: scenario
---
Given the owner logged in and got an opaque session token, when they use it every 11 hours, then it stays valid until 7 days after it was issued and no longer; when they leave it unused for 12 hours it expires; when they log out it is revoked at once; only the SHA-256 hash of its key is stored.