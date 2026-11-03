---
title: Sign out everywhere on the settings page
status: active
origin:
  kind: agent
  recorded_at: '2026-11-03T02:41:16.568Z'
id: SCEN-ui-sign-out-everywhere
type: scenario
---
Given the owner's settings page, when the owner presses Sign out everywhere, then the form posts to /logout so the service worker empties its cache, the core revokes every session, the cookie goes and the login page opens; a refusal is shown and keeps the cookie.