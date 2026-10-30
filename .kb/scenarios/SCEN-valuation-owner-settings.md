---
title: Anyone sees the reporting currency and number format; only the owner reads and changes the settings
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T21:48:03.539Z'
id: SCEN-valuation-owner-settings
type: scenario
---
Given no chosen format, when anyone reads the display settings, then they get the reporting currency and the pack's default format; when the owner reads the settings, then they also get the wrappers and the marginal rate; a read session is refused; the owner's changes are checked and recorded.