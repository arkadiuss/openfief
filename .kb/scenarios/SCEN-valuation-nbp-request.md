---
title: The NBP adapter asks for one bounded range with the caller's client and its own User-Agent
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T05:57:03.905Z'
id: SCEN-valuation-nbp-request
type: scenario
---
Given an HTTP client supplied by the caller, when the adapter is asked for 2 to 7 January 2025, then it sends one GET to api.nbp.pl/api/exchangerates/tables/A/2025-01-02/2025-01-07/?format=json with Accept application/json and User-Agent openfief, because api.nbp.pl refuses Go's default agent; given a range over 93 days or ending before it starts, then it fails with an invalid range error without sending a request; given a status other than 200 or 404, then it fails with an unexpected status error.