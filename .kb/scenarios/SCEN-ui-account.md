---
title: An account page shows its holdings, lots, cash, checks and paged transactions
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T19:49:13.267Z'
id: SCEN-ui-account
type: scenario
---
Given an account in the synthetic record, when the owner opens it, then the page shows holdings with their stored values and basis, cash per currency, open lots, balance assertion checks and the first page of transactions with a link to the next, whose rows htmx loads with the page token; an unknown account is 404.