---
title: Every PLN value and cash figure of the read API uses the NBP table A rate of its day
status: active
tags:
  - acceptance:p10-6
origin:
  kind: agent
  recorded_at: '2026-11-07T04:21:38.145Z'
id: SCEN-e2e-nbp-rates-every-conversion
type: scenario
---
Given the fixture NBP table A rates of GBP, CHF and NOK, rising each weekday, recorded by the nbp-fx instance, when the read API serves the valuations of the IBKR and XTB accounts on the go-live day and a week later, then each holding PLN value and the day cash and total in PLN equal the amounts times that day NBP rate, rounded to grosze, as computed by hand. Runs on both kind and k3d.