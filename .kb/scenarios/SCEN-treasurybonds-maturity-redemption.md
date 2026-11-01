---
title: A matured lot is redeemed once at its redemption value
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T07:44:23.907Z'
id: SCEN-treasurybonds-maturity-redemption
type: scenario
---
Given an open lot of a bond series that matured on or before today, when the treasury-bonds connector syncs, then it books one redemption on the maturity date at the value the Polish pack gives the lot then, with Belka tax on its interest and purchase discount by the account's wrapper, keyed per lot so a second sync changes nothing; missing index readings make it wait as a problem.