---
title: Ghostfolio sees a redemption as a sell
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T07:44:32.117Z'
id: SCEN-views-ghostfolio-redemption-sell
type: scenario
---
Given a redemption of two units for 210 with 1.90 withheld, when the Ghostfolio view projects the account, then it is a SELL of two at 105 with the withheld tax as its fee.