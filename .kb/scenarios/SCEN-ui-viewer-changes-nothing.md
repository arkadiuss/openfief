---
title: Every POST route of the UI refuses a viewer
status: active
origin:
  kind: agent
  ref: 'S6 task brief: viewers, people, enrollment links, viewer reads, revoke (D-65, D-209, D-210, D-165, D-199), 2026-11-13'
  recorded_at: '2026-11-13T20:00:00.000Z'
id: SCEN-ui-viewer-changes-nothing
type: scenario
---
Posting to each route registered for POST with the owner's guard, as a viewer, gives 403 and reaches no service; every other POST route is a sign-in, enrollment or sign-out step; the pages show Viewer access and the viewer's name.