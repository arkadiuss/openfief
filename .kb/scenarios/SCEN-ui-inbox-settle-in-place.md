---
title: One field or button settles an item in the Inbox
status: active
origin:
  kind: agent
  recorded_at: '2026-11-06T04:32:02.769Z'
id: SCEN-ui-inbox-settle-in-place
type: scenario
---
Given open items of a ticker to confirm, a flagged price check, a suggested or broken transfer, an unmapped broker account and a sync that keeps failing, when the owner confirms the ISIN, acknowledges the check, confirms or rejects the transfer, maps the broker account or presses Retry now in the Inbox, then the UI hands it to the core and comes back to the Inbox; any other item links its accounts or the page where it is settled.