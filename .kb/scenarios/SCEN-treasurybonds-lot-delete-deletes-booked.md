---
title: Deleting a bond lot deletes its coupons and their cash
status: active
origin:
  kind: agent
  recorded_at: '2026-11-02T10:24:29.880Z'
id: SCEN-treasurybonds-lot-delete-deletes-booked
type: scenario
---
Given a COI lot held in a bond account with two coupons the treasury-bonds connector booked for it, when the owner deletes the lot, then the core sends the connector the account's current transactions that instance delivered, keyed without the core's instance prefix, and the connector delivers a deletion of the lot's own transactions and of each coupon, redemption and payout withdrawal booked under the lot's key, so the account holds no current transaction of the lot and no cash from it, while another lot's coupons and other sources stay (D-154).