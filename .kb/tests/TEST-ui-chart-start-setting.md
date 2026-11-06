---
title: The UI saves the charts' start as the owner's setting, clears it, and marks an account opened on its seeding day
status: active
sourceFile: internal/views/ui/chartsetting_test.go
verification_scope: unit
origin:
  kind: agent
  recorded_at: '2026-11-02T08:22:11.321Z'
id: TEST-ui-chart-start-setting
type: test
---
Runs the built-in UI over its fake read API and fake SettingsService and AccountService. The owner's date field posts to /chart-start, which saves the portfolio's start and, apart from it, an account's; the page then starts its charts on the saved start without naming it in the URL, Back to the default clears it, and a start link previews without saving. A read session can't save and gets the preview form; a malformed start is refused before the core is asked. The account forms send the opened-on-seeding-day checkbox checked, unchecked, or not at all when the form didn't show it. A false pass would need the fake to store a start the UI never sent.