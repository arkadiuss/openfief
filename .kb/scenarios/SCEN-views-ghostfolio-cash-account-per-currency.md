---
title: An account's cash in other currencies gets owned Ghostfolio accounts of its own
status: active
origin:
  kind: agent
  recorded_at: '2026-10-31T19:00:15.091Z'
id: SCEN-views-ghostfolio-cash-account-per-currency
type: scenario
---
Given an account of the record whose cash postings are in the Ghostfolio user's base currency (PLN) and also in EUR and USD, when the view projects the record, then it owns the account's own Ghostfolio account in PLN and one account per other currency, in code order, each named by the account's name and the currency's code and identified by the UUIDv5 of the instance, the account and the currency, which never changes. Each holds only its currency's balance history: one balance at the end of each date that currency's cash changes. An account whose cash is all in another currency still gets its own account. A balance too large for a double leaves out only that owned account's balances, as a problem naming the account and the currency.