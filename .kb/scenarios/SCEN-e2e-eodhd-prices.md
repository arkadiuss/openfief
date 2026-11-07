---
title: End-of-day prices come from the EODHD market-data plugin
status: active
tags:
  - acceptance:p10-6
origin:
  kind: agent
  recorded_at: '2026-11-07T04:21:38.927Z'
id: SCEN-e2e-eodhd-prices
type: scenario
---
Given the eodhd instance synced against the fixture, when the read API lists the prices of the three held instruments in the go-live week, then each weekday has the fixture EODHD close from the eodhd instance, marked as the source the listing is valued at, and the valuations use those closes. Runs on both kind and k3d.