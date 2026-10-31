---
title: The owner's delete and remap forms reach the core and show its refusals
status: active
origin:
  kind: agent
  recorded_at: '2026-10-31T14:35:16.795Z'
id: SCEN-ui-delete-and-remap
type: scenario
---
Given the owner on the sleeves or accounts page, when they post a delete of a sleeve, kind or account or a remap of a broker account, then the UI hands it to the core with the session and shows the core's refusal when it refuses; a read session gets no such form and may not post one.