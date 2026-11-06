---
title: A seeded IKE without its earlier payments raises a missing-setting item
status: active
origin:
  kind: agent
  recorded_at: '2026-11-01T20:11:28.160Z'
id: SCEN-attention-earlier-payments-missing
type: scenario
---
Given three IKE accounts, one seeded from opening cash without earlier payments, one seeded with them entered and one with its whole history, when the missing-settings check runs, then only the first is a condition, subject earlier-payments and the account, and the built-in UI sends the owner to the accounts page to enter them.