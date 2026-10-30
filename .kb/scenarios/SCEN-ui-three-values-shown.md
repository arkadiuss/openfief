---
title: The overview and account pages show gross, net and liquidation side by side
status: active
origin:
  kind: agent
  recorded_at: '2026-10-30T02:37:12.596Z'
id: SCEN-ui-three-values-shown
type: scenario
---
Given an IKZE holding with a net value and no liquidation value for want of the marginal rate, and a holding behind a removed instance, when the owner opens the overview and the account page, then the total, each account and each holding show all three values, the missing one saying why, and the holding is marked no longer updated with the instance named.