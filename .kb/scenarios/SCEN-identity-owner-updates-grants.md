---
title: The owner changes the grants of a token, a plugin or a viewer in one audited call and lists the view plugins with their grants, and no one else may
status: active
origin:
  kind: agent
  ref: 'S5 task brief: the Access page, script tokens and view plugins (D-209, D-210), 2026-11-13'
  recorded_at: '2026-11-13T18:00:00.000Z'
id: SCEN-identity-owner-updates-grants
type: scenario
---
Given the owner, UpdateGrants ends the grants named in remove and makes those named in add, each with an entry, and returns what is held; adding what is held and removing what is not change nothing; an unknown holder, including a viewer that does not exist or was revoked, or a target to add that is not there, or a malformed target, fails the whole call before any change. ListViewPlugins returns the plugins that read with their own identity, never the built-in UI, with their grants in force, and fails as unavailable when the core cannot read them. A demo session, no session and a caller that is no view plugin are refused both calls over Connect.