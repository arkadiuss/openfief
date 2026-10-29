---
title: The owner enters a bond lot on a bond account's page and the UI hands it to the core
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T02:14:29.396Z'
id: SCEN-ui-bond-lot-entry
type: scenario
---
Given an owner-scoped session on the page of an account at the treasury bond registry, when the owner posts a lot's series, quantity, purchase date and series terms, then the UI hands it to the core's BondLotService with the plugin's token and goes back to the account's open lots, or shows the core's reason with the form as posted; a read session never sees the form and is refused.