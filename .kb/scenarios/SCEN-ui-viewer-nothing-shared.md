---
title: A viewer granted nothing sees a plain empty state
status: active
origin:
  kind: agent
  ref: 'S6b task brief: the Access page''s People section and the viewers end-to-end test (D-65, D-183, D-209, D-210), 2026-11-13'
  recorded_at: '2026-11-13T22:00:00.000Z'
id: SCEN-ui-viewer-nothing-shared
type: scenario
---
Given a viewer whose session the core lists no account, the overview, Accounts and Sleeves pages say "Nothing has been shared with you yet", with no chart and no error, and a viewer with an account sees their pages as before; the owner without accounts is not told anything was shared.