---
title: A sale's tax cost includes its lots' buy fees, and only sales are realized
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T20:11:08.709Z'
id: SCEN-valuation-realized-gains-buy-fees
type: scenario
---
Given a sale whose consumed lot carries a 2 USD buy fee, when its gain for tax is counted, then the fee at the rate of the day before the lot was opened is part of its cost, as it is of a held lot's tax cost in its row; and given an account's dividend, it is no sale, so it never enters the offsetting of gains and losses.