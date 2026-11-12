---
title: The UI's value, performance, allocation and price charts draw exactly what the read API answers for the same days
status: active
tags:
  - acceptance:p10-8
origin:
  kind: agent
  recorded_at: '2026-11-08T07:27:29.529Z'
id: SCEN-e2e-charts-match-read-api
type: scenario
---
Given the data path's record and the owner signed in to the built-in UI in Chromium, when the owner opens the overview's value over time (gross and all three values), its performance and its allocation by asset class, currency, institution and sleeve of the gross, net and liquidation values, the data path accounts' pages and the data path instruments' pages, then the days and values each Chart.js chart holds once drawn, and each allocation bar's value, share and width, equal what the read API answers a script token for the same days and dimensions, worked out as the UI asks: the totals from the performance's first day, the account's values from its performance's start, a year of prices per listing with the primary source first. Runs on both kind and k3d.