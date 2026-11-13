---
title: A viewer reads only what they are granted, through the real core
status: active
origin:
  kind: agent
  ref: 'S6 task brief: viewers, people, enrollment links, viewer reads, revoke (D-65, D-209, D-210, D-165, D-199), 2026-11-13'
  recorded_at: '2026-11-13T20:00:00.000Z'
id: SCEN-identity-viewer-views-granted-accounts
type: scenario
---
An owner invites a viewer with a sleeve grant; GET /v1/accounts lists the accounts in that sleeve and no other, the others are not found, and the policy list equals what each account's check gives. A viewer with no grants views nothing; an error reading grants or accounts refuses.