---
title: Each settings change is an audit entry and an unrecorded change is not made
status: active
origin:
  kind: agent
  ref: the MS1 brief, from D-211, 2026-11-13
  recorded_at: '2026-11-14T00:24:15.490Z'
id: SCEN-models-settings-changes-audited
type: scenario
---
Given the settings use cases, when the owner changes the switch, a toggle, a route or a tier then one entry names the owner's session and the setting, a failing log leaves the settings as they were, and a refused or invalid change writes none.

## Context

M1 audited model calls and the maintainer's decisions but not the settings that allow them. A change of the switch, a toggle, a route or a tier's models is now an entry model_settings_changed, like the log's other entries failing closed.

## Source

> Settings: change each audited - check M1's audit coverage for settings changes, add if missing.

Source: the MS1 brief, from D-211, 2026-11-13
