---
title: A revaluation raises and clears a gap item
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T08:55:43.784Z'
id: SCEN-valuation-gap-item-per-account
type: scenario
---
Given an account holding an ETF priced in USD whose source's last price is 9 days before today, when the account is revalued, then a market-data-gap item for the account and the price series names the source and the last price's date, and the USD series with a recent rate is cleared; when prices resume, the next revaluation clears the item.