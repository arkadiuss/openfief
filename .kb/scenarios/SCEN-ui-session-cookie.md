---
title: The session cookie is host-only, HttpOnly, Secure and SameSite=Strict
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T19:48:56.949Z'
id: SCEN-ui-session-cookie
type: scenario
---
Given a session from the core, when the UI stores it, then the cookie is named __Host-openfief_session with Path=/, HttpOnly, Secure and SameSite=Strict, on HTTPS and on http://localhost alike, and the origin the core gets comes from X-Forwarded-Proto and the Host.