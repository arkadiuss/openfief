---
title: Syncs follow the accounts their instance is bound to
status: active
origin:
  kind: agent
  recorded_at: '2026-11-06T06:14:57.802Z'
id: SCEN-views-syncs-for-viewers
type: scenario
---
Given xtb-main bound to acc-1 and nbp-fx bound to no account, when a caller who may view acc-1 reads the sync statuses, attempts and recent activity, then it sees xtb-main's; a caller who may not view acc-1 sees none of xtb-main's, and only the owner and today's other holders of read access see nbp-fx's, never a viewer.