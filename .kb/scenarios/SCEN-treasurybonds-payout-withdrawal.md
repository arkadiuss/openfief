---
title: A paid-out coupon or redemption is withdrawn the same day
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T07:44:26.144Z'
id: SCEN-treasurybonds-payout-withdrawal
type: scenario
---
Given an account whose payout setting is paid out since a day, when a coupon or maturity falls due on or after that day, then the connector books it and then a withdrawal of the cash it brought on the same day, keyed after it; what fell due before the day stays as cash.