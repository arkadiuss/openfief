---
title: The tickers page offers a trading currency on London
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T23:37:59.604Z'
id: SCEN-ui-ticker-currency-choice
type: scenario
---
Given a ticker waiting on an exchange that quotes in more than one currency, when the owner opens the tickers page, then its form offers those currencies, required when the proposal names no currency and defaulting to the listing's when it does; the posted pick reaches the core with the ISIN, the core's refusal naming the choices is shown, and a currency longer than three letters is refused first.