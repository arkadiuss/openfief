---
title: The owner turns agents and models on in Settings and no one else sees the page
status: active
origin:
  kind: agent
  ref: the MS1 brief, from D-212, 2026-11-13
  recorded_at: '2026-11-14T00:24:08.406Z'
id: SCEN-ui-models-page-owner-only
type: scenario
---
Given the Models page, when the owner opens it then the switch and each feature's toggle show their state, a form changes each through the core with the owner's session, the MCP toggle says it is not built yet, and a read, demo or viewer session gets 403 for the page and its forms and sees no link to it.

## Context

D-212 puts the switch and its toggles in Settings; the core's ModelSettingsService existed with no page. The owner alone changes them, so a viewer or the demo never sees the page or its link. The MCP server's toggle is stored only, and the page says so plainly.

## Source

> Agents and models are off by default, behind one switch in Settings, with a toggle under it for each feature: the Ask box's model (D-203), the maintainer (D-204) and the MCP server (D-213).

Source: the MS1 brief, from D-212, 2026-11-13
