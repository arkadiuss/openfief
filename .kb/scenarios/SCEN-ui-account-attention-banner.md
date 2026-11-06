---
title: An account's page shows its own items
status: active
origin:
  kind: agent
  recorded_at: '2026-11-06T04:32:12.308Z'
id: SCEN-ui-account-attention-banner
type: scenario
---
Given open items tied to an account, a broken transfer's on both legs' accounts, when its page opens, then a banner at the top lists them, each linked to its place in the Inbox; an account without items has no banner.