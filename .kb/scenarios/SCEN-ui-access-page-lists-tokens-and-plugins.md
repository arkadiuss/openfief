---
title: The Access page lists each script token with its status, expiry, MCP flag and grants in words, and each view plugin with its grants, and the owner changes them from it
status: active
origin:
  kind: agent
  ref: 'S5 task brief: the Access page, script tokens and view plugins (D-209, D-210), 2026-11-13'
  recorded_at: '2026-11-13T18:00:00.000Z'
id: SCEN-ui-access-page-lists-tokens-and-plugins
type: scenario
---
Given the owner's session, the Access page shows each script token with its status, creation and expiry dates, whether it may use MCP and what it may view by sleeve and account name, with Revoke and a grants form on an active one and neither on a revoked one; and each installed view plugin that reads with its own identity with its grants and a grants form; a core that cannot read its cluster leaves the tokens and says so in place of the plugins. A new token's form posts its name, days, MCP flag and grants, shows the secret once and never again, and a lifetime or grant out of range never reaches the core; a refusal shows the core's reason and keeps the form; a token whose grants SpiceDB did not take says nothing was kept. A grants form reaches the core as one UpdateGrants with exactly the boxes changed.