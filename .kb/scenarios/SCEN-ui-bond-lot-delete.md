---
title: The owner deletes a bond lot from a bond account's open lots
status: active
origin:
  kind: agent
  recorded_at: '2026-10-30T07:37:16.899Z'
id: SCEN-ui-bond-lot-delete
type: scenario
---
Given an owner-scoped session on the page of an account at the treasury bond registry, when the owner presses a lot's Delete button, then the UI hands the lot's series and purchase date to the core's BondLotService and goes back to the open lots, or shows the core's reason; a read session never sees the button and is refused.