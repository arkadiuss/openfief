---
title: The settings ask to confirm with a passkey and go ahead
status: active
origin:
  kind: agent
  recorded_at: '2026-11-03T02:41:14.102Z'
id: SCEN-ui-passkey-check-prompt
type: scenario
---
Given the owner's settings page and a session whose passkey check is stale, when the owner creates a script token, issues an enrollment link, removes a passkey or adds a passkey on this host, then the UI shows Confirm with your passkey, runs the assertion with the browser's authenticator on the button, and posts the same action again, which then goes ahead; in Chromium with a virtual authenticator the token's secret is shown after the check.