---
title: While the model portfolio is off, saved or not, the read API answers not found and the overview draws none of it
status: active
tags:
  - acceptance:p10-19
origin:
  kind: agent
  recorded_at: '2026-11-12T16:08:58.890Z'
id: SCEN-e2e-model-portfolio-off-shows-nothing
type: scenario
---
Given the data path's record and the owner signed in to the built-in UI, when the model portfolio has never been saved, and again when the owner has saved two targets with the Show the model portfolio box unchecked, and again after it was turned on and off, then the read API answers not found for GetModelPortfolio and the overview's performance chart draws the time-weighted return alone, with no second line and no model portfolio section or drawdown table (D-52). Runs on both kind and k3d.