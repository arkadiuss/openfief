---
title: The owner manages passkeys, script tokens and enrollment links
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T19:49:07.250Z'
id: SCEN-ui-settings-actions
type: scenario
---
Given an owner session, when the owner creates a script token, then its secret is shown in that response only; revoking a token and removing a passkey reach the core; the core's refusal to remove the last passkey is shown with its reason; an enrollment link for another origin is shown once.