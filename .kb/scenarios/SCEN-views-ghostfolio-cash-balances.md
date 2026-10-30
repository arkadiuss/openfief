---
title: Each owned account's cash balance follows the account's cash
status: active
origin:
  kind: agent
  recorded_at: '2026-10-30T01:20:53.948Z'
id: SCEN-views-ghostfolio-cash-balances
type: scenario
---
Given an account whose cash in the Ghostfolio user's base currency moves on several dates, when the view projects it, then the owned account gets one balance at the end of each date the cash changes, none on a date that nets to nothing, a transaction left out still moving the cash; cash in another currency is reported as a problem per account and currency; a balance too large for a double leaves out the account's balances. Against Ghostfolio 3.93.0, the first sync sets three balances and a removed transaction's balance is deleted.