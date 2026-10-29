---
title: A sync makes the owned accounts hold exactly the projection, and a second sync changes nothing
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T00:54:28.283Z'
id: SCEN-views-ghostfolio-reconcile
type: scenario
---
Given owned accounts holding a hand-edited activity, a missing one, a stray one and a copy, when the view syncs, then it imports the missing and the desired version of the edited one before deleting the edited, stray and copied ones; given an owned account that should hold nothing, then one bulk delete filtered to it; given an activity moved to another owned account, then it is deleted before the import it would block; when the view syncs again, then the plan is empty.