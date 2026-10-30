---
title: OKI's yearly tax waits for its act and then taxes the average above the limit
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T21:47:49.378Z'
id: SCEN-pl-oki-act-as-data
type: scenario
---
Given no OKI act in force, when the yearly tax is asked, then nothing is computed; given a synthetic act in force from a date with a limit and rate, when daily balances are summed over OKI accounts, then the average of the year's days on or after that date, half up to grosze, is taxed at the rate on the part above the limit.