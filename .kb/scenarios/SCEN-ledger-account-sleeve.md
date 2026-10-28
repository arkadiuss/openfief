---
title: An account without a sleeve is rejected
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T05:23:55.737Z'
id: SCEN-ledger-account-sleeve
type: scenario
---
Given an account at an institution with an ID and a name, when it names a sleeve, then it is valid; when it names no sleeve, then validation fails with an invalid account error. Assumes the sleeve is referred to by its ID and that the sleeve itself is checked by the portfolio context.