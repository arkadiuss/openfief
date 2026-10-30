---
title: The owner confirms or rejects a suggested transfer on the transfers page
status: active
origin:
  kind: agent
  recorded_at: '2026-10-30T06:32:33.576Z'
id: SCEN-ui-transfers-page
type: scenario
---
Given a suggested withdrawal and deposit between two accounts, when the owner presses Confirm the UI sends the pair to TransferService and the page lists it as confirmed; Not a transfer lists it as rejected, and either decision can be corrected later. A read session only sees the lists; the overview's performance card links the page while suggestions wait.