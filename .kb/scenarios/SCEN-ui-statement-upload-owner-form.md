---
title: The owner uploads a statement and sees what it became
status: active
origin:
  kind: agent
  recorded_at: '2026-10-30T05:17:54.759Z'
id: SCEN-ui-statement-upload-owner-form
type: scenario
---
Given an owner session on the page of an account whose connector takes files, when the owner posts a file with the preselected mode, then the UI streams it to the core with no mode so the core chooses, and shows the mode the core chose, the accounts a statement of linked accounts was booked into and the broker accounts not mapped; a read session sees no form and is refused with 403.