---
title: A holding is the sum of current postings up to the date
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T05:39:01.804Z'
id: SCEN-ledger-holding-derivation
type: scenario
---
Given an account with an opening lot of 3 units, a buy whose row was corrected from 10 to 12 units, a sale of 4 and a dividend, when its holding of the instrument is asked for at the end of each date, then it is 3 on go-live, 15 after the buy, and 11 after the sale and after the dividend; units of the same instrument in another account never count.