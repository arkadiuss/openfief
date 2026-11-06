---
title: The Inbox settles a suggested transfer's item in place
status: active
origin:
  kind: agent
  recorded_at: '2026-11-06T06:14:56.510Z'
id: SCEN-ui-inbox-transfer-suggested-item
type: scenario
---
Given an open transfer-suggested item for test/w1 and test/d1, when the owner opens the Inbox, then the item shows among the decisions with both accounts, dates and amount, Confirm and Not a transfer, and counts in the navigation; after Confirm the item is no longer open; a read session sees it waiting for the owner.