---
title: The built-in UI's sleeve and account pages and its first-run wizard
status: active
sourceFile: internal/views/ui/onboarding_test.go
origin:
  kind: agent
  recorded_at: '2026-10-29T12:45:31.658Z'
id: TEST-ui-onboarding
type: test
---
Handler tests of the UI against fake core services: the owner sees the sleeve forms and a read session only the tables; each sleeve, kind and account form reaches the core with the plugin's token, the owner's session and the posted fields, a broker account picked as instance and ID, and returns only to an allowed page; a refusal shows the core's reason with the form's page; a read session is refused before the core; the accounts page shows the unmapped broker account; an owner with no account lands on the wizard, whose steps show sleeves, accounts and the end, and a read session can't open it.