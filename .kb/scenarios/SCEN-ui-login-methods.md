---
title: The login page offers the ways in the core allows
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T19:48:45.756Z'
id: SCEN-ui-login-methods
type: scenario
---
Given the core says passkeys exist and demo sessions are or are not available at the browser's origin, when the owner opens /login, then the page shows Sign in with passkey and shows View demo only when demo sessions are available; when the core can't answer, the page still renders and says so.