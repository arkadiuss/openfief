---
title: Anyone sees the price checks; only the owner acts on them
status: active
origin:
  kind: agent
  recorded_at: '2026-10-30T05:17:56.860Z'
id: SCEN-ui-price-checks-page
type: scenario
---
Given a flagged check, a matched one before it and a currency mismatch of another instrument at its own 5%, when any session opens the price checks page, then each instrument shows its latest check with both prices, sources, the difference in percent, the status and the threshold, its history behind it, and the overview counts and links the open one; only an owner session gets the acknowledge button and the threshold form, which takes a percent and sends the core a fraction.