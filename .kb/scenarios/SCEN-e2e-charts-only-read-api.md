---
title: While the UI draws its charts it calls the core outside the read API only for the session check and lookups without numbers, and reaches nothing but the core
status: active
tags:
  - acceptance:p10-8
origin:
  kind: agent
  recorded_at: '2026-11-08T07:27:30.962Z'
id: SCEN-e2e-charts-only-read-api
type: scenario
---
Given the data path's record, when the built-in UI draws the overview's, the accounts' and the instruments' charts for the signed-in owner, then the core's log over those spans shows the ui plugin calling no service but the read API, the AuthService (the session check), StatementService's ListStatementAccounts, TickerService's ListConfirmedTickers and AskModelService's GetAskModel (one boolean: whether the Ask box sends questions to a model, D-203), none of which answers a value, price, return or allocation, and read API requests logged in each span; and the UI plugin's NetworkPolicy lets it reach only the core's plugin port, with no proxy, no egress token and no allowed host, so the charts' numbers can come from the read API alone. Runs on both kind and k3d.