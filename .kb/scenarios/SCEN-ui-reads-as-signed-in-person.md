---
title: Every read a viewer's pages make carries the viewer's session token and shows only what the core lists them
status: active
origin:
  kind: agent
  ref: 'S6 task brief: viewers, people, enrollment links, viewer reads, revoke (D-65, D-209, D-210, D-165, D-199), 2026-11-13'
  recorded_at: '2026-11-13T20:00:00.000Z'
id: SCEN-ui-reads-as-signed-in-person
type: scenario
---
Visiting the overview, accounts, sleeves, plugins and inbox as a viewer makes reads that all carry the viewer's session token and none the plugin's; the accounts page shows the granted account and not another. The read client adds no Authorization of its own and passes a caller's session token untouched.