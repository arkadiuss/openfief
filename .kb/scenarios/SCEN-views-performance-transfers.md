---
title: Only a transfer the owner confirmed is left out of the cash flows
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T20:54:39.162Z'
id: SCEN-views-performance-transfers
type: scenario
---
Given 100 USD withdrawn from acc-1 on day 1 and deposited into acc-2 on day 2, when the owner confirmed the pair the money is in transit on day 1 at the day's rate and the index moves only with the rate; when the pair is only suggested or was rejected, the withdrawal and the deposit are cash flows of the time-weighted return. A link whose legs no longer match counts for nothing.