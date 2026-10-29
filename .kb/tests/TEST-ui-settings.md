---
title: The UI's settings manage passkeys, script tokens and enrollment links for the owner
status: active
sourceFile: internal/views/ui/settings_test.go
origin:
  kind: agent
  recorded_at: '2026-10-28T19:48:32.191Z'
id: TEST-ui-settings
type: test
---
Runs the UI against the fake AuthService with an owner session. Asserts the settings page lists passkeys and script tokens with their actions, a new script token's secret is shown in the creation response only, an out of range lifetime never reaches the core, revoke and remove reach the core with the right IDs, the core's refusal to remove the last passkey is shown with its reason, and an enrollment link for another origin is requested with that target origin and shown once.