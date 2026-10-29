---
title: The owner and read sessions see installed plugins and their trust, and change nothing
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T04:02:33.809Z'
id: SCEN-ui-installed-plugins
type: scenario
---
Given a signed-in session of either scope, when it opens /plugins, or the owner opens the settings page, then the UI lists each installed instance with its capabilities, trust and phase from the read API, offers no control that changes a plugin or grant, and shows a sentence when the list is unavailable; an account's transactions name their connector instance and mark a removed one.