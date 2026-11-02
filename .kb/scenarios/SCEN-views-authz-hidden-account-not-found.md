---
title: A hidden account answers as a missing one
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T22:46:10.184Z'
id: SCEN-views-authz-hidden-account-not-found
type: scenario
---
Given a caller the authorizer lets view acc-1 only, when it asks the read API about acc-2's account, transactions, holdings, lots, cash, balance assertions, valuations or valuation status, then each answer is not found with the same body as for an account that is not stored, and acc-1's answers succeed.