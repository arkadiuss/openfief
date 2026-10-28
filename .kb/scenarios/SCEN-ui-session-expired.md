---
title: Pages without a valid session go to the login page
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T19:48:59.056Z'
id: SCEN-ui-session-expired
type: scenario
---
Given no session cookie, or a session the core answers UNAUTHENTICATED for, when the browser opens a page of the record, then the UI redirects to /login (an htmx request gets HX-Redirect), deletes a stale cookie, and never calls the read API without a session.