---
title: A manually edited activity in an owned Ghostfolio account is reverted, and the owner's own account is untouched
status: active
tags:
  - acceptance:p10-9
origin:
  kind: agent
  recorded_at: '2026-11-08T08:59:33.080Z'
id: SCEN-e2e-ghostfolio-edit-reverted
type: scenario
---
Given the owned accounts hold the projection and the owner keeps an activity and a cash balance of their own in the account Ghostfolio made at sign-up, when the owner edits one owned activity by hand through Ghostfolio's API as its UI does, one more unit with the comment kept, and the view syncs, then the sync imports the projected activity and deletes the edited one and nothing else, the owned accounts hold exactly the projection again with only that activity under a new ID and their balances unchanged, and the owner's own account still holds its activity and balance with the same IDs and values. Runs on both kind and k3d.