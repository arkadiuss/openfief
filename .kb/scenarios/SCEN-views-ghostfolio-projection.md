---
title: Each transaction becomes Ghostfolio activities commented with its key
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T00:54:24.475Z'
id: SCEN-views-ghostfolio-projection
type: scenario
---
Given a buy of 3 ETF1 for 1261.50 USD with a 1.25 USD fee, when it is projected, then it becomes a BUY of 3 at 420.5 USD with fee 1.25 and the comment openfief:test/a1/b1; given a buy with a fee in EUR, then the fee becomes a FEE of its own commented #fee:EUR; given a dividend of 12.50 USD with 1.88 USD withheld, then a DIVIDEND of quantity 1 at 12.5 with fee 1.88; given a deposit, then no activity; given an instrument with no listing, then a problem and no activity.