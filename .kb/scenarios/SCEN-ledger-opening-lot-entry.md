---
title: An opening lot keeps its original open date and cost
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T05:24:09.401Z'
id: SCEN-ledger-opening-lot-entry
type: scenario
---
Given a broker position reported at go-live on 2026-03-02 as 10 units first bought on 2019-05-17 for 812.34 USD, when it is booked as an opening lot, then the holding gains 10 units at 812.34 USD against equity, recorded on the go-live date with 2019-05-17 as the lot's open date; an open date after go-live is rejected.