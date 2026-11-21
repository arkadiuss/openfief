---
title: The price card shows the instrument, its price, day and source, and the period's four closes
status: active
origin:
  kind: agent
  ref: D-226, the author's decision of 2026-11-21 (docs/DESIGN.md)
  recorded_at: '2026-11-21T12:00:00.000Z'
id: SCEN-ask-market-price-card
type: scenario
---
Given a quote, when the market price tool answers, then its card names the instrument and its latest price in the owner's number format, says the day of the close, the source and the ISIN, and lists the first, highest, lowest and latest close with their days; a ticker that names several instruments gives a card that lists them with their ISINs and no price; an instrument name that carries markup reaches the page escaped. Assumes the read API's answer.
