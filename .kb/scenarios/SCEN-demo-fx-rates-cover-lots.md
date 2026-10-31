---
title: Every demo holding has its net and liquidation values
status: active
origin:
  kind: agent
  recorded_at: '2026-10-31T12:17:02.038Z'
id: SCEN-demo-fx-rates-cover-lots
type: scenario
---
Given the demo dataset, whose IBKR opening lots were opened up to 420 days before go-live in dollars and euros, when the loader loads it into an empty database from the committed NBP rates, then every holding of every account on today has a gross, a net and a liquidation value, none missing for want of an FX rate, because the rates start two weeks before the earliest lot's open date. Assumes the Polish pack converts a lot's cost at the rate of the business day before its open date.